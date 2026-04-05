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
	"fanapi/internal/db"
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
// promptTokens / completTokens 从响应尾部的 usage 字段读取（最精确）。
// outputChars 在流式传输过程中实时累计输出文本字节数，作为用户中断时的兜底估算依据
// （约 4 字节 ≈ 1 token）。
type usageState struct {
	protocol      string
	promptTokens  int64
	completTokens int64
	outputChars   int64  // 实时累计输出字符数（兜底估算）
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
				if msg, ok := chunk["message"].(map[string]interface{}); ok {
					if usg, ok := msg["usage"].(map[string]interface{}); ok {
						if n, _ := usg["input_tokens"].(float64); n > 0 {
							u.promptTokens = int64(n)
						}
					}
				}
			case "message_delta":
				if usg, ok := chunk["usage"].(map[string]interface{}); ok {
					if n, _ := usg["output_tokens"].(float64); n > 0 {
						u.completTokens = int64(n)
					}
				}
			case "content_block_delta":
				// 实时累计输出字符（兜底）
				if delta, ok := chunk["delta"].(map[string]interface{}); ok {
					if text, _ := delta["text"].(string); text != "" {
						u.outputChars += int64(len(text))
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
			if meta, ok := chunk["usageMetadata"].(map[string]interface{}); ok {
				if n, _ := meta["promptTokenCount"].(float64); n > 0 {
					u.promptTokens = int64(n)
				}
				if n, _ := meta["candidatesTokenCount"].(float64); n > 0 {
					u.completTokens = int64(n)
				}
			}
			// 实时累计输出字符（兜底）
			if candidates, ok := chunk["candidates"].([]interface{}); ok && len(candidates) > 0 {
				if cand, ok := candidates[0].(map[string]interface{}); ok {
					if content, ok := cand["content"].(map[string]interface{}); ok {
						if parts, ok := content["parts"].([]interface{}); ok {
							for _, p := range parts {
								if pm, ok := p.(map[string]interface{}); ok {
									if text, _ := pm["text"].(string); text != "" {
										u.outputChars += int64(len(text))
									}
								}
							}
						}
					}
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
			// 实时累计输出字符（用户中断时兜底）
			if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						if content, _ := delta["content"].(string); content != "" {
							u.outputChars += int64(len(content))
						}
					}
				}
			}
		}
	}
}

