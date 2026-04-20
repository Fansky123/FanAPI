package protocol

import (
	"encoding/json"
	"strings"

	"github.com/google/uuid"
)

// SSEConverter converts SSE lines from one protocol format to another.
// Convert is called for each line read from the upstream response body.
// Flush is called once after the scanner reaches EOF to emit any trailing lines.
// Both methods return zero or more output lines; each will be written followed by "\n".
type SSEConverter interface {
	Convert(line string) []string
	Flush() []string
}

// NewSSEConverter returns an SSEConverter for the given (sourceProto → clientProto) pair.
// Returns nil when no conversion is needed (same format, or unsupported pair).
func NewSSEConverter(sourceProto, clientProto string) SSEConverter {
	if sourceProto == clientProto {
		return nil
	}
	switch {
	case sourceProto == ProtocolClaude && clientProto == ProtocolOpenAI:
		return &claudeToOpenAISSE{}
	case sourceProto == ProtocolGemini && clientProto == ProtocolOpenAI:
		return &geminiToOpenAISSE{}
	case sourceProto == ProtocolOpenAI && clientProto == ProtocolClaude:
		return &openAIToClaudeSSE{}
	case sourceProto == ProtocolOpenAI && clientProto == ProtocolResponses:
		return &openAIToResponsesSSE{}
	default:
		// Unsupported pair: pass lines through unchanged so the client at least gets something.
		return nil
	}
}

// ─────────────────────────────────────────────
// Claude SSE → OpenAI SSE
// ─────────────────────────────────────────────

type claudeToOpenAISSE struct {
	msgID       string
	model       string
	lastEvent   string
	inputTokens int64
	sentRole    bool
	doneSent    bool
}

func (c *claudeToOpenAISSE) Convert(line string) []string {
	if line == "" {
		return nil // skip Claude's blank event delimiters; we emit our own
	}
	if strings.HasPrefix(line, "event: ") {
		c.lastEvent = strings.TrimPrefix(line, "event: ")
		return nil
	}
	if !strings.HasPrefix(line, "data: ") {
		return nil
	}
	payload := strings.TrimPrefix(line, "data: ")

	var chunk map[string]interface{}
	if json.Unmarshal([]byte(payload), &chunk) != nil {
		return nil
	}

	switch c.lastEvent {
	case "message_start":
		if msg, ok := chunk["message"].(map[string]interface{}); ok {
			c.msgID, _ = msg["id"].(string)
			c.model, _ = msg["model"].(string)
			if usg, ok := msg["usage"].(map[string]interface{}); ok {
				if n, _ := usg["input_tokens"].(float64); n > 0 {
					c.inputTokens = int64(n)
				}
			}
		}
		return c.emitRoleChunk()

	case "content_block_delta":
		if delta, ok := chunk["delta"].(map[string]interface{}); ok {
			if text, _ := delta["text"].(string); text != "" {
				return c.emitTextChunk(text)
			}
		}

	case "message_delta":
		stopReason := "stop"
		var outputTokens int64
		if delta, ok := chunk["delta"].(map[string]interface{}); ok {
			if sr, _ := delta["stop_reason"].(string); sr != "" {
				switch sr {
				case "max_tokens":
					stopReason = "length"
				case "tool_use":
					stopReason = "tool_calls"
				}
			}
		}
		if usg, ok := chunk["usage"].(map[string]interface{}); ok {
			if n, _ := usg["output_tokens"].(float64); n > 0 {
				outputTokens = int64(n)
			}
		}
		return c.emitFinishChunk(stopReason, outputTokens)

	case "message_stop":
		if !c.doneSent {
			c.doneSent = true
			return []string{"data: [DONE]", ""}
		}
	}
	return nil
}

func (c *claudeToOpenAISSE) Flush() []string {
	if !c.doneSent {
		c.doneSent = true
		return []string{"data: [DONE]", ""}
	}
	return nil
}

func (c *claudeToOpenAISSE) emitRoleChunk() []string {
	if c.sentRole {
		return nil
	}
	c.sentRole = true
	out := map[string]interface{}{
		"id":     c.msgID,
		"object": "chat.completion.chunk",
		"model":  c.model,
		"choices": []interface{}{map[string]interface{}{
			"index":         0,
			"delta":         map[string]interface{}{"role": "assistant", "content": ""},
			"finish_reason": nil,
		}},
	}
	b, _ := json.Marshal(out)
	return []string{"data: " + string(b), ""}
}

func (c *claudeToOpenAISSE) emitTextChunk(text string) []string {
	out := map[string]interface{}{
		"id":     c.msgID,
		"object": "chat.completion.chunk",
		"model":  c.model,
		"choices": []interface{}{map[string]interface{}{
			"index":         0,
			"delta":         map[string]interface{}{"content": text},
			"finish_reason": nil,
		}},
	}
	b, _ := json.Marshal(out)
	return []string{"data: " + string(b), ""}
}

