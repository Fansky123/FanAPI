package handler

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"fanapi/internal/billing"
	"fanapi/internal/db"
	"fanapi/internal/model"
	"fanapi/internal/notify"
	"fanapi/internal/protocol"
	"fanapi/internal/script"
	"fanapi/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	protocolOpenAI    = "openai"
	protocolClaude    = "claude"
	protocolGemini    = "gemini"
	protocolResponses = "responses"
)

// OpenAIModels returns an OpenAI-compatible model list for clients that discover
// available models through GET /v1/models.
func OpenAIModels(c *gin.Context) {
	var channels []model.Channel
	if err := db.Engine.Where("is_active = true AND type = ?", "llm").
		Cols("id", "model", "display_name", "created_at").
		OrderBy("id ASC").
		Find(&channels); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type modelInfo struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		OwnedBy string `json:"owned_by"`
	}

	seen := make(map[string]bool)
	data := make([]modelInfo, 0, len(channels))
	for _, ch := range channels {
		id := ch.Model
		if ch.DisplayName != "" {
			id = ch.DisplayName
		}
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		created := int64(0)
		if !ch.CreatedAt.IsZero() {
			created = ch.CreatedAt.Unix()
		}
		data = append(data, modelInfo{
			ID:      id,
			Object:  "model",
			Created: created,
			OwnedBy: "fanapi",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data":   data,
	})
}

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
// imageCount 统计 delta content 中出现的 markdown 图片数量（![ 语法），用于多模态图片计费。
type usageState struct {
	protocol            string
	promptTokens        int64
	completTokens       int64
	cacheCreationTokens int64  // Claude 写入缓存 token（1.25x）
	cacheReadTokens     int64  // Claude/OpenAI/Gemini 命中缓存 token（折才价）
	outputChars         int64  // 实时累计输出字符数（兜底估算）
	imageCount          int64  // 多模态图片生成：响应中检测到的图片数量
	lastEvent           string // Claude 专用：记录上一个 "event:" 行的值
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
						if n, _ := usg["cache_creation_input_tokens"].(float64); n > 0 {
							u.cacheCreationTokens = int64(n)
						}
						if n, _ := usg["cache_read_input_tokens"].(float64); n > 0 {
							u.cacheReadTokens = int64(n)
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
				// Gemini Context Caching: cachedContentTokenCount
				if n, _ := meta["cachedContentTokenCount"].(float64); n > 0 {
					u.cacheReadTokens = int64(n)
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
				// OpenAI prompt caching: prompt_tokens_details.cached_tokens
				if details, ok := usg["prompt_tokens_details"].(map[string]interface{}); ok {
					if n, _ := details["cached_tokens"].(float64); n > 0 {
						u.cacheReadTokens = int64(n)
					}
				}
			}
			// Responses API (response.completed): usage 嵌套在 chunk["response"]["usage"]
			// 字段名为 input_tokens / output_tokens，缓存命中在 input_tokens_details.cached_tokens
			if resp, ok := chunk["response"].(map[string]interface{}); ok {
				if usg, ok := resp["usage"].(map[string]interface{}); ok {
					if n, _ := usg["input_tokens"].(float64); n > 0 {
						u.promptTokens = int64(n)
					}
					if n, _ := usg["output_tokens"].(float64); n > 0 {
						u.completTokens = int64(n)
					}
					if details, ok := usg["input_tokens_details"].(map[string]interface{}); ok {
						if n, _ := details["cached_tokens"].(float64); n > 0 {
							u.cacheReadTokens = int64(n)
						}
					}
				}
			}
			// 实时累计输出字符（用户中断时兜底）；同时统计 markdown 图片数量（多模态模型）
			if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						if content, _ := delta["content"].(string); content != "" {
							u.outputChars += int64(len(content))
							// 统计 markdown 图片（![...](url) 格式）及 base64 内联图片
							// 每个 ![ 或 data:image/ 出现一次视为一张图片
							u.imageCount += int64(strings.Count(content, "!["))
							u.imageCount += int64(strings.Count(content, "data:image/"))
						}
					}
				}
			}
			// Responses API (response.output_text.delta): delta 直接是字符串
			if chunkType, _ := chunk["type"].(string); strings.HasPrefix(chunkType, "response.output_text") {
				if text, _ := chunk["delta"].(string); text != "" {
					u.outputChars += int64(len(text))
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
		result := map[string]interface{}{
			"prompt_tokens":     u.promptTokens,
			"completion_tokens": u.completTokens,
		}
		if u.cacheCreationTokens > 0 {
			result["cache_creation_tokens"] = u.cacheCreationTokens
		}
		if u.cacheReadTokens > 0 {
			result["cache_read_tokens"] = u.cacheReadTokens
		}
		if u.imageCount > 0 {
			result["image_count"] = u.imageCount
		}
		return result
	}
	if u.outputChars == 0 && u.imageCount == 0 {
		// 完全没有数据（连接失败等），不作结算
		return nil
	}
	if u.imageCount > 0 && u.outputChars == 0 {
		// 仅有图片输出（纯图片生成模型，无 token usage），按图片计费路径结算
		return map[string]interface{}{
			"image_count": u.imageCount,
		}
	}
	// 兜底估算：用于用户中断或上游未返回 usage 的场景
	// 4 字节 ≈ 1 token，乘以 1.1 留出余量
	estimatedOutput := int64(float64(u.outputChars)/4.0*1.1) + 1
	estimatedInput := billing.EstimateTokensFromRequest(req)
	result := map[string]interface{}{
		"prompt_tokens":     estimatedInput,
		"completion_tokens": estimatedOutput,
		"estimated":         true, // 标记为估算值，便于排查
	}
	if u.imageCount > 0 {
		result["image_count"] = u.imageCount
	}
	return result
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
func LLMProxy(c *gin.Context) {
	c.Set("client_proto", protocolOpenAI)
	llmProxy(c)
}

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
func ClaudeProxy(c *gin.Context) {
	c.Set("client_proto", protocolClaude)
	llmProxy(c)
}

// GeminiProxy 处理 POST /v1/gemini（Gemini generateContent 兼容格式，非原生路径）。
// 客户端发送 Gemini generateContent 风格请求，收到 Gemini 风格响应。
// 如需兼容 Google AI SDK 原生 URL 路径，请使用 /v1beta/models/{path}。
//
// @Summary      Gemini generateContent 兼容接口（非原生路径）
// @Description  接收 Gemini generateContent 风格请求体并按统一路由转发；这不是 Google Gemini 的原生 URL 路径。若需兼容 Google AI SDK 原生路径，请使用 /v1beta/models/{path}。
// @Tags         LLM
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        channel_id  query     int     false  "渠道 ID（兼容旧版）"
// @Param        body        body      object  true   "Gemini generateContent 风格请求体"
// @Success      200   {object}  object  "Gemini 风格响应"
// @Failure      400   {object}  object  "参数错误"
// @Failure      402   {object}  object  "余额不足"
// @Router       /v1/gemini [post]
func GeminiProxy(c *gin.Context) {
	c.Set("client_proto", protocolGemini)
	llmProxy(c)
}

// GeminiNativeProxy 处理 POST /v1beta/models/{model}:generateContent 和
// POST /v1beta/models/{model}:streamGenerateContent（Google Gemini SDK 原生路径格式）。
//
// 兼容官方 Google AI SDK，只需将 SDK 的 baseURL 指向本服务即可。
// model 从 URL 路径中自动提取并注入请求体，stream 由 URL action 自动推断，
// 无需客户端修改任何请求体字段。
//
// @Summary      Gemini 原生路径兼容（Google AI SDK）
// @Description  兼容 Google Gemini SDK 原生路径格式：非流式访问 :generateContent，流式访问 :streamGenerateContent?alt=sse。model 从 URL 路径自动提取。
// @Tags         LLM
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        path  path  string  true  "{model}:generateContent 或 {model}:streamGenerateContent"
// @Success      200   {object}  object  "Gemini 格式响应；流式时为 SSE"
// @Failure      400   {object}  object  "参数错误"
// @Failure      402   {object}  object  "余额不足"
// @Router       /v1beta/models/{path} [post]
func GeminiNativeProxy(c *gin.Context) {
	// Gin wildcard 路由 /v1beta/models/*path 捕获的值形如 /gemini-2.5-flash:generateContent
	path := strings.TrimPrefix(c.Param("path"), "/")

	// 解析模型名和 action（格式：{model}:{action}）
	parts := strings.SplitN(path, ":", 2)
	modelName := parts[0]
	action := ""
	if len(parts) == 2 {
		action = parts[1]
	}

	// 流式判断：action 为 streamGenerateContent，或查询参数 alt=sse
	isStream := strings.HasPrefix(action, "streamGenerateContent") || c.Query("alt") == "sse"

	// 读取请求体，注入 model 和 stream 字段后替换
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "读取请求体失败"})
		return
	}
	if len(bodyBytes) == 0 {
		bodyBytes = []byte("{}")
	}
	var reqData map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体 JSON 格式错误"})
		return
	}

	// 注入 model（用于渠道路由）
	if modelName != "" {
		reqData["model"] = modelName
	}
	// 注入 stream（llmProxy 在协议转换前读取此字段）
	if isStream {
		reqData["stream"] = true
	}

	newBody, _ := json.Marshal(reqData)
	c.Request.Body = io.NopCloser(bytes.NewReader(newBody))
	c.Request.ContentLength = int64(len(newBody))

	c.Set("client_proto", protocolGemini)
	llmProxy(c)
}

