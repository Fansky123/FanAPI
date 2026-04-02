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

// LLMProxy 处理 POST /v1/llm，以 SSE 流式代理 LLM 请求并完成双阶段计费：
//  1. 请求时：预扣输入估算费 + 输出最大 token 费（totalHold）
//  2. SSE 结束后：从响应 usage 字段取实际费用，与 totalHold 做差额退/补。
func LLMProxy(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	apiKeyID, _ := c.Get("api_key_id")
	var apiKeyIDVal int64
	if apiKeyID != nil {
		apiKeyIDVal = apiKeyID.(int64)
	}

	// 从 query 获取渠道 ID
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

	// 读取请求体
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

	// 加载渠道配置（Redis 缓存 + DB 回源）
	ch, err := service.GetChannel(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// yaegi 脚本映射入参（将平台统一格式 → 第三方 API 格式）
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
	if totalHold > 0 {
		if chargeErr := billing.Charge(c.Request.Context(), userID, totalHold); chargeErr != nil {
			c.JSON(http.StatusPaymentRequired, gin.H{"error": chargeErr.Error()})
			return
		}
	}

	// 关联 ID，用于串联同一次请求的 hold + settle 两条流水
	corrID := uuid.New().String()

	if totalHold > 0 {
		_ = service.WriteTx(c.Request.Context(), userID, channelID, apiKeyIDVal, corrID, "hold", totalHold, model.JSON{
			"input_hold":  inputHold,
			"output_hold": outputHold,
		})
	}

	// 构建上游 HTTP 请求
	upstreamBody, _ := json.Marshal(mappedReq)
	timeout := time.Duration(ch.TimeoutMs) * time.Millisecond
	httpClient := &http.Client{Timeout: timeout}

	upReq, err := http.NewRequestWithContext(c.Request.Context(), ch.Method, ch.BaseURL, bytes.NewReader(upstreamBody))
	if err != nil {
		llmRefundAndAbort(c, userID, totalHold, err.Error())
		return
	}
	upReq.Header.Set("Content-Type", "application/json")
	upReq.Header.Set("Accept", "text/event-stream")
	// 注入渠道固定请求头（如 Authorization: Bearer <key>）
	for k, v := range ch.Headers {
		if sv, ok := v.(string); ok {
			upReq.Header.Set(k, sv)
		}
	}

	resp, err := httpClient.Do(upReq)
	if err != nil {
		llmRefundAndAbort(c, userID, totalHold, "上游请求失败: "+err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyErr, _ := io.ReadAll(resp.Body)
		llmRefundAndAbort(c, userID, totalHold, fmt.Sprintf("上游返回 %d: %s", resp.StatusCode, string(bodyErr)))
		return
	}

	// 设置 SSE 响应头，透传流给客户端
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("X-Accel-Buffering", "no")

	// 收集最后一帧包含 usage 信息的数据（OpenAI / chatfire 格式兼容）
	var lastUsageData map[string]interface{}
	scanner := bufio.NewScanner(resp.Body)
	c.Stream(func(w io.Writer) bool {
		if !scanner.Scan() {
			return false
		}
		line := scanner.Text()
		fmt.Fprintf(w, "%s\n", line)

		// 从 data: ... 行中提取 usage 字段
		if strings.HasPrefix(line, "data: ") {
			payload := strings.TrimPrefix(line, "data: ")
			if payload != "[DONE]" {
				var chunk map[string]interface{}
				if json.Unmarshal([]byte(payload), &chunk) == nil {
					if usage, ok := chunk["usage"]; ok {
						if usageMap, ok := usage.(map[string]interface{}); ok {
							lastUsageData = usageMap
						}
					}
				}
			}
		}
		return true
	})

	// ---- 结算阶段 ----
	// CalcActualCost 会根据 input_from_response 标志决定是否从响应 usage 中重新计算输入费用。
	// actualCost 可能大于 totalHold（当实际输入 token 超出估算），此时需补扣差额。
	if lastUsageData != nil {
		respData := map[string]interface{}{"usage": lastUsageData}
		actualCost, settleErr := billing.CalcActualCost(ch, reqData, respData)
		if settleErr == nil {
			delta := totalHold - actualCost
			if delta > 0 {
				// 实际费用低于预扣，退回差额
				_ = billing.Refund(c.Request.Context(), userID, delta)
			} else if delta < 0 {
				// 实际费用超出预扣（输入 token 估算偏低），补扣差额
				// 补扣失败不影响已返回的响应，记录日志即可
				_ = billing.Charge(c.Request.Context(), userID, -delta)
			}
			_ = service.WriteTx(c.Request.Context(), userID, channelID, apiKeyIDVal, corrID, "settle", actualCost, model.JSON{
				"actual_cost": actualCost,
				"held":        totalHold,
				"usage":       lastUsageData,
			})
		}
	} else if totalHold > 0 {
		// 未收到 usage 数据（如上游异常截断），退回全部预扣保护用户余额
		_ = billing.Refund(c.Request.Context(), userID, totalHold)
	}
}

// llmRefundAndAbort 退款并终止请求（上游失败时调用）。
func llmRefundAndAbort(c *gin.Context, userID, credits int64, errMsg string) {
	if credits > 0 {
		_ = billing.Refund(c.Request.Context(), userID, credits)
	}
	c.JSON(http.StatusBadGateway, gin.H{"error": errMsg})
}
