package handler

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"fanapi/internal/billing"
	"fanapi/internal/model"
	"fanapi/internal/script"
	"fanapi/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	protocolOpenAI = "openai"
	protocolClaude = "claude"
	protocolGemini = "gemini"
)

func effectiveProtocol(ch *model.Channel) string {
	if ch.Protocol == "" {
		return protocolOpenAI
	}
	return ch.Protocol
}

// usageState 在 SSE 流中收集 token 用量，支持 OpenAI / Claude / Gemini 三种协议。
// 最终通过 normalized() 统一输出 {prompt_tokens, completion_tokens} 格式供计费使用。
type usageState struct {
	protocol      string
	promptTokens  int64
	completTokens int64
	lastEvent     string // Claude 专用：记录上一个 "event:" 行的值
}

func (u *usageState) processLine(line string) {
	switch u.protocol {
	case protocolClaude:
		if strings.HasPrefix(line, "event: ") {
			u.lastEvent = strings.TrimPrefix(line, "event: ")
			return
		}
		if strings.HasPrefix(line, "data: ") {
			payload := strings.TrimPrefix(line, "data: ")
			var chunk map[string]interface{}
			if json.Unmarshal([]byte(payload), &chunk) != nil {
				return
			}
			switch u.lastEvent {
			case "message_start":
				// {"message": {"usage": {"input_tokens": N}}}
				if msg, ok := chunk["message"].(map[string]interface{}); ok {
					if usg, ok := msg["usage"].(map[string]interface{}); ok {
						if n, _ := usg["input_tokens"].(float64); n > 0 {
							u.promptTokens = int64(n)
						}
					}
				}
			case "message_delta":
				// {"usage": {"output_tokens": M}}
				if usg, ok := chunk["usage"].(map[string]interface{}); ok {
					if n, _ := usg["output_tokens"].(float64); n > 0 {
						u.completTokens = int64(n)
					}
				}
			}
		}

	case protocolGemini:
		if strings.HasPrefix(line, "data: ") {
			payload := strings.TrimPrefix(line, "data: ")
			var chunk map[string]interface{}
			if json.Unmarshal([]byte(payload), &chunk) != nil {
				return
			}
			// 每片 Gemini SSE chunk 都可能携带 usageMetadata，以最后一片为准
			if meta, ok := chunk["usageMetadata"].(map[string]interface{}); ok {
				if n, _ := meta["promptTokenCount"].(float64); n > 0 {
					u.promptTokens = int64(n)
				}
				if n, _ := meta["candidatesTokenCount"].(float64); n > 0 {
					u.completTokens = int64(n)
				}
			}
		}

	default: // openai
		if strings.HasPrefix(line, "data: ") {
			payload := strings.TrimPrefix(line, "data: ")
			if payload == "[DONE]" {
				return
			}
			var chunk map[string]interface{}
			if json.Unmarshal([]byte(payload), &chunk) != nil {
				return
			}
			if usg, ok := chunk["usage"].(map[string]interface{}); ok {
				if n, _ := usg["prompt_tokens"].(float64); n > 0 {
					u.promptTokens = int64(n)
				}
				if n, _ := usg["completion_tokens"].(float64); n > 0 {
					u.completTokens = int64(n)
				}
			}
		}
	}
}

// normalized 返回标准化的 usage map（prompt_tokens / completion_tokens）。
// 若未收到任何用量数据则返回 nil。
func (u *usageState) normalized() map[string]interface{} {
	if u.promptTokens == 0 && u.completTokens == 0 {
		return nil
	}
	return map[string]interface{}{
		"prompt_tokens":     u.promptTokens,
		"completion_tokens": u.completTokens,
	}
}

// LLMProxy 处理 POST /v1/chat/completions（OpenAI 标准格式）。
// 客户端发送 OpenAI 格式请求，收到 OpenAI 格式 SSE 响应。
func LLMProxy(c *gin.Context) { llmProxy(c) }