func (c *claudeToOpenAISSE) emitFinishChunk(reason string, outputTokens int64) []string {
	out := map[string]interface{}{
		"id":     c.msgID,
		"object": "chat.completion.chunk",
		"model":  c.model,
		"choices": []interface{}{map[string]interface{}{
			"index":         0,
			"delta":         map[string]interface{}{},
			"finish_reason": reason,
		}},
		"usage": map[string]interface{}{
			"prompt_tokens":     c.inputTokens,
			"completion_tokens": outputTokens,
			"total_tokens":      c.inputTokens + outputTokens,
		},
	}
	b, _ := json.Marshal(out)
	return []string{"data: " + string(b), ""}
}

// ─────────────────────────────────────────────
// Gemini SSE → OpenAI SSE
// ─────────────────────────────────────────────

type geminiToOpenAISSE struct {
	doneSent bool
}

func (g *geminiToOpenAISSE) Convert(line string) []string {
	if line == "" || !strings.HasPrefix(line, "data: ") {
		return nil
	}
	payload := strings.TrimPrefix(line, "data: ")

	var chunk map[string]interface{}
	if json.Unmarshal([]byte(payload), &chunk) != nil {
		return nil
	}

	var text string
	var finishReason interface{} = nil
	isFinish := false

	if candidates, ok := chunk["candidates"].([]interface{}); ok && len(candidates) > 0 {
		if cand, ok := candidates[0].(map[string]interface{}); ok {
			if contentObj, ok := cand["content"].(map[string]interface{}); ok {
				if parts, ok := contentObj["parts"].([]interface{}); ok {
					for _, p := range parts {
						if pm, ok := p.(map[string]interface{}); ok {
							if t, ok := pm["text"].(string); ok {
								text += t
							}
						}
					}
				}
			}
			if fr, ok := cand["finishReason"].(string); ok && fr != "" && fr != "FINISH_REASON_UNSPECIFIED" {
				isFinish = true
				if fr == "MAX_TOKENS" {
					finishReason = "length"
				} else {
					finishReason = "stop"
				}
			}
		}
	}

	deltaChunk := map[string]interface{}{
		"id":     "chatcmpl-gemini",
		"object": "chat.completion.chunk",
		"model":  "",
		"choices": []interface{}{map[string]interface{}{
			"index":         0,
			"delta":         map[string]interface{}{"content": text},
			"finish_reason": finishReason,
		}},
	}

	if meta, ok := chunk["usageMetadata"].(map[string]interface{}); ok {
		in, _ := meta["promptTokenCount"].(float64)
		out, _ := meta["candidatesTokenCount"].(float64)
		deltaChunk["usage"] = map[string]interface{}{
			"prompt_tokens":     int64(in),
			"completion_tokens": int64(out),
			"total_tokens":      int64(in + out),
		}
	}

	b, _ := json.Marshal(deltaChunk)
	result := []string{"data: " + string(b), ""}

	if isFinish && !g.doneSent {
		g.doneSent = true
		result = append(result, "data: [DONE]", "")
	}

	if text == "" && !isFinish {
		return nil // 跳过没有内容且非结束的中间块（如纯 usageMetadata chunk）
	}

	return result
}

func (g *geminiToOpenAISSE) Flush() []string {
	if !g.doneSent {
		g.doneSent = true
		return []string{"data: [DONE]", ""}
	}
	return nil
}

// ─────────────────────────────────────────────
// OpenAI SSE → Claude SSE
// ─────────────────────────────────────────────

type openAIToClaudeSSE struct {
	msgID        string
	model        string
	inputTokens  int64
	outputTokens int64
	sentStart    bool
	doneSent     bool
}

func (o *openAIToClaudeSSE) Convert(line string) []string {
	if line == "" {
		return nil
	}
	if !strings.HasPrefix(line, "data: ") {
		return nil
	}
	payload := strings.TrimPrefix(line, "data: ")
	if payload == "[DONE]" {
		o.doneSent = true
		return o.stopEvents()
	}

	var chunk map[string]interface{}
	if json.Unmarshal([]byte(payload), &chunk) != nil {
		return nil
	}

	if o.msgID == "" {
		o.msgID, _ = chunk["id"].(string)
		o.model, _ = chunk["model"].(string)
	}
	if usg, ok := chunk["usage"].(map[string]interface{}); ok {
		if pt, _ := usg["prompt_tokens"].(float64); pt > 0 {
			o.inputTokens = int64(pt)
		}
		if ct, _ := usg["completion_tokens"].(float64); ct > 0 {
			o.outputTokens = int64(ct)
		}
	}

	choices, _ := chunk["choices"].([]interface{})
	if len(choices) == 0 {
		return nil
	}
	choice, _ := choices[0].(map[string]interface{})
	if choice == nil {
		return nil
	}

	var result []string

	if !o.sentStart {
		o.sentStart = true
		result = append(result, o.messageStartLines()...)
		result = append(result, o.contentBlockStartLines()...)
	}

	delta, _ := choice["delta"].(map[string]interface{})
	if content, _ := delta["content"].(string); content != "" {
		result = append(result, o.contentDeltaLines(content)...)
	}

	return result
}

