// Package protocol handles request/response format conversion between
// OpenAI, Claude (Anthropic), and Gemini (Google) API formats.
//
// Conversion matrix (input format → channel protocol):
//   - OpenAI  → openai  : pass-through (no-op)
//   - OpenAI  → claude  : ConvertRequest / ConvertSyncResponse
//   - OpenAI  → gemini  : ConvertRequest / ConvertSyncResponse
//
// All functions operate on plain map[string]interface{} so they compose
// cleanly with the existing request_script / response_script JS hooks.
package protocol

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	ProtocolOpenAI = "openai"
	ProtocolClaude = "claude"
	ProtocolGemini = "gemini"
)

// ConvertRequest converts an OpenAI-format request map to the target protocol.
// Returns the same map unchanged when targetProtocol == "openai".
func ConvertRequest(req map[string]interface{}, targetProtocol string) (map[string]interface{}, error) {
	switch targetProtocol {
	case ProtocolClaude:
		return openAIToClaude(req)
	case ProtocolGemini:
		return openAIToGemini(req)
	default:
		return req, nil
	}
}

// ConvertSyncResponse converts a sync (non-streaming) response body from the
// upstream protocol back to OpenAI format.
func ConvertSyncResponse(respBody []byte, sourceProtocol string) ([]byte, error) {
	switch sourceProtocol {
	case ProtocolClaude:
		return claudeToOpenAI(respBody)
	case ProtocolGemini:
		return geminiToOpenAI(respBody)
	default:
		return respBody, nil
	}
}

// NormalizeUsage extracts {prompt_tokens, completion_tokens} from a raw
// upstream response according to the source protocol.
func NormalizeUsage(resp map[string]interface{}, sourceProtocol string) map[string]interface{} {
	switch sourceProtocol {
	case ProtocolClaude:
		if usg, ok := resp["usage"].(map[string]interface{}); ok {
			in, _ := usg["input_tokens"].(float64)
			out, _ := usg["output_tokens"].(float64)
			return map[string]interface{}{
				"prompt_tokens":     int64(in),
				"completion_tokens": int64(out),
				"total_tokens":      int64(in + out),
			}
		}
	case ProtocolGemini:
		if meta, ok := resp["usageMetadata"].(map[string]interface{}); ok {
			in, _ := meta["promptTokenCount"].(float64)
			out, _ := meta["candidatesTokenCount"].(float64)
			return map[string]interface{}{
				"prompt_tokens":     int64(in),
				"completion_tokens": int64(out),
				"total_tokens":      int64(in + out),
			}
		}
	default:
		if usg, ok := resp["usage"].(map[string]interface{}); ok {
			pt, _ := usg["prompt_tokens"].(float64)
			ct, _ := usg["completion_tokens"].(float64)
			return map[string]interface{}{
				"prompt_tokens":     int64(pt),
				"completion_tokens": int64(ct),
				"total_tokens":      int64(pt + ct),
			}
		}
	}
	return nil
}

// ─────────────────────────────────────────────
// OpenAI → Claude
// ─────────────────────────────────────────────