// ClaudeProxy 处理 POST /v1/messages（Anthropic Claude 原生格式）。
// 客户端发送 Claude 原生格式请求，收到 Claude 原生格式 SSE 响应。
func ClaudeProxy(c *gin.Context) { llmProxy(c) }

// GeminiProxy 处理 POST /v1/gemini（Google Gemini 原生格式）。
// 客户端发送 Gemini 原生格式请求，收到 Gemini 原生格式 SSE 响应。
func GeminiProxy(c *gin.Context) { llmProxy(c) }

// llmProxy 是三条 LLM 路由的共同实现。
// 设计原则：请求体原样透传给上游，响应体原样返回给客户端，无任何格式转换。
// Channel.Protocol 仅用于决定上游认证头的写法和从 SSE 流提取 usage 的方式（计费用）。
// 如需格式转换，请在渠道配置中填写 request_script / response_script。
func llmProxy(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	apiKeyID, _ := c.Get("api_key_id")
	var apiKeyIDVal int64
	if apiKeyID != nil {
		apiKeyIDVal = apiKeyID.(int64)
	}

	channelIDStr := c.Query("channel_id")
	if channelIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channel_id 必填"})
		return
	}
	channelID, err := strconv.ParseInt(channelIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channel_id 格式错误"})
		return
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "读取请求体失败"})
		return
	}
	var reqData map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体 JSON 格式错误"})
		return
	}

	ch, err := service.GetChannel(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	protocol := effectiveProtocol(ch)

	// 请求映射：仅当渠道配置了 request_script 时执行脚本转换，否则原样透传。
	mappedReq := reqData
	if ch.RequestScript != "" {
		mapped, scriptErr := script.RunMapRequest(ch.RequestScript, reqData)
		if scriptErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "入参映射错误: " + scriptErr.Error()})
			return
		}
		mappedReq = mapped
	}

	// 计算预扣金额并原子扣除
	inputHold, outputHold, calcErr := billing.Calc(ch, reqData)
	if calcErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "计费计算错误: " + calcErr.Error()})
		return
	}
	totalHold := inputHold + outputHold
	upstreamCostHold, _ := billing.CalcUpstreamCost(ch, reqData)

	if totalHold > 0 {
		if chargeErr := billing.Charge(c.Request.Context(), userID, totalHold); chargeErr != nil {
			c.JSON(http.StatusPaymentRequired, gin.H{"error": chargeErr.Error()})
			return
		}
	}

	corrID := uuid.New().String()
	if totalHold > 0 {
		_ = service.WriteTx(c.Request.Context(), userID, channelID, apiKeyIDVal, corrID, "hold", totalHold, upstreamCostHold, model.JSON{
			"input_hold":  inputHold,
			"output_hold": outputHold,
		})
	}

	// 号池 Sticky Key 分配
	entityID := apiKeyIDVal
	if entityID == 0 {
		entityID = userID
	}
	var poolKey *model.PoolKey
	if ch.KeyPoolID > 0 {
		pk, pkErr := service.GetOrAssignPoolKey(c.Request.Context(), ch.KeyPoolID, entityID)
		if pkErr != nil {
			llmRefundAndAbort(c, userID, totalHold, "key pool error: "+pkErr.Error())
			return
		}
		poolKey = pk
	}

	// 发送上游请求（429 时轮转 Key 重试一次）
	resp, err := sendLLMRequest(c, ch, mappedReq, poolKey, protocol)
	if err != nil {
		llmRefundAndAbort(c, userID, totalHold, "上游请求失败: "+err.Error())
		return
	}

	if resp.StatusCode == http.StatusTooManyRequests && ch.KeyPoolID > 0 && poolKey != nil {
		resp.Body.Close()
		newKey, rotErr := service.MarkExhaustedAndRotate(c.Request.Context(), ch.KeyPoolID, poolKey.ID, entityID)
		if rotErr == nil {
			poolKey = newKey
			resp, err = sendLLMRequest(c, ch, mappedReq, poolKey, protocol)
			if err != nil {
				llmRefundAndAbort(c, userID, totalHold, "上游请求失败(重试): "+err.Error())
				return
			}
		}
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyErr, _ := io.ReadAll(resp.Body)
		llmRefundAndAbort(c, userID, totalHold, fmt.Sprintf("上游返回 %d: %s", resp.StatusCode, string(bodyErr)))
		return
	}

	// SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("X-Accel-Buffering", "no")

	// 流式透传并按协议提取 usage（用于结算）
	usage := &usageState{protocol: protocol}
	scanner := bufio.NewScanner(resp.Body)
	c.Stream(func(w io.Writer) bool {
		if !scanner.Scan() {
			return false
		}
		line := scanner.Text()
		fmt.Fprintf(w, "%s\n", line)
		usage.processLine(line)
		return true
	})

	lastUsageData := usage.normalized()

	// ---- 结算阶段 ----
	if lastUsageData != nil {
		respData := map[string]interface{}{"usage": lastUsageData}
		actualCost, settleErr := billing.CalcActualCost(ch, reqData, respData)
		actualUpstreamCost, _ := billing.CalcActualUpstreamCost(ch, reqData, respData)
		if settleErr == nil {
			delta := totalHold - actualCost
			if delta > 0 {
				_ = billing.Refund(c.Request.Context(), userID, delta)
			} else if delta < 0 {
				_ = billing.Charge(c.Request.Context(), userID, -delta)
			}
			_ = service.WriteTx(c.Request.Context(), userID, channelID, apiKeyIDVal, corrID, "settle", actualCost, actualUpstreamCost, model.JSON{
				"actual_cost": actualCost,
				"held":        totalHold,
				"usage":       lastUsageData,
			})
		}
	} else if totalHold > 0 {
		_ = billing.Refund(c.Request.Context(), userID, totalHold)
	}
}

