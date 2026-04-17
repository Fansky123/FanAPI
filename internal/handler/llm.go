package handler

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"fanapi/internal/billing"
	"fanapi/internal/db"
	"fanapi/internal/model"
	"fanapi/internal/protocol"
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

	default: // OpenAI 协议
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
//
// @Summary      OpenAI 兼容对话（Chat Completions）
// @Description  发送 OpenAI 格式对话请求，支持流式（SSE）和非流式；将 model 字段填写渠道的 routing_model，服务端自动路由到真实上游模型。
// @Tags         LLM
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        body  body      object  true  "请求体，参考 OpenAI Chat Completions API；model 填渠道名称（routing_model）"
// @Success      200   {object}  object  "OpenAI 格式响应；stream=true 时为 SSE 流"
// @Failure      400   {object}  object  "参数错误"
// @Failure      402   {object}  object  "余额不足"
// @Failure      503   {object}  object  "无可用渠道"
// @Router       /v1/chat/completions [post]
func LLMProxy(c *gin.Context) { llmProxy(c) }

// ClaudeProxy 处理 POST /v1/messages（Anthropic Claude 原生格式）。
// 客户端发送 Claude 原生格式请求，收到 Claude 原生格式 SSE 响应。
//
// @Summary      Anthropic Claude 原生对话
// @Description  发送 Anthropic Messages API 格式请求，支持流式（SSE）；model 填渠道的 routing_model。
// @Tags         LLM
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        body  body      object  true  "Claude Messages 请求体；model 填渠道名称"
// @Success      200   {object}  object  "Claude 格式响应；stream=true 时为 SSE 流"
// @Failure      400   {object}  object  "参数错误"
// @Failure      402   {object}  object  "余额不足"
// @Router       /v1/messages [post]
func ClaudeProxy(c *gin.Context) { llmProxy(c) }

// GeminiProxy 处理 POST /v1/gemini（Google Gemini 原生格式）。
// 客户端发送 Gemini 原生格式请求，收到 Gemini 原生格式 SSE 响应。
//
// @Summary      Google Gemini 原生对话
// @Description  发送 Gemini generateContent 格式请求；model 填渠道的 routing_model（可省略，由 channel_id 指定）。
// @Tags         LLM
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        channel_id  query     int     false  "渠道 ID（兼容旧版）"
// @Param        body        body      object  true   "Gemini generateContent 请求体"
// @Success      200   {object}  object  "Gemini 格式响应"
// @Failure      400   {object}  object  "参数错误"
// @Failure      402   {object}  object  "余额不足"
// @Router       /v1/gemini [post]
func GeminiProxy(c *gin.Context) { llmProxy(c) }