// ResponsesProxy 处理 POST /v1/responses（OpenAI Responses API，Codex CLI 使用）。
// 客户端发送 Responses API 格式请求，收到 Responses API 格式响应（同步或 SSE）。
//
// 请求格式：
//
//	{
//	  "model": "codex-mini-latest",
//	  "input": "hello" / [{"role":"user","content":"hello"}],
//	  "instructions": "You are helpful.",  // 等价于 system 消息
//	  "stream": true,
//	  "max_output_tokens": 4096
//	}
//
// @Summary      OpenAI Responses API（Codex CLI 兼容）
// @Description  兼容 OpenAI Responses API（POST /v1/responses），Codex CLI 默认使用此接口。将 Responses API 格式请求转换为 Chat Completions 格式转发上游，并将响应转换回 Responses API 格式。
// @Tags         LLM
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        body  body      object  true  "Responses API 请求体；model 填渠道名称（routing_model）"
// @Success      200   {object}  object  "Responses API 格式响应；stream=true 时为 SSE 流"
// @Failure      400   {object}  object  "参数错误"
// @Failure      402   {object}  object  "余额不足"
// @Failure      503   {object}  object  "无可用渠道"
// @Router       /v1/responses [post]
func ResponsesProxy(c *gin.Context) {
	c.Set("client_proto", protocolResponses)
	llmProxy(c)
}

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

	// 余额前置检查：通用余额 <= 0 时直接拒绝，无论模型定价是否为 0
	if bal, balErr := billing.GetBalance(c.Request.Context(), userID); balErr == nil && bal <= 0 {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "余额不足，请充值后继续使用"})
		return
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "读取请求体失败"})
		return
	}
	// 保存原始请求体字节，供 passthrough_body 模式使用（在任何解析或修改之前）
	c.Set("raw_body", bodyBytes)
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
			stableChannels, err = service.SelectChannelStableForUser(c.Request.Context(), routingModel, userGroup)
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

	// 捕获用户传入的路由键（用于专属模型积分扣减），必须在模型名覆盖之前保存
	routingKey, _ := reqData["model"].(string)

	// 用渠道配置的真实模型名覆盖用户传入的路由键
	if ch.Model != "" {
		reqData["model"] = ch.Model
	}
	// 在协议转换前保存模型名（Gemini 转换后 body 不含 model 字段，但 URL 替换需要用到）
	resolvedModel, _ := reqData["model"].(string)

	proto := effectiveProtocol(ch)

	// 获取客户端协议（由 LLMProxy/ClaudeProxy/GeminiProxy 写入 context）
	clientProto := protocolOpenAI
	if cp, ok := c.Get("client_proto"); ok {
		if s, ok := cp.(string); ok && s != "" {
			clientProto = s
		}
	}

	// isStream 必须在协议转换前读取（Gemini 转换后 body 不含 stream 字段）
	isStream, _ := reqData["stream"].(bool)

	// 保存原始客户端格式请求，用于：
	// 1. 计费估算（billing 读取 messages 字段，此字段在 Gemini 转换后不存在）
	// 2. 换渠道重试（下一渠道需要原始格式重新转换，而不是已转换格式）
	origReqData := make(map[string]interface{}, len(reqData))
	for k, v := range reqData {
		origReqData[k] = v
	}

	needsResponsesCompatConvert := clientProto == protocolResponses &&
		proto == protocolResponses &&
		hasTopLevelNonEmptyMessages(reqData)

	// 客户端格式 ≠ 渠道格式时，需要请求格式转换链：
	//   客户端格式 → OpenAI（若客户端本身就是 OpenAI 则跳过） → 渠道格式（若渠道本身是 OpenAI 则跳过）
	// 客户端格式 == 渠道格式时直接透传，不做任何转换（Responses 兼容分支除外）。
	// passthrough_body=true 时跳过所有转换，直接使用原始请求体字节。
	if !ch.PassthroughBody && ch.RequestScript == "" && (clientProto != proto || needsResponsesCompatConvert) {
		working := reqData
		// Step 1: 客户端格式 → OpenAI
		if clientProto != protocolOpenAI {
			norm, normErr := protocol.NormalizeClientRequest(working, clientProto)
			if normErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式转换错误: " + normErr.Error()})
				return
			}
			norm["model"] = resolvedModel
			working = norm
		}
		// Step 2: OpenAI → 渠道格式
		if proto != protocolOpenAI {
			conv, convErr := protocol.ConvertRequest(working, proto)
			if convErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "请求格式转换错误: " + convErr.Error()})
				return
			}
			working = conv
		}
		reqData = working
	}

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

	// 2. request_script（JS）映射（有脚本时跳过自动协议转换，由脚本自行处理）
	// passthrough_body=true 时也跳过脚本，确保请求体不被任何逻辑修改。
	mappedReq := reqData
	if !ch.PassthroughBody && ch.RequestScript != "" {
		mapped, scriptErr := script.RunMapRequest(ch.RequestScript, reqData, poolKeyValue)
		if scriptErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "入参映射错误: " + scriptErr.Error()})
			return
		}
		mappedReq = mapped
	}

	// 3a. Claude 渠道：补全必填字段（max_tokens 为 Claude API 强制要求，无论客户端格式如何）
	// passthrough_body=true 时跳过，避免破坏原始请求体的完整性签名。
	if !ch.PassthroughBody && proto == protocolClaude {
		if _, ok := mappedReq["max_tokens"]; !ok {
			mappedReq["max_tokens"] = 4096
		}
	}

	// 3b. 流式注入 include_usage（OpenAI 协议专用）
	// passthrough_body=true 时跳过，避免修改原始请求体。
	if !ch.PassthroughBody && isStream && proto == protocolOpenAI {
		mappedReq["stream"] = true
		if _, hasOpts := mappedReq["stream_options"]; !hasOpts {
			mappedReq["stream_options"] = map[string]interface{}{"include_usage": true}
		} else if opts, ok := mappedReq["stream_options"].(map[string]interface{}); ok {
			opts["include_usage"] = true
		}
	}

	// 4. 计算预扣金额（含用户分组定价）
	// 使用原始客户端格式请求（origReqData）：Gemini 转换后不含 messages 字段，会导致 token 估算为 0
	inputHold, outputHold, calcErr := billing.CalcForUser(ch, origReqData, userGroup)
	if calcErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "计费计算错误: " + calcErr.Error()})
		return
	}
	totalHold := inputHold + outputHold
	upstreamCostHold, _ := billing.CalcUpstreamCost(ch, origReqData)

	var modelCreditCharged int64
	var generalCreditCharged int64
	if totalHold > 0 {
		// 优先扣减专属模型积分，不足部分再扣通用余额
		if routingKey != "" {
			modelCreditCharged, _ = billing.ChargeModelCredit(c.Request.Context(), userID, routingKey, totalHold)
		}
		generalCreditCharged = totalHold - modelCreditCharged

		if generalCreditCharged > 0 {
			balanceBefore, balanceErr := billing.GetBalance(c.Request.Context(), userID)
			if balanceErr != nil {
				log.Printf("[llm-hold] get balance failed user_id=%d channel_id=%d channel=%q model=%q input_hold=%d output_hold=%d total_hold=%d err=%v",
					userID, channelID, ch.Name, resolvedModel, inputHold, outputHold, totalHold, balanceErr)
			} else {
				log.Printf("[llm-hold] try charge user_id=%d channel_id=%d channel=%q model=%q input_hold=%d output_hold=%d total_hold=%d balance_before=%d model_credit_charged=%d",
					userID, channelID, ch.Name, resolvedModel, inputHold, outputHold, totalHold, balanceBefore, modelCreditCharged)
			}
			if chargeErr := billing.Charge(c.Request.Context(), userID, generalCreditCharged); chargeErr != nil {
				// 退回已扣的模型积分
				if modelCreditCharged > 0 {
					_ = billing.RefundModelCredit(c.Request.Context(), userID, routingKey, modelCreditCharged)
				}
				if balanceErr != nil {
					log.Printf("[llm-hold] charge failed user_id=%d channel_id=%d channel=%q model=%q input_hold=%d output_hold=%d total_hold=%d err=%v",
						userID, channelID, ch.Name, resolvedModel, inputHold, outputHold, totalHold, chargeErr)
				} else {
					log.Printf("[llm-hold] charge failed user_id=%d channel_id=%d channel=%q model=%q input_hold=%d output_hold=%d total_hold=%d balance_before=%d err=%v",
						userID, channelID, ch.Name, resolvedModel, inputHold, outputHold, totalHold, balanceBefore, chargeErr)
				}
				c.JSON(http.StatusPaymentRequired, gin.H{"error": chargeErr.Error()})
				return
			}
		}
		// 记录扣款信息，供后续退款时优先还原
		c.Set("model_credit_routing_key", routingKey)
		c.Set("model_credit_charged", modelCreditCharged)
		c.Set("model_credit_general_charged", generalCreditCharged)
	}

	poolKeyIDVal := int64(0)
	if poolKey != nil {
		poolKeyIDVal = poolKey.ID
	}

	corrID := uuid.New().String()
	c.Header("X-Corr-Id", corrID)
	if totalHold > 0 {
		_ = service.WriteTx(c.Request.Context(), userID, channelID, apiKeyIDVal, poolKeyIDVal, corrID, "hold", totalHold, upstreamCostHold, modelCreditCharged, model.JSON{
			"input_hold":  inputHold,
			"output_hold": outputHold,
			"user_group":  userGroup,
		})
	}

	// 5. 写入 LLM 请求日志
	modelName := resolvedModel
	// 预先计算实际上游 URL（与 sendLLMRequest 中逻辑保持一致）
	upstreamURL := strings.ReplaceAll(ch.BaseURL, "{model}", modelName)
	if strings.Contains(upstreamURL, "{stream_action}") {
		if isStream {
			upstreamURL = strings.ReplaceAll(upstreamURL, "{stream_action}", "streamGenerateContent")
		} else {
			upstreamURL = strings.ReplaceAll(upstreamURL, "{stream_action}", "generateContent")
		}
	}
	upstreamMethod := ch.Method
	if upstreamMethod == "" {
		upstreamMethod = "POST"
	}
	llmLog := &model.LLMLog{
		UserID:    userID,
		ChannelID: channelID,
		APIKeyID:  apiKeyIDVal,
		CorrID:    corrID,
		Model:     modelName,
		IsStream:  isStream,
		Transport: func() string {
			if isStream {
				return "sse"
			}
			return "http"
		}(),
		UpstreamURL:     upstreamURL,
		UpstreamMethod:  upstreamMethod,
		UpstreamRequest: model.JSON(mappedReq),
		ClientRequest:   model.JSON(origReqData), // 用户原始请求（协议转换前）
		Status:          "pending",
	}
	_, _ = db.Engine.Insert(llmLog)

	// 6. 号池 Key 已在步骤1分配，直接发送上游请求
	// 7. 发送上游请求
	sentHeaders, resp, err := sendLLMRequest(c, ch, mappedReq, poolKey, proto, resolvedModel, isStream)
	if sentHeaders != nil {
		// 异步写入请求头（不阻塞主流程）
		logID := llmLog.ID
		go func() {
			db.Engine.Where("id = ?", logID).Cols("upstream_headers").
				Update(&model.LLMLog{UpstreamHeaders: model.JSON(func() map[string]interface{} {
					m := make(map[string]interface{}, len(sentHeaders))
					for k, v := range sentHeaders {
						m[k] = v
					}
					return m
				}())})
		}()
	}
	if err != nil {
		service.RecordChannelError(c.Request.Context(), channelID)
		// 尝试换渠道重试
		if len(triedIDs) < maxRetries {
			if nextCh := selectNextChannel(reqData, triedIDs, stableChannels); nextCh != nil {
				// 退回已扣的 hold
				if totalHold > 0 {
					mcRefunded := llmRefundCredits(c, userID, totalHold)
					_ = service.WriteTx(c.Request.Context(), userID, channelID, apiKeyIDVal, poolKeyIDVal, corrID, "refund", totalHold, upstreamCostHold, mcRefunded, model.JSON{"reason": "channel_retry"})
					_, _ = db.Engine.Where("corr_id = ?", corrID).Cols("status", "error_msg").
						Update(&model.LLMLog{Status: "error", ErrorMsg: "channel retry"})
				}
				llmProxyWithChannel(c, nextCh, origReqData, userID, apiKeyIDVal, userGroup, triedIDs, stableChannels)
				return
			}
		}
		llmRefundAndAbort(c, corrID, userID, totalHold, upstreamCostHold, poolKeyIDVal, 0, "上游请求失败: "+err.Error())
		return
	}

	// 429 时轮转 Key 重试一次（同渠道）
	if resp.StatusCode == http.StatusTooManyRequests && ch.KeyPoolID > 0 && poolKey != nil {
		resp.Body.Close()
		newKey, rotErr := service.MarkExhaustedAndRotate(c.Request.Context(), ch.KeyPoolID, poolKey.ID, entityID)
		if rotErr == nil {
			poolKey = newKey
			poolKeyIDVal = newKey.ID // 更新 poolKeyIDVal，确保后续结算流水关联正确的号商
			_, resp, err = sendLLMRequest(c, ch, mappedReq, poolKey, proto, resolvedModel, isStream)
			if err != nil {
				service.RecordChannelError(c.Request.Context(), channelID)
				llmRefundAndAbort(c, corrID, userID, totalHold, upstreamCostHold, poolKeyIDVal, 0, "上游请求失败(重试): "+err.Error())
				return
			}
		}
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyErr, _ := io.ReadAll(resp.Body)
		service.RecordChannelError(c.Request.Context(), channelID)

		// 跑 error_script 识别业务错误（含 fatal=余额不足等永久故障）。
		// 4xx body 通常是 JSON，5xx 也常带结构化错误；非 JSON 时跳过。
		var bizErr string
		var fatal bool
		var bodyJSON map[string]interface{}
		if json.Unmarshal(bodyErr, &bodyJSON) == nil && bodyJSON != nil {
			if ch.ErrorScript != "" {
				bizErr, fatal, _ = script.RunCheckError(ch.ErrorScript, bodyJSON)
			}
			if bizErr == "" {
				if detected, ok := script.DetectUpstreamError(bodyJSON); ok {
					bizErr = detected
				}
			}
		}
		if fatal {
			_ = service.PatchChannelActive(c.Request.Context(), channelID, false)
			log.Printf("[llm] disable channel id=%d fatal_err=%q status=%d", channelID, bizErr, resp.StatusCode)
			go func(name string, id int64, reason string) {
				// 防止通知失败阻塞主流程
				defer func() { recover() }()
				if err := notify.SendLarkChannelDisabled(name, id, reason); err != nil {
					log.Printf("[lark notify] failed: %v", err)
				}
			}(ch.Name, ch.ID, bizErr)
		}

		// 5xx 总是尝试换渠道；4xx 仅在 error_script 命中时换（fatal 或普通业务错误均触发）。
		shouldRetry := resp.StatusCode >= 500 || bizErr != ""
		if shouldRetry && len(triedIDs) < maxRetries {
			if nextCh := selectNextChannel(reqData, triedIDs, stableChannels); nextCh != nil {
				if totalHold > 0 {
					mcRefunded := llmRefundCredits(c, userID, totalHold)
					_ = service.WriteTx(c.Request.Context(), userID, channelID, apiKeyIDVal, poolKeyIDVal, corrID, "refund", totalHold, upstreamCostHold, mcRefunded, model.JSON{"reason": "channel_retry"})
					retryMsg := "channel retry"
					if bizErr != "" {
						retryMsg = "channel retry: " + bizErr
					}
					_, _ = db.Engine.Where("corr_id = ?", corrID).Cols("status", "error_msg").
						Update(&model.LLMLog{Status: "error", ErrorMsg: retryMsg})
				}
				llmProxyWithChannel(c, nextCh, origReqData, userID, apiKeyIDVal, userGroup, triedIDs, stableChannels)
				return
			}
		}

		abortMsg := bizErr
		if abortMsg == "" {
			abortMsg = fmt.Sprintf("上游返回 %d: %s", resp.StatusCode, string(bodyErr))
		}
		llmRefundAndAbort(c, corrID, userID, totalHold, upstreamCostHold, poolKeyIDVal, resp.StatusCode, abortMsg)
		return
	}

	service.RecordChannelSuccess(c.Request.Context(), channelID)

	// ---- 同步响应 ----
	if !isStream {
		respBytes, _ := io.ReadAll(resp.Body)

		// 先解析原始上游响应（格式与 proto 一致），用于 usage 提取、error_script 检测和日志记录。
		// 必须在协议转换之前解析，否则 Claude/Gemini 的 usage 字段名会被改写为 OpenAI 格式，
		// 导致 NormalizeUsage 找不到对应字段而返回 nil，触发错误的全额退款。
		var origRespJSON map[string]interface{}
		_ = json.Unmarshal(respBytes, &origRespJSON)

		if origRespJSON != nil {
			// 200 但 body 内含 error：优先跑 error_script，未命中时走通用 OpenAI error 检测。
			bizErr := ""
			fatal := false
			if ch.ErrorScript != "" {
				if detected, fatalDetected, scriptErr := script.RunCheckError(ch.ErrorScript, origRespJSON); scriptErr == nil {
					bizErr = detected
					fatal = fatalDetected
				}
			}
			if bizErr == "" {
				if detected, ok := script.DetectUpstreamError(origRespJSON); ok {
					bizErr = detected
				}
			}
			if bizErr != "" {
				service.RecordChannelError(c.Request.Context(), channelID)
				if fatal {
					_ = service.PatchChannelActive(c.Request.Context(), channelID, false)
					log.Printf("[llm] disable channel id=%d fatal_err=%q status=200", channelID, bizErr)
				}
				if len(triedIDs) < maxRetries {
					if nextCh := selectNextChannel(reqData, triedIDs, stableChannels); nextCh != nil {
						if totalHold > 0 {
							mcRefunded := llmRefundCredits(c, userID, totalHold)
							_ = service.WriteTx(c.Request.Context(), userID, channelID, apiKeyIDVal, poolKeyIDVal, corrID, "refund", totalHold, upstreamCostHold, mcRefunded, model.JSON{"reason": "channel_retry"})
							_, _ = db.Engine.Where("corr_id = ?", corrID).Cols("status", "error_msg").
								Update(&model.LLMLog{Status: "error", ErrorMsg: "channel retry: " + bizErr})
						}
						llmProxyWithChannel(c, nextCh, origReqData, userID, apiKeyIDVal, userGroup, triedIDs, stableChannels)
						return
					}
				}
				llmRefundAndAbort(c, corrID, userID, totalHold, upstreamCostHold, poolKeyIDVal, http.StatusOK, bizErr)
				return
			}
		}

		// 从原始上游响应提取 usage（使用渠道原始格式，在任何协议转换之前）
		syncUsage := protocol.NormalizeUsage(origRespJSON, proto)

		// 响应格式转换链：渠道格式 → OpenAI → 客户端格式
		// 客户端格式 == 渠道格式时直接透传；有 response_script 时跳过自动转换。
		if clientProto != proto && ch.ResponseScript == "" {
			// Step 1: 渠道格式 → OpenAI
			if proto != protocolOpenAI {
				if converted, convErr := protocol.ConvertSyncResponse(respBytes, proto); convErr == nil {
					respBytes = converted
				}
			}
			// Step 2: OpenAI → 客户端格式
			if clientProto != protocolOpenAI {
				if converted, convErr := protocol.ConvertResponseToClient(respBytes, clientProto); convErr == nil {
					respBytes = converted
				}
			}
		}

		c.Data(http.StatusOK, "application/json", respBytes)

		var clientRespJSON map[string]interface{}
		_ = json.Unmarshal(respBytes, &clientRespJSON)
		_, _ = db.Engine.Where("corr_id = ?", corrID).Cols("upstream_status", "upstream_response", "client_response").
			Update(&model.LLMLog{
				UpstreamStatus:   http.StatusOK,
				UpstreamResponse: model.JSON(origRespJSON),
				ClientResponse:   model.JSON(clientRespJSON),
			})

		llmSettle(c, ch, origReqData, syncUsage, totalHold, userID, channelID, apiKeyIDVal, poolKeyIDVal, corrID, userGroup)
		return
	}

	// ---- 流式 SSE 响应 ----
	// error_script 业务错误检测：许多上游在额度耗尽时即使请求了流式也会立即以 JSON 返回错误。
	// 在写出任何响应字节前 peek 第一行，若匹配 error_script 则换渠道重试。
	if ch.ErrorScript != "" {
		peekBuf := bufio.NewReader(resp.Body)
		firstLineBytes, peekErr := peekBuf.ReadBytes('\n')
		if (peekErr == nil || len(firstLineBytes) > 0) && len(firstLineBytes) > 0 {
			firstLine := strings.TrimRight(string(firstLineBytes), "\r\n")
			checkSrc := strings.TrimPrefix(firstLine, "data: ")
			if checkSrc != "" && checkSrc != "[DONE]" {
				var firstJSON map[string]interface{}
				if json.Unmarshal([]byte(checkSrc), &firstJSON) == nil {
					bizErr := ""
					fatal := false
					if detected, fatalDetected, scriptErr := script.RunCheckError(ch.ErrorScript, firstJSON); scriptErr == nil {
						bizErr = detected
						fatal = fatalDetected
					}
					if bizErr == "" {
						if detected, ok := script.DetectUpstreamError(firstJSON); ok {
							bizErr = detected
						}
					}
					if bizErr != "" {
						service.RecordChannelError(c.Request.Context(), channelID)
						if fatal {
							_ = service.PatchChannelActive(c.Request.Context(), channelID, false)
							log.Printf("[llm] disable channel id=%d fatal_err=%q status=200(stream)", channelID, bizErr)
						}
						if len(triedIDs) < maxRetries {
							if nextCh := selectNextChannel(reqData, triedIDs, stableChannels); nextCh != nil {
								if totalHold > 0 {
									mcRefunded := llmRefundCredits(c, userID, totalHold)
									_ = service.WriteTx(c.Request.Context(), userID, channelID, apiKeyIDVal, poolKeyIDVal, corrID, "refund", totalHold, upstreamCostHold, mcRefunded, model.JSON{"reason": "channel_retry"})
									_, _ = db.Engine.Where("corr_id = ?", corrID).Cols("status", "error_msg").
										Update(&model.LLMLog{Status: "error", ErrorMsg: "channel retry: " + bizErr})
								}
								llmProxyWithChannel(c, nextCh, origReqData, userID, apiKeyIDVal, userGroup, triedIDs, stableChannels)
								return
							}
						}
						llmRefundAndAbort(c, corrID, userID, totalHold, upstreamCostHold, poolKeyIDVal, http.StatusOK, bizErr)
						return
					}
				}
			}
			// 第一行正常：将其拼回，后续 scanner 照常读取
			resp.Body = io.NopCloser(io.MultiReader(strings.NewReader(string(firstLineBytes)), peekBuf))
		}
	} else {
		peekBuf := bufio.NewReader(resp.Body)
		firstLineBytes, peekErr := peekBuf.ReadBytes('\n')
		if (peekErr == nil || len(firstLineBytes) > 0) && len(firstLineBytes) > 0 {
			firstLine := strings.TrimRight(string(firstLineBytes), "\r\n")
			checkSrc := strings.TrimPrefix(firstLine, "data: ")
			if checkSrc != "" && checkSrc != "[DONE]" {
				var firstJSON map[string]interface{}
				if json.Unmarshal([]byte(checkSrc), &firstJSON) == nil {
					if bizErr, ok := script.DetectUpstreamError(firstJSON); ok {
						service.RecordChannelError(c.Request.Context(), channelID)
						if len(triedIDs) < maxRetries {
							if nextCh := selectNextChannel(reqData, triedIDs, stableChannels); nextCh != nil {
								if totalHold > 0 {
									mcRefunded := llmRefundCredits(c, userID, totalHold)
									_ = service.WriteTx(c.Request.Context(), userID, channelID, apiKeyIDVal, poolKeyIDVal, corrID, "refund", totalHold, upstreamCostHold, mcRefunded, model.JSON{"reason": "channel_retry"})
									_, _ = db.Engine.Where("corr_id = ?", corrID).Cols("status", "error_msg").
										Update(&model.LLMLog{Status: "error", ErrorMsg: "channel retry: " + bizErr})
								}
								llmProxyWithChannel(c, nextCh, origReqData, userID, apiKeyIDVal, userGroup, triedIDs, stableChannels)
								return
							}
						}
						llmRefundAndAbort(c, corrID, userID, totalHold, upstreamCostHold, poolKeyIDVal, http.StatusOK, bizErr)
						return
					}
				}
			}
			resp.Body = io.NopCloser(io.MultiReader(strings.NewReader(string(firstLineBytes)), peekBuf))
		}
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("X-Accel-Buffering", "no")

	usage := &usageState{protocol: proto}
	// SSE 格式转换器：客户端格式 != 渠道格式时需要转换 SSE 流
	var sseConv protocol.SSEConverter
	if clientProto != proto {
		sseConv = protocol.NewSSEConverter(proto, clientProto)
	}

	// 收集上游原始 SSE 行用于日志，超过 200KB 后停止收集但继续推流。
	const maxSSELogBytes = 200 * 1024
	var rawSSELines []string
	var rawSSEBytes int

	scanner := bufio.NewScanner(resp.Body)
	c.Stream(func(w io.Writer) bool {
		if !scanner.Scan() {
			if sseConv != nil {
				for _, l := range sseConv.Flush() {
					fmt.Fprintf(w, "%s\n", l)
				}
			}
			return false
		}
		line := scanner.Text()
		usage.processLine(line) // 始终用上游格式解析，保证计费准确
		if rawSSEBytes < maxSSELogBytes {
			rawSSELines = append(rawSSELines, line)
			rawSSEBytes += len(line) + 1
		}
		if sseConv != nil {
			for _, l := range sseConv.Convert(line) {
				fmt.Fprintf(w, "%s\n", l)
			}
		} else {
			fmt.Fprintf(w, "%s\n", line)
		}
		return true
	})

	_, _ = db.Engine.Where("corr_id = ?", corrID).Cols("upstream_status", "upstream_response", "client_response").
		Update(&model.LLMLog{
			UpstreamStatus:   http.StatusOK,
			UpstreamResponse: model.JSON{"lines": rawSSELines},
			ClientResponse:   buildStreamClientResponse(rawSSELines, proto),
		})

	llmSettle(c, ch, origReqData, usage.normalized(origReqData), totalHold, userID, channelID, apiKeyIDVal, poolKeyIDVal, corrID, userGroup)
}