// sendLLMRequest 构建并发送对上游 LLM 的 HTTP 请求。
// protocol 决定认证头格式：
//   - openai / gemini：标准 "Authorization: Bearer KEY"
//   - claude：改为 "x-api-key: KEY" + "anthropic-version: 2023-06-01"
func sendLLMRequest(c *gin.Context, ch *model.Channel, reqData map[string]interface{}, poolKey *model.PoolKey, protocol string) (*http.Response, error) {
	body, _ := json.Marshal(reqData)
	timeout := time.Duration(ch.TimeoutMs) * time.Millisecond
	httpClient := &http.Client{Timeout: timeout}

	upReq, err := http.NewRequestWithContext(c.Request.Context(), ch.Method, ch.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	upReq.Header.Set("Content-Type", "application/json")
	upReq.Header.Set("Accept", "text/event-stream")

	// 复制渠道静态 Headers，同时记录静态 API Key
	apiKey := ""
	for k, v := range ch.Headers {
		if sv, ok := v.(string); ok {
			if strings.EqualFold(k, "authorization") && strings.HasPrefix(sv, "Bearer ") {
				apiKey = strings.TrimPrefix(sv, "Bearer ")
			}
			upReq.Header.Set(k, sv)
		}
	}

	// 号池 Key 覆盖静态 Key
	if poolKey != nil {
		apiKey = poolKey.Value
		upReq.Header.Set("Authorization", "Bearer "+poolKey.Value)
	}

	// Claude 协议：x-api-key 替换 Authorization: Bearer
	if protocol == protocolClaude {
		upReq.Header.Del("Authorization")
		if apiKey != "" {
			upReq.Header.Set("x-api-key", apiKey)
		}
		upReq.Header.Set("anthropic-version", "2023-06-01")
	}

	return httpClient.Do(upReq)
}

// llmRefundAndAbort 退款并终止请求（上游失败时调用）。
func llmRefundAndAbort(c *gin.Context, userID, credits int64, errMsg string) {
	if credits > 0 {
		_ = billing.Refund(c.Request.Context(), userID, credits)
	}
	c.JSON(http.StatusBadGateway, gin.H{"error": errMsg})
}
