package protocol

import (
	"encoding/json"
	"strings"
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