func hasTopLevelNonEmptyMessages(req map[string]interface{}) bool {
	msgs, ok := req["messages"].([]interface{})
	return ok && len(msgs) > 0
}

// selectNextChannel 为重试选择下一个渠道，排除已尝试过的渠道 ID。
// 仅稳定密钥（stableChannels 非空）支持兜底重试，按价格升序列表顺序选取下一个未尝试的渠道。
// 低价密钥不做跨渠道重试，直接返回 nil。
func selectNextChannel(_ map[string]any, excludeIDs []int64, stableChannels []model.Channel) *model.Channel {
	if len(stableChannels) == 0 {
		return nil
	}

	excluded := make(map[int64]bool, len(excludeIDs))
	for _, id := range excludeIDs {
		excluded[id] = true
	}

	for i := range stableChannels {
		if !excluded[stableChannels[i].ID] {
			ch := stableChannels[i]
			return &ch
		}
	}
	return nil
}

// llmSettle 执行结算：与预扣金额对比，退还多扣或补扣差额，并写计费流水。
// usageData 为精确或估算的 {prompt_tokens, completion_tokens}；
// 为 nil 时（连接在任何输出前断开）全额退款。
func llmSettle(c *gin.Context, ch *model.Channel, reqData, usageData map[string]interface{},
	totalHold, userID, channelID, apiKeyIDVal, poolKeyIDVal int64, corrID string, userGroup string) {
	ctx := c.Request.Context()
	upstreamCostHold, _ := billing.CalcUpstreamCost(ch, reqData)

	// 非 token 计费（image/video/audio/count/custom）：预扣即精确值，上游成功即结算完毕，不依赖 usageData。
	// 例外：billing_type=image 且响应中检测到实际图片数量（image_count），按实际图片数调差。
	if ch.BillingType != "token" {
		if ch.BillingType == "image" && usageData != nil {
			var imgCount int64
			switch v := usageData["image_count"].(type) {
			case int64:
				imgCount = v
			case float64:
				imgCount = int64(v)
			}
			if imgCount > 0 {
				// 预扣时使用的图片数量（来自请求 n 字段，默认 1）
				preCount := int64(1)
				switch v := reqData["n"].(type) {
				case float64:
					if v > 0 {
						preCount = int64(v)
					}
				case int64:
					if v > 0 {
						preCount = v
					}
				}
				if imgCount != preCount {
					// 计算单张图片的价格：将 reqData 中 n 改为 1 后调用 CalcForUser
					singleReq := make(map[string]interface{}, len(reqData)+1)
					for k, v := range reqData {
						singleReq[k] = v
					}
					singleReq["n"] = float64(1)
					costPerImage, _, _ := billing.CalcForUser(ch, singleReq, userGroup)
					delta := imgCount - preCount
					if delta > 0 {
						mcCharged := llmChargeExtra(c, userID, costPerImage*delta)
						_ = service.WriteTx(ctx, userID, channelID, apiKeyIDVal, poolKeyIDVal, corrID, "settle",
							costPerImage*delta, 0, mcCharged, model.JSON{
								"reason":      "image_count_adjust",
								"image_count": imgCount,
								"pre_count":   preCount,
							})
					} else {
						refundAmt := costPerImage * (-delta)
						mcRefunded := llmRefundCredits(c, userID, refundAmt)
						_ = service.WriteTx(ctx, userID, channelID, apiKeyIDVal, poolKeyIDVal, corrID, "refund",
							refundAmt, 0, mcRefunded, model.JSON{
								"reason":      "image_count_adjust",
								"image_count": imgCount,
								"pre_count":   preCount,
							})
					}
				}
			}
		}
		_, _ = db.Engine.Where("corr_id = ?", corrID).Cols("status", "usage", "error_msg").
			Update(&model.LLMLog{Status: "ok", Usage: model.JSON(usageData), ErrorMsg: ""})
		return
	}

	if usageData == nil {
		if totalHold > 0 {
			mcRefunded := llmRefundCredits(c, userID, totalHold)
			_ = service.WriteTx(ctx, userID, channelID, apiKeyIDVal, poolKeyIDVal, corrID, "refund", totalHold, upstreamCostHold, mcRefunded, model.JSON{"reason": "no_output"})
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
			// 分离结算：预扣已扣除估算输入费用，此处结算差额（输出 + 缓存折扣调整）。
			// delta = actualCost - totalHold
			//   > 0：实际费用超出预扣（有输出/补扣），需再扣差额
			//   < 0：实际费用低于预扣（高缓存命中率导致输入成本降低），需退还差额
			//   = 0：刚好持平，无需操作
			outputCost := actualCost - totalHold
			outputUpstreamCost := actualUpstreamCost - upstreamCostHold
			if outputCost < 0 {
				// 实际费用低于预扣：退还多扣部分（常见于 Prompt Cache 命中率较高的场景）
				refundAmt := -outputCost
				mcRefunded := llmRefundCredits(c, userID, refundAmt)
				upstreamRefund := int64(0)
				if outputUpstreamCost < 0 {
					upstreamRefund = -outputUpstreamCost
				}
				_ = service.WriteTx(ctx, userID, channelID, apiKeyIDVal, poolKeyIDVal, corrID, "refund", refundAmt, upstreamRefund, mcRefunded, model.JSON{
					"actual_cost": actualCost,
					"held":        totalHold,
					"usage":       usageData,
					"reason":      "cache_discount",
				})
			} else {
				mcCharged := int64(0)
				if outputCost > 0 {
					mcCharged = llmChargeExtra(c, userID, outputCost)
				}
				upstreamSettle := outputUpstreamCost
				if upstreamSettle < 0 {
					upstreamSettle = 0
				}
				_ = service.WriteTx(ctx, userID, channelID, apiKeyIDVal, poolKeyIDVal, corrID, "settle", outputCost, upstreamSettle, mcCharged, model.JSON{
					"actual_cost": actualCost,
					"held":        totalHold,
					"usage":       usageData,
				})
			}
		} else {
			// input_from_response=true 或非 token 类型：预扣为估算，结算修正差额。
			// 预扣已从 DB 扣除 totalHold，此处补充差额使总扣款等于实际费用。
			delta := totalHold - actualCost
			if delta > 0 {
				// 实际费用低于预估：退还多扣部分
				mcRefunded := llmRefundCredits(c, userID, delta)
				upstreamDelta := upstreamCostHold - actualUpstreamCost
				if upstreamDelta < 0 {
					upstreamDelta = 0
				}
				_ = service.WriteTx(ctx, userID, channelID, apiKeyIDVal, poolKeyIDVal, corrID, "refund", delta, upstreamDelta, mcRefunded, model.JSON{
					"actual_cost": actualCost,
					"held":        totalHold,
					"usage":       usageData,
				})
			} else if delta < 0 {
				// 实际费用高于预估：补扣差额
				mcCharged := llmChargeExtra(c, userID, -delta)
				upstreamExtra := actualUpstreamCost - upstreamCostHold
				if upstreamExtra < 0 {
					upstreamExtra = 0
				}
				_ = service.WriteTx(ctx, userID, channelID, apiKeyIDVal, poolKeyIDVal, corrID, "settle", -delta, upstreamExtra, mcCharged, model.JSON{
					"actual_cost": actualCost,
					"held":        totalHold,
					"usage":       usageData,
				})
			}
		}
	}
	_, _ = db.Engine.Where("corr_id = ?", corrID).Cols("status", "usage", "error_msg").
		Update(&model.LLMLog{Status: "ok", Usage: model.JSON(usageData), ErrorMsg: ""})
}