// llmProxy 是三条 LLM 路由的共同实现。
// 支持：
//   - 多渠道负载均衡（加权随机 + 优先级 + 错误率自动屏蔽）
//   - 稳定密钥：按售价升序尝试，失败自动切换更贵的渠道
//   - 格式互转（OpenAI ↔ Claude / Gemini）
//   - 认证扩展（bearer / query_param / basic / sigv4）
//   - 用户分组定价
//   - 失败自动重试（最多 3 个不同渠道）
func llmProxy(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	apiKeyID, _ := c.Get("api_key_id")
	var apiKeyIDVal int64
	if apiKeyID != nil {
		apiKeyIDVal = apiKeyID.(int64)
	}

	// 获取用户 group（用于分组定价）
	var userGroup string
	if raw, ok := c.Get("user_group"); ok {
		userGroup, _ = raw.(string)
	}

	// 获取密钥类型（稳定密钥使用价格升序路由）
	keyType, _ := c.Get("key_type")
	isStable := keyType == "stable"

	channelIDStr := c.Query("channel_id")

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

	// 渠道选择
	var ch *model.Channel
	var triedIDs []int64
	var stableChannels []model.Channel // 稳定密钥：按价格排好序的渠道列表

	if channelIDStr != "" {
		// 直接指定 channel_id，不走负载均衡
		channelID, parseErr := strconv.ParseInt(channelIDStr, 10, 64)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "channel_id 格式错误"})
			return
		}
		ch, err = service.GetChannel(c.Request.Context(), channelID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
	} else {
		routingModel, _ := reqData["model"].(string)
		if routingModel == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请在请求体 model 字段填写模型名称，或通过 channel_id 参数指定渠道"})
			return
		}
		if isStable {
			// 稳定密钥：获取按价格升序排列的渠道列表
			stableChannels, err = service.SelectChannelStable(c.Request.Context(), routingModel)
			if err != nil {
				// 兜底：按 name 精确查找（兼容旧行为）
				ch, err = service.GetChannelByName(c.Request.Context(), routingModel)
				if err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "渠道不存在: " + routingModel})
					return
				}
			} else {
				ch = &stableChannels[0]
			}
		} else {
			ch, err = service.SelectChannel(c.Request.Context(), routingModel)
			if err != nil {
				// 兜底：按 name 精确查找（兼容旧行为）
				ch, err = service.GetChannelByName(c.Request.Context(), routingModel)
				if err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "渠道不存在: " + routingModel})
					return
				}
			}
		}
	}

	llmProxyWithChannel(c, ch, reqData, userID, apiKeyIDVal, userGroup, triedIDs, stableChannels)
}