// normalized 返回标准化的 usage map（prompt_tokens / completion_tokens）供计费使用。
// 优先使用响应尾部精确的 usage 字段；若流被中断（无 usage），则根据实时累计的
// outputChars 估算 completion_tokens，并从请求消息内容估算 prompt_tokens，
// 确保用户中断时仍按实际消耗计费，不会全额退款。
func (u *usageState) normalized(req map[string]interface{}) map[string]interface{} {
	if u.promptTokens > 0 || u.completTokens > 0 {
		// 精确值：来自响应尾部 usage 字段
		return map[string]interface{}{
			"prompt_tokens":     u.promptTokens,
			"completion_tokens": u.completTokens,
		}
	}
	if u.outputChars == 0 {
		// 完全没有数据（连接失败等），不作结算
		return nil
	}
	// 兜底估算：用于用户中断或上游未返回 usage 的场景
	// 4 字节 ≈ 1 token，乘以 1.1 留出余量
	estimatedOutput := int64(float64(u.outputChars)/4.0*1.1) + 1
	estimatedInput := billing.EstimateTokensFromRequest(req)
	return map[string]interface{}{
		"prompt_tokens":     estimatedInput,
		"completion_tokens": estimatedOutput,
		"estimated":         true, // 标记为估算值，便于排查
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

	// 判断客户端是否请求流式输出
	isStream := false
	if sv, ok := mappedReq["stream"].(bool); ok {
		isStream = sv
	}

	// 流式：强制注入 include_usage，确保结尾 chunk 携带精确 token 数
	if isStream && effectiveProtocol(ch) == protocolOpenAI {
		mappedReq["stream"] = true
		if _, hasOpts := mappedReq["stream_options"]; !hasOpts {
			mappedReq["stream_options"] = map[string]interface{}{"include_usage": true}
		} else if opts, ok := mappedReq["stream_options"].(map[string]interface{}); ok {
			opts["include_usage"] = true
		}
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
	// 返回给客户端，用于关联计费流水（与 billing_transactions.corr_id 一致）
	c.Header("X-Corr-Id", corrID)
	if totalHold > 0 {
		_ = service.WriteTx(c.Request.Context(), userID, channelID, apiKeyIDVal, corrID, "hold", totalHold, upstreamCostHold, model.JSON{
			"input_hold":  inputHold,
			"output_hold": outputHold,
		})
	}

	// 写入 LLM 请求日志（pending 状态，请求完成后更新）
	modelName, _ := mappedReq["model"].(string)
	llmLog := &model.LLMLog{
		UserID:          userID,
		ChannelID:       channelID,
		APIKeyID:        apiKeyIDVal,
		CorrID:          corrID,
		Model:           modelName,
		IsStream:        isStream,
		UpstreamRequest: model.JSON(mappedReq),
		Status:          "pending",
	}
	_, _ = db.Engine.Insert(llmLog)

	// 号池 Sticky Key 分配
	entityID := apiKeyIDVal
	if entityID == 0 {
		entityID = userID
	}
	var poolKey *model.PoolKey
	if ch.KeyPoolID > 0 {
		pk, pkErr := service.GetOrAssignPoolKey(c.Request.Context(), ch.KeyPoolID, entityID)
		if pkErr != nil {
			llmRefundAndAbort(c, corrID, userID, totalHold, 0, "key pool error: "+pkErr.Error())
			return
		}
		poolKey = pk
	}

	// 发送上游请求（429 时轮转 Key 重试一次）
	resp, err := sendLLMRequest(c, ch, mappedReq, poolKey, protocol)
	if err != nil {
		llmRefundAndAbort(c, corrID, userID, totalHold, 0, "上游请求失败: "+err.Error())
		return
	}

	if resp.StatusCode == http.StatusTooManyRequests && ch.KeyPoolID > 0 && poolKey != nil {
		resp.Body.Close()
		newKey, rotErr := service.MarkExhaustedAndRotate(c.Request.Context(), ch.KeyPoolID, poolKey.ID, entityID)
		if rotErr == nil {
			poolKey = newKey
			resp, err = sendLLMRequest(c, ch, mappedReq, poolKey, protocol)
			if err != nil {
				llmRefundAndAbort(c, corrID, userID, totalHold, 0, "上游请求失败(重试): "+err.Error())
				return
			}
		}
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyErr, _ := io.ReadAll(resp.Body)
		llmRefundAndAbort(c, corrID, userID, totalHold, resp.StatusCode, fmt.Sprintf("上游返回 %d: %s", resp.StatusCode, string(bodyErr)))
		return
	}

	// ---- 同步响应 ----
	if !isStream {
		respBytes, _ := io.ReadAll(resp.Body)

		// 从响应 JSON 的 usage 字段精确取 token 数
		var respJSON map[string]interface{}
		var syncUsage map[string]interface{}
		if json.Unmarshal(respBytes, &respJSON) == nil {
			if usg, ok := respJSON["usage"].(map[string]interface{}); ok {
				pt, _ := usg["prompt_tokens"].(float64)
				ct, _ := usg["completion_tokens"].(float64)
				syncUsage = map[string]interface{}{
					"prompt_tokens":     int64(pt),
					"completion_tokens": int64(ct),
				}
			}
		}

		// 透传给客户端
		c.Data(http.StatusOK, resp.Header.Get("Content-Type"), respBytes)

		// 同步模式：把完整响应体也存入日志
		_, _ = db.Engine.Where("corr_id = ?", corrID).Cols("upstream_status", "upstream_response").
			Update(&model.LLMLog{UpstreamStatus: http.StatusOK, UpstreamResponse: model.JSON(respJSON)})

		// 结算
		llmSettle(c, ch, reqData, syncUsage, totalHold, userID, channelID, apiKeyIDVal, corrID)
		return
	}

	// ---- 流式 SSE 响应 ----
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

	// 流式：把结尾的 usage chunk 存入日志（没精确 usage 时也留存已收到的分析信息）
	_, _ = db.Engine.Where("corr_id = ?", corrID).Cols("upstream_status").
		Update(&model.LLMLog{UpstreamStatus: http.StatusOK})

	llmSettle(c, ch, reqData, usage.normalized(reqData), totalHold, userID, channelID, apiKeyIDVal, corrID)
}

// llmSettle 执行结算：与预扣金额对比，退还多扣或补扣差额，并写计费流水。
// usageData 为精确或估算的 {prompt_tokens, completion_tokens}；
// 为 nil 时（连接在任何输出前断开）全额退款。
func llmSettle(c *gin.Context, ch *model.Channel, reqData, usageData map[string]interface{},
	totalHold, userID, channelID, apiKeyIDVal int64, corrID string) {
	if usageData == nil {
		if totalHold > 0 {
			_ = billing.Refund(c.Request.Context(), userID, totalHold)
		}
		_, _ = db.Engine.Where("corr_id = ?", corrID).Cols("status").
			Update(&model.LLMLog{Status: "refunded"})
		return
	}
	respData := map[string]interface{}{"usage": usageData}
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
			"usage":       usageData,
		})
	}
	_, _ = db.Engine.Where("corr_id = ?", corrID).Cols("status", "usage").
		Update(&model.LLMLog{Status: "ok", Usage: model.JSON(usageData)})
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
// corrID 不为空时同步更新 LLMLog 的错误状态。
func llmRefundAndAbort(c *gin.Context, corrID string, userID, credits int64, upstreamStatus int, errMsg string) {
	if credits > 0 {
		_ = billing.Refund(c.Request.Context(), userID, credits)
	}
	if corrID != "" {
		_, _ = db.Engine.Where("corr_id = ?", corrID).Cols("status", "upstream_status", "error_msg").
			Update(&model.LLMLog{Status: "error", UpstreamStatus: upstreamStatus, ErrorMsg: errMsg})
	}
	c.JSON(http.StatusBadGateway, gin.H{"error": errMsg})
}