// buildStreamClientResponse 从上游 SSE 原始行中提取并组装文本内容，
// 存入 client_response 供用户端日志展示平台返回了什么。
func buildStreamClientResponse(lines []string, proto string) model.JSON {
	var buf strings.Builder
	var lastEvent string
	for _, line := range lines {
		switch proto {
		case protocolClaude:
			if strings.HasPrefix(line, "event: ") {
				lastEvent = strings.TrimPrefix(line, "event: ")
				continue
			}
			if lastEvent == "content_block_delta" && strings.HasPrefix(line, "data: ") {
				var chunk map[string]interface{}
				if json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &chunk) == nil {
					if delta, ok := chunk["delta"].(map[string]interface{}); ok {
						if text, _ := delta["text"].(string); text != "" {
							buf.WriteString(text)
						}
					}
				}
			}
		case protocolGemini:
			if strings.HasPrefix(line, "data: ") {
				var chunk map[string]interface{}
				if json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &chunk) == nil {
					if candidates, ok := chunk["candidates"].([]interface{}); ok && len(candidates) > 0 {
						if cand, ok := candidates[0].(map[string]interface{}); ok {
							if content, ok := cand["content"].(map[string]interface{}); ok {
								if parts, ok := content["parts"].([]interface{}); ok {
									for _, p := range parts {
										if pm, ok := p.(map[string]interface{}); ok {
											if t, _ := pm["text"].(string); t != "" {
												buf.WriteString(t)
											}
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
					continue
				}
				var chunk map[string]interface{}
				if json.Unmarshal([]byte(payload), &chunk) == nil {
					if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
						if choice, ok := choices[0].(map[string]interface{}); ok {
							if delta, ok := choice["delta"].(map[string]interface{}); ok {
								if text, _ := delta["content"].(string); text != "" {
									buf.WriteString(text)
								}
							}
						}
					}
				}
			}
		}
	}
	text := buf.String()
	if text == "" {
		return nil
	}
	return model.JSON{"content": text, "stream": true}
}

// sendLLMRequest 构建并发送对上游 LLM 的 HTTP 请求。
// proto 决定认证默认方式，ch.AuthType 可覆盖为：
//   - "bearer"     (默认) Authorization: Bearer KEY
//   - "query_param" 将 KEY 作为查询参数附加到 URL
//   - "basic"      HTTP Basic Auth，KEY 格式为 "user:pass"
//   - "sigv4"      AWS Signature V4，KEY 格式为 "ACCESS_KEY:SECRET_KEY"
func sendLLMRequest(c *gin.Context, ch *model.Channel, reqData map[string]interface{}, poolKey *model.PoolKey, _ string, resolvedModel string, isStream bool) (map[string]string, *http.Response, error) {
	// passthrough_body=true：直接使用客户端原始请求体，不做任何序列化/修改
	var body []byte
	if ch.PassthroughBody {
		if rb, ok := c.Get("raw_body"); ok {
			if rawBytes, ok := rb.([]byte); ok {
				body = rawBytes
			}
		}
	}
	if len(body) == 0 {
		body, _ = json.Marshal(reqData)
	}
	timeout := time.Duration(ch.TimeoutMs) * time.Millisecond
	httpClient := &http.Client{Timeout: timeout}

	// 支持 {model} 占位符，将渠道配置的模型名注入 URL
	// 例如：https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent
	// 支持 {stream_action} 占位符，根据请求是否流式选择 Gemini 端点：
	//   流式  → streamGenerateContent?alt=sse
	//   非流式 → generateContent
	targetURL := ch.BaseURL
	if resolvedModel != "" {
		targetURL = strings.ReplaceAll(targetURL, "{model}", resolvedModel)
	}
	if strings.Contains(targetURL, "{stream_action}") {
		if isStream {
			targetURL = strings.ReplaceAll(targetURL, "{stream_action}", "streamGenerateContent")
			if strings.Contains(targetURL, "?") {
				targetURL += "&alt=sse"
			} else {
				targetURL += "?alt=sse"
			}
		} else {
			targetURL = strings.ReplaceAll(targetURL, "{stream_action}", "generateContent")
		}
	}

	method := ch.Method
	if method == "" {
		method = http.MethodPost
	}
	upReq, err := http.NewRequestWithContext(c.Request.Context(), method, targetURL, bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}
	upReq.Header.Set("Content-Type", "application/json")
	upReq.Header.Set("Accept", "text/event-stream")

	// passthrough_headers=true：将客户端请求头原样转发给上游，
	// 保留 User-Agent、Anthropic-Version、Anthropic-Beta 等身份标识头。
	// 渠道 Headers（如 Authorization）在之后写入，可覆盖客户端头。
	if ch.PassthroughHeaders {
		// 跳过这些头：Authorization（由渠道 Headers 覆盖）、逐跳传输头、路由元数据头
		passthroughSkip := map[string]bool{
			"Authorization":     true,
			"Host":              true,
			"Content-Length":    true,
			"Transfer-Encoding": true,
			"Connection":        true,
			"Upgrade":           true,
			"Proxy-Connection":  true,
		}
		for k, vals := range c.Request.Header {
			if !passthroughSkip[k] {
				upReq.Header[k] = vals
			}
		}
	}

	// 将渠道 Headers 里的占位符替换后写入请求
	// 支持 {{pool_key}} / {{}} 注入号池 Key，以及其他动态占位符
	poolKeyVal := ""
	if poolKey != nil {
		poolKeyVal = poolKey.Value
	}
	for k, v := range ch.Headers {
		if sv, ok := v.(string); ok {
			upReq.Header.Set(k, script.ResolveHeaderValue(sv, poolKeyVal))
		}
	}

	// URL 里也支持 {{pool_key}} / {{}} 占位符（如 Gemini ?key={{}} 写法）
	if strings.Contains(upReq.URL.RawQuery, "%7B%7B") || strings.Contains(targetURL, "{{") {
		newURL := script.ResolveHeaderValue(upReq.URL.String(), poolKeyVal)
		if u, err2 := url.Parse(newURL); err2 == nil {
			upReq.URL = u
		}
	}

	// 采集完整请求头（用于管理端日志排查，含完整 API Key）
	if err := applyChannelAuth(upReq, ch, poolKeyVal, body); err != nil {
		return nil, nil, err
	}

	sanitizedHeaders := make(map[string]string, len(upReq.Header))
	for k, vals := range upReq.Header {
		sanitizedHeaders[k] = strings.Join(vals, ", ")
	}

	resp, err := httpClient.Do(upReq)
	return sanitizedHeaders, resp, err
}

// applyChannelAuth applies key-pool credentials according to the channel auth type.
func applyChannelAuth(req *http.Request, ch *model.Channel, key string, body []byte) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil
	}

	authType := strings.ToLower(strings.TrimSpace(ch.AuthType))
	if authType == "" {
		authType = "bearer"
	}

	switch authType {
	case "bearer":
		if req.Header.Get("Authorization") == "" {
			req.Header.Set("Authorization", "Bearer "+key)
		}
	case "query_param":
		paramName := strings.TrimSpace(ch.AuthParamName)
		if paramName == "" {
			paramName = "key"
		}
		q := req.URL.Query()
		if q.Get(paramName) == "" {
			q.Set(paramName, key)
			req.URL.RawQuery = q.Encode()
		}
	case "basic":
		if req.Header.Get("Authorization") != "" {
			return nil
		}
		parts := strings.SplitN(key, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("basic key format should be user:pass")
		}
		req.SetBasicAuth(parts[0], parts[1])
	case "sigv4":
		if req.Header.Get("Authorization") != "" {
			return nil
		}
		region := strings.TrimSpace(ch.AuthRegion)
		if region == "" {
			region = "us-east-1"
		}
		serviceName := strings.TrimSpace(ch.AuthService)
		if serviceName == "" {
			serviceName = "execute-api"
		}
		return signSigV4(req, key, region, serviceName, body)
	default:
		return fmt.Errorf("unsupported auth_type: %s", authType)
	}
	return nil
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
// llmRefundCredits 按优先级退款：优先退回通用余额，再退专属模型积分（与扣款顺序相反）。
// 调用时自动更新 gin context 中记录的已扣款数量，保证多次退款不会重复退回。

// llmChargeExtra 结算补扣：优先消耗专属模型积分，不足部分再扣通用余额。
// 同步更新 gin context 中记录的已扣款数量，保证后续退款计算正确。
// 返回从专属模型积分中扣除的数量，供 WriteTx 记录。
func llmChargeExtra(c *gin.Context, userID, amount int64) int64 {
	if amount <= 0 {
		return 0
	}
	ctx := c.Request.Context()

	modelExtraCharged := int64(0)
	if rk, ok := c.Get("model_credit_routing_key"); ok {
		if routingKey, ok := rk.(string); ok && routingKey != "" {
			modelExtraCharged, _ = billing.ChargeModelCredit(ctx, userID, routingKey, amount)
		}
	}

	generalExtraCharged := amount - modelExtraCharged
	if generalExtraCharged > 0 {
		_ = billing.Charge(ctx, userID, generalExtraCharged)
	}

	// 更新 context 中的累计扣款记录，供后续退款使用
	if modelExtraCharged > 0 {
		mc := int64(0)
		if v, ok := c.Get("model_credit_charged"); ok {
			if val, ok := v.(int64); ok {
				mc = val
			}
		}
		c.Set("model_credit_charged", mc+modelExtraCharged)
	}
	if generalExtraCharged > 0 {
		gc := int64(0)
		if v, ok := c.Get("model_credit_general_charged"); ok {
			if val, ok := v.(int64); ok {
				gc = val
			}
		}
		c.Set("model_credit_general_charged", gc+generalExtraCharged)
	}
	return modelExtraCharged
}

// llmRefundCredits 退款：优先退通用余额，再退专属模型积分（与扣款顺序相反）。
// 返回从专属模型积分中退还的数量，供 WriteTx 记录。
func llmRefundCredits(c *gin.Context, userID, amount int64) int64 {
	if amount <= 0 {
		return 0
	}
	ctx := c.Request.Context()

	// 读取本次请求的扣款记录
	modelCharged := int64(0)
	if mc, ok := c.Get("model_credit_charged"); ok {
		if v, ok := mc.(int64); ok {
			modelCharged = v
		}
	}
	generalCharged := int64(0)
	if gc, ok := c.Get("model_credit_general_charged"); ok {
		if v, ok := gc.(int64); ok {
			generalCharged = v
		}
	}

	// 优先退通用余额，再退模型积分
	generalRefund := int64(0)
	modelRefund := int64(0)
	if amount <= generalCharged {
		generalRefund = amount
	} else {
		generalRefund = generalCharged
		modelRefund = amount - generalCharged
		if modelRefund > modelCharged {
			modelRefund = modelCharged
		}
	}

	if generalRefund > 0 {
		_ = billing.Refund(ctx, userID, generalRefund)
		c.Set("model_credit_general_charged", generalCharged-generalRefund)
	}
	if modelRefund > 0 {
		if rk, ok := c.Get("model_credit_routing_key"); ok {
			if routingKey, ok := rk.(string); ok && routingKey != "" {
				_ = billing.RefundModelCredit(ctx, userID, routingKey, modelRefund)
				c.Set("model_credit_charged", modelCharged-modelRefund)
			}
		}
	}
	return modelRefund
}

func llmRefundAndAbort(c *gin.Context, corrID string, userID, credits, upstreamCost, poolKeyIDVal int64, upstreamStatus int, errMsg string) {
	userMsg := service.UserFacingErrorMessage(errMsg)
	if userMsg != errMsg {
		log.Printf("[llm] request %s failed: %s", corrID, errMsg)
	}
	if credits > 0 {
		mcRefunded := llmRefundCredits(c, userID, credits)
		_ = service.WriteTx(c.Request.Context(), userID, 0, 0, poolKeyIDVal, corrID, "refund", credits, upstreamCost, mcRefunded, model.JSON{"reason": "upstream_error"})
	}
	if corrID != "" {
		_, _ = db.Engine.Where("corr_id = ?", corrID).Cols("status", "upstream_status", "error_msg").
			Update(&model.LLMLog{Status: "error", UpstreamStatus: upstreamStatus, ErrorMsg: errMsg})
	}
	c.JSON(http.StatusBadGateway, gin.H{"error": userMsg})
}