// llmProxyWithChannel 执行实际的上游请求，支持失败重试（换渠道）。
// stableChannels 非空时使用稳定模式（价格升序），否则使用正常负载均衡。
func llmProxyWithChannel(c *gin.Context, ch *model.Channel, reqData map[string]interface{},
	userID, apiKeyIDVal int64, userGroup string, triedIDs []int64, stableChannels []model.Channel) {

	const maxRetries = 3

	channelID := ch.ID
	triedIDs = append(triedIDs, channelID)

	// 用渠道配置的真实模型名覆盖用户传入的路由键
	if ch.Model != "" {
		reqData["model"] = ch.Model
	}
	// 在协议转换前保存模型名（Gemini 转换后 body 不含 model 字段，但 URL 替换需要用到）
	resolvedModel, _ := reqData["model"].(string)

	proto := effectiveProtocol(ch)

	// 1. 号池 Sticky Key 分配（在 request_script 之前，以便脚本可用 poolKey 变量）
	entityID := apiKeyIDVal
	if entityID == 0 {
		entityID = userID
	}
	var poolKey *model.PoolKey
	if ch.KeyPoolID > 0 {
		pk, pkErr := service.GetOrAssignPoolKey(c.Request.Context(), ch.KeyPoolID, entityID)
		if pkErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "key pool error: " + pkErr.Error()})
			return
		}
		poolKey = pk
	}
	poolKeyValue := ""
	if poolKey != nil {
		poolKeyValue = poolKey.Value
	}

	// 2. request_script（JS）映射（在协议转换之前，允许微调原始请求）
	mappedReq := reqData
	if ch.RequestScript != "" {
		mapped, scriptErr := script.RunMapRequest(ch.RequestScript, reqData, poolKeyValue)
		if scriptErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "入参映射错误: " + scriptErr.Error()})
			return
		}
		mappedReq = mapped
	}

	// 2. 协议格式转换（OpenAI 入参 → 目标协议格式）
	// 注意：仅当目标协议不是 openai 且没有 request_script 的情况下才自动转换；
	// 如果配置了 request_script，认为脚本已经处理了格式，直接透传。
	if proto != protocolOpenAI && ch.RequestScript == "" {
		converted, convErr := protocol.ConvertRequest(mappedReq, proto)
		if convErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "请求格式转换错误: " + convErr.Error()})
			return
		}
		mappedReq = converted
	}

	// 3. 流式注入 include_usage（OpenAI 协议专用）
	isStream := false
	if sv, ok := mappedReq["stream"].(bool); ok {
		isStream = sv
	}
	if isStream && proto == protocolOpenAI {
		mappedReq["stream"] = true
		if _, hasOpts := mappedReq["stream_options"]; !hasOpts {
			mappedReq["stream_options"] = map[string]interface{}{"include_usage": true}
		} else if opts, ok := mappedReq["stream_options"].(map[string]interface{}); ok {
			opts["include_usage"] = true
		}
	}

	// 4. 计算预扣金额（含用户分组定价）
	inputHold, outputHold, calcErr := billing.CalcForUser(ch, reqData, userGroup)
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
	c.Header("X-Corr-Id", corrID)
	if totalHold > 0 {
		_ = service.WriteTx(c.Request.Context(), userID, channelID, apiKeyIDVal, corrID, "hold", totalHold, upstreamCostHold, model.JSON{
			"input_hold":  inputHold,
			"output_hold": outputHold,
			"user_group":  userGroup,
		})
	}

	// 5. 写入 LLM 请求日志
	modelName := resolvedModel
	// 预先计算实际上游 URL（与 sendLLMRequest 中逻辑保持一致）
	upstreamURL := strings.ReplaceAll(ch.BaseURL, "{model}", modelName)
	upstreamMethod := ch.Method
	if upstreamMethod == "" {
		upstreamMethod = "POST"
	}
	llmLog := &model.LLMLog{
		UserID:          userID,
		ChannelID:       channelID,
		APIKeyID:        apiKeyIDVal,
		CorrID:          corrID,
		Model:           modelName,
		IsStream:        isStream,
		UpstreamURL:     upstreamURL,
		UpstreamMethod:  upstreamMethod,
		UpstreamRequest: model.JSON(mappedReq),
		Status:          "pending",
	}
	_, _ = db.Engine.Insert(llmLog)

	// 6. 号池 Key 已在步骤1分配，直接发送上游请求
	// 7. 发送上游请求
	resp, err := sendLLMRequest(c, ch, mappedReq, poolKey, proto, resolvedModel)
	if err != nil {
		service.RecordChannelError(c.Request.Context(), channelID)
		// 尝试换渠道重试
		if len(triedIDs) < maxRetries {
			if nextCh := selectNextChannel(c, reqData, triedIDs, stableChannels); nextCh != nil {
				// 退回已扣的 hold
				if totalHold > 0 {
					_ = billing.Refund(c.Request.Context(), userID, totalHold)
					_ = service.WriteTx(c.Request.Context(), userID, channelID, apiKeyIDVal, corrID, "refund", totalHold, upstreamCostHold, model.JSON{"reason": "channel_retry"})
					_, _ = db.Engine.Where("corr_id = ?", corrID).Cols("status", "error_msg").
						Update(&model.LLMLog{Status: "error", ErrorMsg: "channel retry"})
				}
				llmProxyWithChannel(c, nextCh, reqData, userID, apiKeyIDVal, userGroup, triedIDs, stableChannels)
				return
			}
		}
		llmRefundAndAbort(c, corrID, userID, totalHold, upstreamCostHold, 0, "上游请求失败: "+err.Error())
		return
	}

	// 429 时轮转 Key 重试一次（同渠道）
	if resp.StatusCode == http.StatusTooManyRequests && ch.KeyPoolID > 0 && poolKey != nil {
		resp.Body.Close()
		newKey, rotErr := service.MarkExhaustedAndRotate(c.Request.Context(), ch.KeyPoolID, poolKey.ID, entityID)
		if rotErr == nil {
			poolKey = newKey
			resp, err = sendLLMRequest(c, ch, mappedReq, poolKey, proto, resolvedModel)
			if err != nil {
				service.RecordChannelError(c.Request.Context(), channelID)
				llmRefundAndAbort(c, corrID, userID, totalHold, upstreamCostHold, 0, "上游请求失败(重试): "+err.Error())
				return
			}
		}
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyErr, _ := io.ReadAll(resp.Body)
		service.RecordChannelError(c.Request.Context(), channelID)
		// 5xx 时尝试换渠道
		if resp.StatusCode >= 500 && len(triedIDs) < maxRetries {
			if nextCh := selectNextChannel(c, reqData, triedIDs, stableChannels); nextCh != nil {
				if totalHold > 0 {
					_ = billing.Refund(c.Request.Context(), userID, totalHold)
					_ = service.WriteTx(c.Request.Context(), userID, channelID, apiKeyIDVal, corrID, "refund", totalHold, upstreamCostHold, model.JSON{"reason": "channel_retry"})
					_, _ = db.Engine.Where("corr_id = ?", corrID).Cols("status", "error_msg").
						Update(&model.LLMLog{Status: "error", ErrorMsg: "channel retry"})
				}
				llmProxyWithChannel(c, nextCh, reqData, userID, apiKeyIDVal, userGroup, triedIDs, stableChannels)
				return
			}
		}
		llmRefundAndAbort(c, corrID, userID, totalHold, upstreamCostHold, resp.StatusCode, fmt.Sprintf("上游返回 %d: %s", resp.StatusCode, string(bodyErr)))
		return
	}

	service.RecordChannelSuccess(c.Request.Context(), channelID)

	// ---- 同步响应 ----
	if !isStream {
		respBytes, _ := io.ReadAll(resp.Body)

		// 协议反转：将上游响应统一转换回 OpenAI 格式
		if proto != protocolOpenAI && ch.ResponseScript == "" {
			converted, convErr := protocol.ConvertSyncResponse(respBytes, proto)
			if convErr == nil {
				respBytes = converted
			}
		}

		// 从响应 JSON 提取 usage（支持多协议）
		var respJSON map[string]interface{}
		var syncUsage map[string]interface{}
		if json.Unmarshal(respBytes, &respJSON) == nil {
			syncUsage = protocol.NormalizeUsage(respJSON, proto)
		}

		c.Data(http.StatusOK, "application/json", respBytes)

		_, _ = db.Engine.Where("corr_id = ?", corrID).Cols("upstream_status", "upstream_response").
			Update(&model.LLMLog{UpstreamStatus: http.StatusOK, UpstreamResponse: model.JSON(respJSON)})

		llmSettle(c, ch, reqData, syncUsage, totalHold, userID, channelID, apiKeyIDVal, corrID, userGroup)
		return
	}

	// ---- 流式 SSE 响应 ----
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("X-Accel-Buffering", "no")

	usage := &usageState{protocol: proto}
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

	_, _ = db.Engine.Where("corr_id = ?", corrID).Cols("upstream_status").
		Update(&model.LLMLog{UpstreamStatus: http.StatusOK})

	llmSettle(c, ch, reqData, usage.normalized(reqData), totalHold, userID, channelID, apiKeyIDVal, corrID, userGroup)
}