func openAIToClaude(req map[string]interface{}) (map[string]interface{}, error) {
	out := make(map[string]interface{})

	if m, ok := req["model"].(string); ok {
		out["model"] = m
	}

	// max_tokens (Claude requires this field)
	if mt, ok := req["max_tokens"]; ok {
		out["max_tokens"] = mt
	} else if mc, ok := req["max_completion_tokens"]; ok {
		out["max_tokens"] = mc
	} else {
		out["max_tokens"] = 4096
	}

	if t, ok := req["temperature"]; ok {
		out["temperature"] = t
	}
	if tp, ok := req["top_p"]; ok {
		out["top_p"] = tp
	}
	if s, ok := req["stream"]; ok {
		out["stream"] = s
	}

	// system + messages
	messages, _ := req["messages"].([]interface{})
	var systemMsg string
	var claudeMessages []map[string]interface{}

	for _, m := range messages {
		msg, ok := m.(map[string]interface{})
		if !ok {
			continue
		}
		role, _ := msg["role"].(string)
		switch role {
		case "system":
			if c, ok := msg["content"].(string); ok {
				if systemMsg != "" {
					systemMsg += "\n"
				}
				systemMsg += c
			}
		case "user", "assistant":
			claudeMsg := map[string]interface{}{"role": role}
			switch c := msg["content"].(type) {
			case string:
				claudeMsg["content"] = []map[string]interface{}{
					{"type": "text", "text": c},
				}
			case []interface{}:
				// already array — pass through (vision etc.)
				claudeMsg["content"] = c
			default:
				claudeMsg["content"] = msg["content"]
			}
			claudeMessages = append(claudeMessages, claudeMsg)
		case "tool":
			// tool result
			toolCallID, _ := msg["tool_call_id"].(string)
			content, _ := msg["content"].(string)
			claudeMessages = append(claudeMessages, map[string]interface{}{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type":        "tool_result",
						"tool_use_id": toolCallID,
						"content":     content,
					},
				},
			})
		}
	}

	if len(claudeMessages) == 0 {
		return nil, fmt.Errorf("no valid messages after conversion")
	}
	if systemMsg != "" {
		out["system"] = systemMsg
	}
	out["messages"] = claudeMessages

	// tools
	if tools, ok := req["tools"].([]interface{}); ok && len(tools) > 0 {
		out["tools"] = convertOpenAIToolsToClaude(tools)
	}

	// tool_choice
	if tc, ok := req["tool_choice"]; ok {
		out["tool_choice"] = convertToolChoiceToClaude(tc)
	}

	return out, nil
}

func convertOpenAIToolsToClaude(tools []interface{}) []map[string]interface{} {
	var out []map[string]interface{}
	for _, t := range tools {
		tm, ok := t.(map[string]interface{})
		if !ok {
			continue
		}
		if tm["type"] != "function" {
			continue
		}
		fn, _ := tm["function"].(map[string]interface{})
		if fn == nil {
			continue
		}
		tool := map[string]interface{}{
			"name": fn["name"],
		}
		if desc, ok := fn["description"].(string); ok {
			tool["description"] = desc
		}
		if params, ok := fn["parameters"]; ok {
			tool["input_schema"] = params
		}
		out = append(out, tool)
	}
	return out
}

func convertToolChoiceToClaude(tc interface{}) interface{} {
	switch v := tc.(type) {
	case string:
		switch v {
		case "auto":
			return map[string]interface{}{"type": "auto"}
		case "none":
			return map[string]interface{}{"type": "any"}
		}
		return map[string]interface{}{"type": "auto"}
	case map[string]interface{}:
		if fn, ok := v["function"].(map[string]interface{}); ok {
			return map[string]interface{}{"type": "tool", "name": fn["name"]}
		}
	}
	return map[string]interface{}{"type": "auto"}
}

// ─────────────────────────────────────────────
// Claude → OpenAI (sync response body)
// ─────────────────────────────────────────────

func claudeToOpenAI(body []byte) ([]byte, error) {
	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return body, nil // pass through on parse error
	}

	id, _ := resp["id"].(string)
	model, _ := resp["model"].(string)

	// Extract content
	var content string
	var toolCalls []map[string]interface{}
	if contents, ok := resp["content"].([]interface{}); ok {
		for _, c := range contents {
			cm, ok := c.(map[string]interface{})
			if !ok {
				continue
			}
			switch cm["type"] {
			case "text":
				content += cm["text"].(string)
			case "tool_use":
				tcID, _ := cm["id"].(string)
				tcName, _ := cm["name"].(string)
				inputBytes, _ := json.Marshal(cm["input"])
				toolCalls = append(toolCalls, map[string]interface{}{
					"id":   tcID,
					"type": "function",
					"function": map[string]interface{}{
						"name":      tcName,
						"arguments": string(inputBytes),
					},
				})
			}
		}
	}

	// finish_reason
	stopReason, _ := resp["stop_reason"].(string)
	finishReason := "stop"
	switch stopReason {
	case "max_tokens":
		finishReason = "length"
	case "tool_use":
		finishReason = "tool_calls"
	case "end_turn":
		finishReason = "stop"
	}

	delta := map[string]interface{}{"role": "assistant", "content": content}
	if len(toolCalls) > 0 {
		delta["content"] = nil
		delta["tool_calls"] = toolCalls
	}
	choice := map[string]interface{}{
		"index":         0,
		"message":       delta,
		"finish_reason": finishReason,
	}

	// usage
	usage := map[string]interface{}{}
	if usg, ok := resp["usage"].(map[string]interface{}); ok {
		in, _ := usg["input_tokens"].(float64)
		out, _ := usg["output_tokens"].(float64)
		usage["prompt_tokens"] = int64(in)
		usage["completion_tokens"] = int64(out)
		usage["total_tokens"] = int64(in + out)
	}

	out := map[string]interface{}{
		"id":      id,
		"object":  "chat.completion",
		"model":   model,
		"choices": []interface{}{choice},
		"usage":   usage,
	}
	return json.Marshal(out)
}