func (o *openAIToClaudeSSE) Flush() []string {
	if !o.doneSent {
		return o.stopEvents()
	}
	return nil
}

func (o *openAIToClaudeSSE) messageStartLines() []string {
	msg := map[string]interface{}{
		"type": "message_start",
		"message": map[string]interface{}{
			"id":    o.msgID,
			"type":  "message",
			"role":  "assistant",
			"model": o.model,
			"usage": map[string]interface{}{"input_tokens": o.inputTokens, "output_tokens": 0},
		},
	}
	b, _ := json.Marshal(msg)
	return []string{"event: message_start", "data: " + string(b), ""}
}

func (o *openAIToClaudeSSE) contentBlockStartLines() []string {
	return []string{
		"event: content_block_start",
		`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
		"",
		"event: ping",
		`data: {"type":"ping"}`,
		"",
	}
}

func (o *openAIToClaudeSSE) contentDeltaLines(text string) []string {
	data := map[string]interface{}{
		"type":  "content_block_delta",
		"index": 0,
		"delta": map[string]interface{}{"type": "text_delta", "text": text},
	}
	b, _ := json.Marshal(data)
	return []string{"event: content_block_delta", "data: " + string(b), ""}
}

func (o *openAIToClaudeSSE) stopEvents() []string {
	outTok := o.outputTokens
	msgDelta := map[string]interface{}{
		"type":  "message_delta",
		"delta": map[string]interface{}{"stop_reason": "end_turn", "stop_sequence": nil},
		"usage": map[string]interface{}{"output_tokens": outTok},
	}
	b, _ := json.Marshal(msgDelta)
	return []string{
		"event: content_block_stop",
		`data: {"type":"content_block_stop","index":0}`,
		"",
		"event: message_delta",
		"data: " + string(b),
		"",
		"event: message_stop",
		`data: {"type":"message_stop"}`,
		"",
	}
}

// ─────────────────────────────────────────────
// OpenAI SSE → Responses API SSE
//
// Codex CLI 使用 OpenAI Responses API（POST /v1/responses），
// 其 SSE 事件格式与 Chat Completions 完全不同。
// 此转换器将上游 OpenAI Chat Completions SSE 流转换为 Responses API SSE 事件。
//
// 事件顺序：
//   response.created → response.output_item.added → response.content_part.added
//   → (N×) response.output_text.delta
//   → response.output_text.done → response.content_part.done
//   → response.output_item.done → response.completed
// ─────────────────────────────────────────────

type openAIToResponsesSSE struct {
	respID   string
	itemID   string
	model    string
	fullText string
	// 状态标记
	headerSent bool
	doneSent   bool
	// token 统计
	inputTokens  int64
	outputTokens int64
}

func (r *openAIToResponsesSSE) Convert(line string) []string {
	if line == "" {
		return nil
	}
	if !strings.HasPrefix(line, "data: ") {
		return nil
	}
	payload := strings.TrimPrefix(line, "data: ")
	if payload == "[DONE]" {
		return nil // 在 Flush 中处理收尾事件
	}

	var chunk map[string]interface{}
	if json.Unmarshal([]byte(payload), &chunk) != nil {
		return nil
	}

	// 首个 chunk：提取 id 和 model，发送 header 事件
	var out []string
	if !r.headerSent {
		r.headerSent = true
		if id, ok := chunk["id"].(string); ok {
			r.respID = id
		} else {
			r.respID = "resp_" + newShortID()
		}
		r.itemID = "msg_" + newShortID()
		if m, ok := chunk["model"].(string); ok {
			r.model = m
		}
		out = append(out, r.emitCreated()...)
		out = append(out, r.emitOutputItemAdded()...)
		out = append(out, r.emitContentPartAdded()...)
	}

	// 收集 usage（最后一个 chunk 会携带）
	if usg, ok := chunk["usage"].(map[string]interface{}); ok {
		if n, _ := usg["prompt_tokens"].(float64); n > 0 {
			r.inputTokens = int64(n)
		}
		if n, _ := usg["completion_tokens"].(float64); n > 0 {
			r.outputTokens = int64(n)
		}
	}

	// 提取 delta 文本
	choices, _ := chunk["choices"].([]interface{})
	if len(choices) == 0 {
		return out
	}
	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return out
	}

	// delta 文本增量
	if delta, ok := choice["delta"].(map[string]interface{}); ok {
		if text, ok := delta["content"].(string); ok && text != "" {
			r.fullText += text
			out = append(out, r.emitTextDelta(text)...)
		}
	}

	return out
}

func (r *openAIToResponsesSSE) Flush() []string {
	if r.doneSent {
		return nil
	}
	r.doneSent = true
	var out []string
	out = append(out, r.emitTextDone()...)
	out = append(out, r.emitContentPartDone()...)
	out = append(out, r.emitOutputItemDone()...)
	out = append(out, r.emitCompleted()...)
	return out
}

func (r *openAIToResponsesSSE) emitCreated() []string {
	resp := map[string]interface{}{
		"type": "response.created",
		"response": map[string]interface{}{
			"id":     r.respID,
			"object": "response",
			"status": "in_progress",
			"model":  r.model,
			"output": []interface{}{},
		},
	}
	b, _ := json.Marshal(resp)
	return []string{"event: response.created", "data: " + string(b), ""}
}

func (r *openAIToResponsesSSE) emitOutputItemAdded() []string {
	item := map[string]interface{}{
		"type":         "response.output_item.added",
		"output_index": 0,
		"item": map[string]interface{}{
			"id":      r.itemID,
			"type":    "message",
			"role":    "assistant",
			"content": []interface{}{},
			"status":  "in_progress",
		},
	}
	b, _ := json.Marshal(item)
	return []string{"event: response.output_item.added", "data: " + string(b), ""}
}

func (r *openAIToResponsesSSE) emitContentPartAdded() []string {
	ev := map[string]interface{}{
		"type":          "response.content_part.added",
		"item_id":       r.itemID,
		"output_index":  0,
		"content_index": 0,
		"part": map[string]interface{}{
			"type": "output_text",
			"text": "",
		},
	}
	b, _ := json.Marshal(ev)
	return []string{"event: response.content_part.added", "data: " + string(b), ""}
}

func (r *openAIToResponsesSSE) emitTextDelta(delta string) []string {
	ev := map[string]interface{}{
		"type":          "response.output_text.delta",
		"item_id":       r.itemID,
		"output_index":  0,
		"content_index": 0,
		"delta":         delta,
	}
	b, _ := json.Marshal(ev)
	return []string{"event: response.output_text.delta", "data: " + string(b), ""}
}

func (r *openAIToResponsesSSE) emitTextDone() []string {
	ev := map[string]interface{}{
		"type":          "response.output_text.done",
		"item_id":       r.itemID,
		"output_index":  0,
		"content_index": 0,
		"text":          r.fullText,
	}
	b, _ := json.Marshal(ev)
	return []string{"event: response.output_text.done", "data: " + string(b), ""}
}

func (r *openAIToResponsesSSE) emitContentPartDone() []string {
	ev := map[string]interface{}{
		"type":          "response.content_part.done",
		"item_id":       r.itemID,
		"output_index":  0,
		"content_index": 0,
		"part": map[string]interface{}{
			"type": "output_text",
			"text": r.fullText,
		},
	}
	b, _ := json.Marshal(ev)
	return []string{"event: response.content_part.done", "data: " + string(b), ""}
}

func (r *openAIToResponsesSSE) emitOutputItemDone() []string {
	ev := map[string]interface{}{
		"type":         "response.output_item.done",
		"output_index": 0,
		"item": map[string]interface{}{
			"id":     r.itemID,
			"type":   "message",
			"role":   "assistant",
			"status": "completed",
			"content": []interface{}{
				map[string]interface{}{
					"type": "output_text",
					"text": r.fullText,
				},
			},
		},
	}
	b, _ := json.Marshal(ev)
	return []string{"event: response.output_item.done", "data: " + string(b), ""}
}

func (r *openAIToResponsesSSE) emitCompleted() []string {
	usage := map[string]interface{}{
		"input_tokens":  r.inputTokens,
		"output_tokens": r.outputTokens,
		"total_tokens":  r.inputTokens + r.outputTokens,
	}
	resp := map[string]interface{}{
		"type": "response.completed",
		"response": map[string]interface{}{
			"id":     r.respID,
			"object": "response",
			"status": "completed",
			"model":  r.model,
			"output": []interface{}{
				map[string]interface{}{
					"id":     r.itemID,
					"type":   "message",
					"role":   "assistant",
					"status": "completed",
					"content": []interface{}{
						map[string]interface{}{
							"type": "output_text",
							"text": r.fullText,
						},
					},
				},
			},
			"usage": usage,
		},
	}
	b, _ := json.Marshal(resp)
	return []string{"event: response.completed", "data: " + string(b), ""}
}

// newShortID 生成短 ID（不含横线）供 Responses API 的 id 字段使用。
func newShortID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}