// selectNextChannel 为重试选择下一个渠道，排除已尝试过的渠道 ID。
// stableChannels 非空时按列表顺序选取下一个未尝试的渠道（稳定密钥模式）。
func selectNextChannel(c *gin.Context, reqData map[string]interface{}, excludeIDs []int64, stableChannels []model.Channel) *model.Channel {
	excluded := make(map[int64]bool, len(excludeIDs))
	for _, id := range excludeIDs {
		excluded[id] = true
	}

	// 稳定密钥：按价格升序列表顺序选取下一个未尝试的渠道
	if len(stableChannels) > 0 {
		for i := range stableChannels {
			if !excluded[stableChannels[i].ID] {
				ch := stableChannels[i]
				return &ch
			}
		}
		return nil
	}

	// 普通密钥：使用负载均衡选取
	routingModel, _ := reqData["model"].(string)
	if routingModel == "" {
		return nil
	}
	nextCh, err := service.SelectChannel(c.Request.Context(), routingModel, excludeIDs...)
	if err != nil {
		return nil
	}
	return nextCh
}

// llmSettle 执行结算：与预扣金额对比，退还多扣或补扣差额，并写计费流水。
// usageData 为精确或估算的 {prompt_tokens, completion_tokens}；
// 为 nil 时（连接在任何输出前断开）全额退款。
func llmSettle(c *gin.Context, ch *model.Channel, reqData, usageData map[string]interface{},
	totalHold, userID, channelID, apiKeyIDVal int64, corrID string, userGroup string) {
	ctx := c.Request.Context()
	upstreamCostHold, _ := billing.CalcUpstreamCost(ch, reqData)

	// 非 token 计费（image/video/audio/count/custom）：预扣即精确值，上游成功即结算完毕，不依赖 usageData。
	if ch.BillingType != "token" {
		_, _ = db.Engine.Where("corr_id = ?", corrID).Cols("status").
			Update(&model.LLMLog{Status: "ok"})
		return
	}

	if usageData == nil {
		if totalHold > 0 {
			_ = billing.Refund(ctx, userID, totalHold)
			_ = service.WriteTx(ctx, userID, channelID, apiKeyIDVal, corrID, "refund", totalHold, upstreamCostHold, model.JSON{"reason": "no_output"})
		}
		_, _ = db.Engine.Where("corr_id = ?", corrID).Cols("status").
			Update(&model.LLMLog{Status: "refunded"})
		return
	}
	respData := map[string]interface{}{"usage": usageData}
	actualCost, settleErr := billing.CalcActualCostForUser(ch, reqData, respData, userGroup)
	actualUpstreamCost, _ := billing.CalcActualUpstreamCost(ch, reqData, respData)
	if settleErr == nil {
		inputFromResponse, _ := ch.BillingConfig["input_from_response"].(bool)
		if !inputFromResponse {
			// 分离结算：预扣已从 DB/Redis 扣除输入费用，结算仅处理输出部分。
			outputCost := actualCost - totalHold
			outputUpstreamCost := actualUpstreamCost - upstreamCostHold
			if outputCost < 0 {
				outputCost = 0
			}
			if outputUpstreamCost < 0 {
				outputUpstreamCost = 0
			}
			if outputCost > 0 {
				_ = billing.Charge(ctx, userID, outputCost)
			}
			_ = service.WriteTx(ctx, userID, channelID, apiKeyIDVal, corrID, "settle", outputCost, outputUpstreamCost, model.JSON{
				"actual_cost": actualCost,
				"held":        totalHold,
				"usage":       usageData,
			})
		} else {
			// input_from_response=true 或非 token 类型：预扣为估算，结算修正差额。
			// 预扣已从 DB 扣除 totalHold，此处补充差额使总扣款等于实际费用。
			delta := totalHold - actualCost
			if delta > 0 {
				// 实际费用低于预估：退还多扣部分
				_ = billing.Refund(ctx, userID, delta)
				upstreamDelta := upstreamCostHold - actualUpstreamCost
				if upstreamDelta < 0 {
					upstreamDelta = 0
				}
				_ = service.WriteTx(ctx, userID, channelID, apiKeyIDVal, corrID, "refund", delta, upstreamDelta, model.JSON{
					"actual_cost": actualCost,
					"held":        totalHold,
					"usage":       usageData,
				})
			} else if delta < 0 {
				// 实际费用高于预估：补扣差额
				_ = billing.Charge(ctx, userID, -delta)
				upstreamExtra := actualUpstreamCost - upstreamCostHold
				if upstreamExtra < 0 {
					upstreamExtra = 0
				}
				_ = service.WriteTx(ctx, userID, channelID, apiKeyIDVal, corrID, "settle", -delta, upstreamExtra, model.JSON{
					"actual_cost": actualCost,
					"held":        totalHold,
					"usage":       usageData,
				})
			}
		}
	}
	_, _ = db.Engine.Where("corr_id = ?", corrID).Cols("status", "usage").
		Update(&model.LLMLog{Status: "ok", Usage: model.JSON(usageData)})
}