// ─────────────────────────────────────────────
// OpenAI → Gemini
// ─────────────────────────────────────────────

func openAIToGemini(req map[string]interface{}) (map[string]interface{}, error) {
	out := make(map[string]interface{})

	messages, _ := req["messages"].([]interface{})
	var systemParts []map[string]interface{}
	var contents []map[string]interface{}

	for _, m := range messages {
		msg, ok := m.(map[string]interface{})
		if !ok {
			continue
		}
		role, _ := msg["role"].(string)
		switch role {
		case "system":
			if c, ok := msg["content"].(string); ok {
				systemParts = append(systemParts, map[string]interface{}{"text": c})
			}
		case "user":
			contents = append(contents, map[string]interface{}{
				"role":  "user",
				"parts": contentToParts(msg["content"]),
			})
		case "assistant":
			contents = append(contents, map[string]interface{}{
				"role":  "model",
				"parts": contentToParts(msg["content"]),
			})
		case "tool":
			toolCallID, _ := msg["tool_call_id"].(string)
			content, _ := msg["content"].(string)
			contents = append(contents, map[string]interface{}{
				"role": "user",
				"parts": []map[string]interface{}{
					{
						"functionResponse": map[string]interface{}{
							"name":     toolCallID,
							"response": map[string]interface{}{"output": content},
						},
					},
				},
			})
		}
	}

	out["contents"] = contents

	if len(systemParts) > 0 {
		out["systemInstruction"] = map[string]interface{}{"parts": systemParts}
	}

	// generationConfig
	genCfg := map[string]interface{}{}
	if mt, ok := req["max_tokens"]; ok {
		genCfg["maxOutputTokens"] = mt
	} else if mc, ok := req["max_completion_tokens"]; ok {
		genCfg["maxOutputTokens"] = mc
	}
	if t, ok := req["temperature"]; ok {
		genCfg["temperature"] = t
	}
	if tp, ok := req["top_p"]; ok {
		genCfg["topP"] = tp
	}
	// stream is controlled via URL suffix for Gemini, not body field

	// response_modalities → generationConfig.responseModalities
	// 用于图片生成等需要 IMAGE 输出的场景（如 gemini-2.5-flash-image）
	if rm, ok := req["response_modalities"]; ok {
		genCfg["responseModalities"] = rm
	}

	if len(genCfg) > 0 {
		out["generationConfig"] = genCfg
	}

	// tools
	if tools, ok := req["tools"].([]interface{}); ok && len(tools) > 0 {
		out["tools"] = []map[string]interface{}{
			{"functionDeclarations": convertOpenAIToolsToGemini(tools)},
		}
	}

	return out, nil
}

func convertOpenAIToolsToGemini(tools []interface{}) []map[string]interface{} {
	var out []map[string]interface{}
	for _, t := range tools {
		tm, ok := t.(map[string]interface{})
		if !ok {
			continue
		}
		if tm["type"] != "function" {
			continue
		}
		fn, _ := tm["function"].(map[string]interface{})
		if fn == nil {
			continue
		}
		decl := map[string]interface{}{
			"name": fn["name"],
		}
		if desc, ok := fn["description"].(string); ok {
			decl["description"] = desc
		}
		if params, ok := fn["parameters"]; ok {
			decl["parameters"] = params
		}
		out = append(out, decl)
	}
	return out
}