// sendLLMRequest 构建并发送对上游 LLM 的 HTTP 请求。
// proto 决定认证默认方式，ch.AuthType 可覆盖为：
//   - "bearer"     (默认) Authorization: Bearer KEY
//   - "query_param" 将 KEY 作为查询参数附加到 URL
//   - "basic"      HTTP Basic Auth，KEY 格式为 "user:pass"
//   - "sigv4"      AWS Signature V4，KEY 格式为 "ACCESS_KEY:SECRET_KEY"
func sendLLMRequest(c *gin.Context, ch *model.Channel, reqData map[string]interface{}, poolKey *model.PoolKey, proto string, resolvedModel string) (*http.Response, error) {
	body, _ := json.Marshal(reqData)
	timeout := time.Duration(ch.TimeoutMs) * time.Millisecond
	httpClient := &http.Client{Timeout: timeout}

	// 支持 {model} 占位符，将渠道配置的模型名注入 URL
	// 例如：https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent
	targetURL := ch.BaseURL
	if resolvedModel != "" {
		targetURL = strings.ReplaceAll(targetURL, "{model}", resolvedModel)
	}

	upReq, err := http.NewRequestWithContext(c.Request.Context(), ch.Method, targetURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	upReq.Header.Set("Content-Type", "application/json")
	upReq.Header.Set("Accept", "text/event-stream")

	// 复制渠道静态 Headers，同时记录静态 API Key
	poolKeyVal := ""
	if poolKey != nil {
		poolKeyVal = poolKey.Value
	}
	poolKeyUsedInHeaders := false
	apiKey := ""
	for k, v := range ch.Headers {
		if sv, ok := v.(string); ok {
			if strings.Contains(sv, "{{pool_key}}") {
				poolKeyUsedInHeaders = true
			}
			resolved := script.ResolveHeaderValue(sv, poolKeyVal)
			if strings.EqualFold(k, "authorization") && strings.HasPrefix(resolved, "Bearer ") {
				apiKey = strings.TrimPrefix(resolved, "Bearer ")
			}
			upReq.Header.Set(k, resolved)
		}
	}

	// 号池 Key 覆盖（fallback：Header 里没用 {{pool_key}} 占位符时走 Authorization: Bearer）
	if poolKey != nil {
		if !poolKeyUsedInHeaders {
			upReq.Header.Set("Authorization", "Bearer "+poolKey.Value)
		}
		apiKey = poolKey.Value
	}

	// ---------- 认证方式 ----------
	authType := ch.AuthType
	if authType == "" {
		authType = "bearer" // 默认
	}
	// Claude 协议默认使用 x-api-key，可被 auth_type 覆盖
	if proto == protocolClaude && authType == "bearer" {
		authType = "claude"
	}

	switch authType {
	case "bearer":
		if apiKey != "" {
			upReq.Header.Set("Authorization", "Bearer "+apiKey)
		}
	case "claude":
		upReq.Header.Del("Authorization")
		if apiKey != "" {
			upReq.Header.Set("x-api-key", apiKey)
		}
		upReq.Header.Set("anthropic-version", "2023-06-01")
	case "query_param":
		upReq.Header.Del("Authorization")
		if apiKey != "" {
			paramName := ch.AuthParamName
			if paramName == "" {
				paramName = "key"
			}
			q := upReq.URL.Query()
			q.Set(paramName, apiKey)
			upReq.URL.RawQuery = q.Encode()
		}
	case "basic":
		upReq.Header.Del("Authorization")
		if apiKey != "" {
			// KEY 格式："user:pass"（或仅密码，此时 user 为空）
			encoded := base64.StdEncoding.EncodeToString([]byte(apiKey))
			upReq.Header.Set("Authorization", "Basic "+encoded)
		}
	case "sigv4":
		upReq.Header.Del("Authorization")
		if apiKey != "" {
			region := ch.AuthRegion
			if region == "" {
				region = "us-east-1"
			}
			service := ch.AuthService
			if service == "" {
				service = "execute-api"
			}
			if signErr := signSigV4(upReq, apiKey, region, service, body); signErr != nil {
				return nil, fmt.Errorf("sigv4 签名失败: %w", signErr)
			}
		}
	}

	return httpClient.Do(upReq)
}

// signSigV4 为请求添加 AWS Signature Version 4 认证头。
// credentialKey 格式："ACCESS_KEY_ID:SECRET_ACCESS_KEY"。
// 实现了标准 AWS SigV4 流程（仅支持 POST + JSON body）。
func signSigV4(req *http.Request, credentialKey, region, svc string, body []byte) error {
	parts := strings.SplitN(credentialKey, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("sigv4 key 格式应为 ACCESS_KEY_ID:SECRET_ACCESS_KEY")
	}
	accessKeyID := parts[0]
	secretKey := parts[1]

	now := time.Now().UTC()
	datestamp := now.Format("20060102")
	amzDate := now.Format("20060102T150405Z")

	req.Header.Set("x-amz-date", amzDate)

	// 构建规范化请求字符串
	parsedURL, _ := url.Parse(req.URL.String())
	canonicalURI := parsedURL.EscapedPath()
	if canonicalURI == "" {
		canonicalURI = "/"
	}
	canonicalQS := parsedURL.RawQuery

	payloadHash := fmt.Sprintf("%x", sha256.Sum256(body))
	req.Header.Set("x-amz-content-sha256", payloadHash)

	host := req.Host
	if host == "" {
		host = parsedURL.Host
	}
	req.Header.Set("Host", host)

	signedHeaders := "content-type;host;x-amz-content-sha256;x-amz-date"
	canonicalHeaders := fmt.Sprintf("content-type:%s\nhost:%s\nx-amz-content-sha256:%s\nx-amz-date:%s\n",
		req.Header.Get("Content-Type"), host, payloadHash, amzDate)

	canonicalReq := strings.Join([]string{
		req.Method,
		canonicalURI,
		canonicalQS,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")

	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", datestamp, region, svc)
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		fmt.Sprintf("%x", sha256.Sum256([]byte(canonicalReq))),
	}, "\n")

	signingKey := hmacSHA256(
		hmacSHA256(
			hmacSHA256(
				hmacSHA256([]byte("AWS4"+secretKey), datestamp),
				region),
			svc),
		"aws4_request")

	signature := fmt.Sprintf("%x", hmacSHA256(signingKey, stringToSign))

	authHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		accessKeyID, credentialScope, signedHeaders, signature)
	req.Header.Set("Authorization", authHeader)
	return nil
}

func hmacSHA256(key []byte, data string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(data))
	return mac.Sum(nil)
}

// llmRefundAndAbort 退款并终止请求（上游失败时调用）。
// corrID 不为空时同步更新 LLMLog 的错误状态。
func llmRefundAndAbort(c *gin.Context, corrID string, userID, credits, upstreamCost int64, upstreamStatus int, errMsg string) {
	if credits > 0 {
		_ = billing.Refund(c.Request.Context(), userID, credits)
		_ = service.WriteTx(c.Request.Context(), userID, 0, 0, corrID, "refund", credits, upstreamCost, model.JSON{"reason": "upstream_error"})
	}
	if corrID != "" {
		_, _ = db.Engine.Where("corr_id = ?", corrID).Cols("status", "upstream_status", "error_msg").
			Update(&model.LLMLog{Status: "error", UpstreamStatus: upstreamStatus, ErrorMsg: errMsg})
	}
	c.JSON(http.StatusBadGateway, gin.H{"error": errMsg})
}