func contentToParts(content interface{}) []map[string]interface{} {
	switch c := content.(type) {
	case string:
		return []map[string]interface{}{{"text": c}}
	case []interface{}:
		var parts []map[string]interface{}
		for _, item := range c {
			im, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			switch im["type"] {
			case "text":
				parts = append(parts, map[string]interface{}{"text": im["text"]})
			case "image_url":
				if iu, ok := im["image_url"].(map[string]interface{}); ok {
					if url, ok := iu["url"].(string); ok {
						if strings.HasPrefix(url, "data:") {
							// base64 inline
							parts = append(parts, map[string]interface{}{
								"inlineData": map[string]interface{}{
									"mimeType": extractMimeType(url),
									"data":     extractBase64Data(url),
								},
							})
						} else {
							parts = append(parts, map[string]interface{}{
								"fileData": map[string]interface{}{
									"mimeType": "image/jpeg",
									"fileUri":  url,
								},
							})
						}
					}
				}
			}
		}
		return parts
	}
	return []map[string]interface{}{{"text": fmt.Sprintf("%v", content)}}
}

// ─────────────────────────────────────────────
// Gemini → OpenAI (sync response body)
// ─────────────────────────────────────────────

func geminiToOpenAI(body []byte) ([]byte, error) {
	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return body, nil
	}

	finishReason := "stop"
	var content string
	var toolCalls []map[string]interface{}
	var inlineImages []map[string]interface{}

	candidates, _ := resp["candidates"].([]interface{})
	if len(candidates) > 0 {
		cand, _ := candidates[0].(map[string]interface{})
		if cand != nil {
			if fr, ok := cand["finishReason"].(string); ok {
				switch fr {
				case "MAX_TOKENS":
					finishReason = "length"
				case "STOP":
					finishReason = "stop"
				}
			}
			if contentObj, ok := cand["content"].(map[string]interface{}); ok {
				parts, _ := contentObj["parts"].([]interface{})
				for _, p := range parts {
					pm, ok := p.(map[string]interface{})
					if !ok {
						continue
					}
					if text, ok := pm["text"].(string); ok {
						content += text
					}
					if id, ok := pm["inlineData"].(map[string]interface{}); ok {
						mime, _ := id["mimeType"].(string)
						data, _ := id["data"].(string)
						if mime != "" && data != "" {
							inlineImages = append(inlineImages, map[string]interface{}{
								"type": "image_url",
								"image_url": map[string]interface{}{
									"url": "data:" + mime + ";base64," + data,
								},
							})
						}
					}
					if fc, ok := pm["functionCall"].(map[string]interface{}); ok {
						name, _ := fc["name"].(string)
						argsBytes, _ := json.Marshal(fc["args"])
						toolCalls = append(toolCalls, map[string]interface{}{
							"id":   "call_" + name,
							"type": "function",
							"function": map[string]interface{}{
								"name":      name,
								"arguments": string(argsBytes),
							},
						})
						finishReason = "tool_calls"
					}
				}
			}
		}
	}

	// 构建 message content：纯文本时用字符串，含图片时用 content array
	var messageContent interface{}
	if len(inlineImages) > 0 {
		var parts []map[string]interface{}
		if content != "" {
			parts = append(parts, map[string]interface{}{"type": "text", "text": content})
		}
		parts = append(parts, inlineImages...)
		messageContent = parts
	} else {
		messageContent = content
	}

	message := map[string]interface{}{"role": "assistant", "content": messageContent}
	if len(toolCalls) > 0 {
		message["content"] = nil
		message["tool_calls"] = toolCalls
	}
	choice := map[string]interface{}{
		"index":         0,
		"message":       message,
		"finish_reason": finishReason,
	}

	usage := map[string]interface{}{}
	if meta, ok := resp["usageMetadata"].(map[string]interface{}); ok {
		in, _ := meta["promptTokenCount"].(float64)
		out, _ := meta["candidatesTokenCount"].(float64)
		usage["prompt_tokens"] = int64(in)
		usage["completion_tokens"] = int64(out)
		usage["total_tokens"] = int64(in + out)
	}

	result := map[string]interface{}{
		"id":      "chatcmpl-gemini",
		"object":  "chat.completion",
		"model":   "",
		"choices": []interface{}{choice},
		"usage":   usage,
	}
	return json.Marshal(result)
}

// ─────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────

func extractMimeType(dataURI string) string {
	if after, ok := strings.CutPrefix(dataURI, "data:"); ok {
		if idx := strings.Index(after, ";"); idx > 0 {
			return after[:idx]
		}
	}
	return "image/jpeg"
}

func extractBase64Data(dataURI string) string {
	if idx := strings.Index(dataURI, ","); idx >= 0 {
		return dataURI[idx+1:]
	}
	return ""
}
