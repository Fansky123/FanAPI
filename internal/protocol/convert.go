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
	ProtocolOpenAI    = "openai"
	ProtocolClaude    = "claude"
	ProtocolGemini    = "gemini"
	ProtocolResponses = "responses" // OpenAI Responses API（Codex CLI 使用）
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
			cacheCreate, _ := usg["cache_creation_input_tokens"].(float64)
			cacheRead, _ := usg["cache_read_input_tokens"].(float64)
			result := map[string]interface{}{
				"prompt_tokens":     int64(in),
				"completion_tokens": int64(out),
				"total_tokens":      int64(in + out),
			}
			if cacheCreate > 0 {
				result["cache_creation_tokens"] = int64(cacheCreate)
			}
			if cacheRead > 0 {
				result["cache_read_tokens"] = int64(cacheRead)
			}
			return result
		}
	case ProtocolGemini:
		if meta, ok := resp["usageMetadata"].(map[string]interface{}); ok {
			in, _ := meta["promptTokenCount"].(float64)
			out, _ := meta["candidatesTokenCount"].(float64)
			cacheRead, _ := meta["cachedContentTokenCount"].(float64)
			result := map[string]interface{}{
				"prompt_tokens":     int64(in),
				"completion_tokens": int64(out),
				"total_tokens":      int64(in + out),
			}
			if cacheRead > 0 {
				result["cache_read_tokens"] = int64(cacheRead)
			}
			return result
		}
	default:
		if usg, ok := resp["usage"].(map[string]interface{}); ok {
			pt, _ := usg["prompt_tokens"].(float64)
			ct, _ := usg["completion_tokens"].(float64)
			result := map[string]interface{}{
				"prompt_tokens":     int64(pt),
				"completion_tokens": int64(ct),
				"total_tokens":      int64(pt + ct),
			}
			// OpenAI prompt caching: prompt_tokens_details.cached_tokens
			if details, ok := usg["prompt_tokens_details"].(map[string]interface{}); ok {
				if n, _ := details["cached_tokens"].(float64); n > 0 {
					result["cache_read_tokens"] = int64(n)
				}
			}
			return result
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
			// Claude 没有 none 等价项，用 auto 并不强制调用工具
			return map[string]interface{}{"type": "auto"}
		case "required":
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

// ─────────────────────────────────────────────
// Client Request Normalization (Native → OpenAI)
// ─────────────────────────────────────────────

// NormalizeClientRequest converts a client's native-format request to OpenAI format.
// Used when clients send Claude or Gemini native format so the conversion pipeline
// always operates on a canonical OpenAI intermediate representation.
// Returns the same map unchanged when clientProto == "openai".
func NormalizeClientRequest(req map[string]interface{}, clientProto string) (map[string]interface{}, error) {
	switch clientProto {
	case ProtocolClaude:
		return claudeRequestToOpenAI(req)
	case ProtocolGemini:
		return geminiRequestToOpenAI(req)
	case ProtocolResponses:
		return responsesToOpenAI(req)
	default:
		return req, nil
	}
}

// ConvertResponseToClient converts an OpenAI-format sync response to the client's native format.
// Used after the upstream response has been normalised to OpenAI via ConvertSyncResponse.
// Returns the same bytes unchanged when clientProto == "openai".
func ConvertResponseToClient(respBytes []byte, clientProto string) ([]byte, error) {
	switch clientProto {
	case ProtocolClaude:
		return openAIToClaudeResponse(respBytes)
	case ProtocolGemini:
		return openAIToGeminiResponse(respBytes)
	case ProtocolResponses:
		return openAIToResponsesSync(respBytes)
	default:
		return respBytes, nil
	}
}

// claudeRequestToOpenAI converts a Claude Messages API request body to OpenAI format.
func claudeRequestToOpenAI(req map[string]interface{}) (map[string]interface{}, error) {
	out := make(map[string]interface{})

	if m, ok := req["model"].(string); ok {
		out["model"] = m
	}
	if mt, ok := req["max_tokens"]; ok {
		out["max_tokens"] = mt
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

	var messages []interface{}

	// Claude top-level system field → OpenAI system message
	if sys, ok := req["system"].(string); ok && sys != "" {
		messages = append(messages, map[string]interface{}{
			"role":    "system",
			"content": sys,
		})
	}

	if msgs, ok := req["messages"].([]interface{}); ok {
		for _, m := range msgs {
			msg, ok := m.(map[string]interface{})
			if !ok {
				continue
			}
			role, _ := msg["role"].(string)

			switch c := msg["content"].(type) {
			case string:
				messages = append(messages, map[string]interface{}{
					"role":    role,
					"content": c,
				})
			case []interface{}:
				// Claude content blocks → OpenAI content
				var textParts []string
				var richParts []map[string]interface{}
				hasRich := false
				toolResultHandled := false

				for _, block := range c {
					bm, ok := block.(map[string]interface{})
					if !ok {
						continue
					}
					switch bm["type"] {
					case "text":
						text, _ := bm["text"].(string)
						textParts = append(textParts, text)
						richParts = append(richParts, map[string]interface{}{"type": "text", "text": text})

					case "image":
						hasRich = true
						if source, ok := bm["source"].(map[string]interface{}); ok {
							switch source["type"] {
							case "base64":
								mime, _ := source["media_type"].(string)
								data, _ := source["data"].(string)
								richParts = append(richParts, map[string]interface{}{
									"type": "image_url",
									"image_url": map[string]interface{}{
										"url": "data:" + mime + ";base64," + data,
									},
								})
							case "url":
								url, _ := source["url"].(string)
								richParts = append(richParts, map[string]interface{}{
									"type":      "image_url",
									"image_url": map[string]interface{}{"url": url},
								})
							}
						}

					case "tool_result":
						// Each tool_result block becomes a separate tool message in OpenAI
						toolResultHandled = true
						toolUseID, _ := bm["tool_use_id"].(string)
						var content string
						switch rc := bm["content"].(type) {
						case string:
							content = rc
						case []interface{}:
							for _, rb := range rc {
								if rbm, ok := rb.(map[string]interface{}); ok {
									if t, _ := rbm["text"].(string); t != "" {
										content += t
									}
								}
							}
						}
						messages = append(messages, map[string]interface{}{
							"role":         "tool",
							"tool_call_id": toolUseID,
							"content":      content,
						})

					case "tool_use":
						// tool_use blocks in assistant messages → tool_calls array
						hasRich = true
						tcID, _ := bm["id"].(string)
						tcName, _ := bm["name"].(string)
						argsBytes, _ := json.Marshal(bm["input"])
						richParts = append(richParts, map[string]interface{}{
							"_tool_use": map[string]interface{}{
								"id":        tcID,
								"name":      tcName,
								"arguments": string(argsBytes),
							},
						})
					}
				}

				if toolResultHandled {
					continue // already appended tool messages
				}

				// Extract tool_use entries from richParts
				var toolCalls []map[string]interface{}
				var cleanParts []map[string]interface{}
				for _, rp := range richParts {
					if tu, ok := rp["_tool_use"].(map[string]interface{}); ok {
						toolCalls = append(toolCalls, map[string]interface{}{
							"id":   tu["id"],
							"type": "function",
							"function": map[string]interface{}{
								"name":      tu["name"],
								"arguments": tu["arguments"],
							},
						})
					} else {
						cleanParts = append(cleanParts, rp)
					}
				}

				outMsg := map[string]interface{}{"role": role}
				if len(toolCalls) > 0 {
					outMsg["content"] = nil
					outMsg["tool_calls"] = toolCalls
				} else if hasRich {
					outMsg["content"] = cleanParts
				} else {
					outMsg["content"] = strings.Join(textParts, "")
				}
				messages = append(messages, outMsg)

			default:
				messages = append(messages, map[string]interface{}{
					"role":    role,
					"content": c,
				})
			}
		}
	}

	out["messages"] = messages

	// tools: Claude format → OpenAI format
	if tools, ok := req["tools"].([]interface{}); ok && len(tools) > 0 {
		out["tools"] = convertClaudeToolsToOpenAI(tools)
	}

	return out, nil
}

func convertClaudeToolsToOpenAI(tools []interface{}) []map[string]interface{} {
	var out []map[string]interface{}
	for _, t := range tools {
		tm, ok := t.(map[string]interface{})
		if !ok {
			continue
		}
		fn := map[string]interface{}{"name": tm["name"]}
		if desc, ok := tm["description"].(string); ok {
			fn["description"] = desc
		}
		if schema, ok := tm["input_schema"]; ok {
			fn["parameters"] = schema
		}
		out = append(out, map[string]interface{}{
			"type":     "function",
			"function": fn,
		})
	}
	return out
}

// geminiRequestToOpenAI converts a Gemini generateContent request body to OpenAI format.
func geminiRequestToOpenAI(req map[string]interface{}) (map[string]interface{}, error) {
	out := make(map[string]interface{})

	if m, ok := req["model"].(string); ok {
		out["model"] = m
	}
	if s, ok := req["stream"]; ok {
		out["stream"] = s
	}

	var messages []interface{}

	// systemInstruction
	if si, ok := req["systemInstruction"].(map[string]interface{}); ok {
		if parts, ok := si["parts"].([]interface{}); ok {
			var sysText string
			for _, p := range parts {
				if pm, ok := p.(map[string]interface{}); ok {
					if t, ok := pm["text"].(string); ok {
						sysText += t
					}
				}
			}
			if sysText != "" {
				messages = append(messages, map[string]interface{}{
					"role":    "system",
					"content": sysText,
				})
			}
		}
	}

	// contents
	if contents, ok := req["contents"].([]interface{}); ok {
		for _, c := range contents {
			cm, ok := c.(map[string]interface{})
			if !ok {
				continue
			}
			role, _ := cm["role"].(string)
			if role == "model" {
				role = "assistant"
			}

			parts, _ := cm["parts"].([]interface{})
			var text string
			var richParts []map[string]interface{}
			hasRich := false

			for _, p := range parts {
				pm, ok := p.(map[string]interface{})
				if !ok {
					continue
				}
				if t, ok := pm["text"].(string); ok {
					text += t
					richParts = append(richParts, map[string]interface{}{"type": "text", "text": t})
				} else if id, ok := pm["inlineData"].(map[string]interface{}); ok {
					hasRich = true
					mime, _ := id["mimeType"].(string)
					data, _ := id["data"].(string)
					richParts = append(richParts, map[string]interface{}{
						"type":      "image_url",
						"image_url": map[string]interface{}{"url": "data:" + mime + ";base64," + data},
					})
				} else if fd, ok := pm["fileData"].(map[string]interface{}); ok {
					hasRich = true
					uri, _ := fd["fileUri"].(string)
					richParts = append(richParts, map[string]interface{}{
						"type":      "image_url",
						"image_url": map[string]interface{}{"url": uri},
					})
				}
			}

			if hasRich {
				messages = append(messages, map[string]interface{}{"role": role, "content": richParts})
			} else {
				messages = append(messages, map[string]interface{}{"role": role, "content": text})
			}
		}
	}

	out["messages"] = messages

	// generationConfig
	if gc, ok := req["generationConfig"].(map[string]interface{}); ok {
		if mt, ok := gc["maxOutputTokens"]; ok {
			out["max_tokens"] = mt
		}
		if t, ok := gc["temperature"]; ok {
			out["temperature"] = t
		}
		if tp, ok := gc["topP"]; ok {
			out["top_p"] = tp
		}
	}

	return out, nil
}

// ─────────────────────────────────────────────
// Response Denormalization (OpenAI → Client Native)
// ─────────────────────────────────────────────

// openAIToClaudeResponse converts an OpenAI sync response to Claude Messages API format.
func openAIToClaudeResponse(body []byte) ([]byte, error) {
	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return body, nil
	}

	id, _ := resp["id"].(string)
	model, _ := resp["model"].(string)

	var content []map[string]interface{}
	stopReason := "end_turn"

	if choices, ok := resp["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if msg, ok := choice["message"].(map[string]interface{}); ok {
				switch c := msg["content"].(type) {
				case string:
					if c != "" {
						content = append(content, map[string]interface{}{"type": "text", "text": c})
					}
				case []interface{}:
					for _, block := range c {
						if bm, ok := block.(map[string]interface{}); ok {
							content = append(content, bm)
						}
					}
				}
				if toolCalls, ok := msg["tool_calls"].([]interface{}); ok {
					for _, tc := range toolCalls {
						if tcm, ok := tc.(map[string]interface{}); ok {
							tcID, _ := tcm["id"].(string)
							fn, _ := tcm["function"].(map[string]interface{})
							name, _ := fn["name"].(string)
							argsStr, _ := fn["arguments"].(string)
							var input interface{}
							_ = json.Unmarshal([]byte(argsStr), &input)
							content = append(content, map[string]interface{}{
								"type":  "tool_use",
								"id":    tcID,
								"name":  name,
								"input": input,
							})
							stopReason = "tool_use"
						}
					}
				}
			}
			if fr, ok := choice["finish_reason"].(string); ok {
				switch fr {
				case "length":
					stopReason = "max_tokens"
				case "tool_calls":
					stopReason = "tool_use"
				}
			}
		}
	}

	usage := map[string]interface{}{"input_tokens": int64(0), "output_tokens": int64(0)}
	if usg, ok := resp["usage"].(map[string]interface{}); ok {
		if pt, ok := usg["prompt_tokens"].(float64); ok {
			usage["input_tokens"] = int64(pt)
		}
		if ct, ok := usg["completion_tokens"].(float64); ok {
			usage["output_tokens"] = int64(ct)
		}
	}

	out := map[string]interface{}{
		"id":            id,
		"type":          "message",
		"role":          "assistant",
		"model":         model,
		"content":       content,
		"stop_reason":   stopReason,
		"stop_sequence": nil,
		"usage":         usage,
	}
	return json.Marshal(out)
}

// openAIToGeminiResponse converts an OpenAI sync response to Gemini generateContent format.
func openAIToGeminiResponse(body []byte) ([]byte, error) {
	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return body, nil
	}

	var content string
	finishReason := "STOP"

	if choices, ok := resp["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if msg, ok := choice["message"].(map[string]interface{}); ok {
				content, _ = msg["content"].(string)
			}
			if fr, ok := choice["finish_reason"].(string); ok && fr == "length" {
				finishReason = "MAX_TOKENS"
			}
		}
	}

	usageMeta := map[string]interface{}{
		"promptTokenCount":     int64(0),
		"candidatesTokenCount": int64(0),
		"totalTokenCount":      int64(0),
	}
	if usg, ok := resp["usage"].(map[string]interface{}); ok {
		pt, _ := usg["prompt_tokens"].(float64)
		ct, _ := usg["completion_tokens"].(float64)
		usageMeta["promptTokenCount"] = int64(pt)
		usageMeta["candidatesTokenCount"] = int64(ct)
		usageMeta["totalTokenCount"] = int64(pt + ct)
	}

	out := map[string]interface{}{
		"candidates": []interface{}{
			map[string]interface{}{
				"content": map[string]interface{}{
					"parts": []interface{}{map[string]interface{}{"text": content}},
					"role":  "model",
				},
				"finishReason": finishReason,
				"index":        0,
			},
		},
		"usageMetadata": usageMeta,
	}
	return json.Marshal(out)
}

// ─────────────────────────────────────────────
// Responses API (OpenAI Responses API / Codex CLI)
// ─────────────────────────────────────────────

// responsesToOpenAI converts an OpenAI Responses API request to OpenAI chat/completions format.
// Responses API fields: model, input (string | array), instructions, stream, max_output_tokens, tools
func responsesToOpenAI(req map[string]interface{}) (map[string]interface{}, error) {
	out := make(map[string]interface{})

	if m, ok := req["model"].(string); ok {
		out["model"] = m
	}
	if s, ok := req["stream"]; ok {
		out["stream"] = s
	}
	if mt, ok := req["max_output_tokens"]; ok {
		out["max_tokens"] = mt
	}
	if t, ok := req["temperature"]; ok {
		out["temperature"] = t
	}
	if tp, ok := req["top_p"]; ok {
		out["top_p"] = tp
	}

	var messages []interface{}

	// instructions → system message
	if inst, ok := req["instructions"].(string); ok && inst != "" {
		messages = append(messages, map[string]interface{}{
			"role":    "system",
			"content": inst,
		})
	}

	// input: string | array of content items
	switch inp := req["input"].(type) {
	case string:
		messages = append(messages, map[string]interface{}{
			"role":    "user",
			"content": inp,
		})
	case []interface{}:
		for _, item := range inp {
			im, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			role, _ := im["role"].(string)
			if role == "" {
				role = "user"
			}
			switch c := im["content"].(type) {
			case string:
				messages = append(messages, map[string]interface{}{
					"role":    role,
					"content": c,
				})
			case []interface{}:
				// content parts: {type: "input_text"|"output_text", text: "..."}
				var parts []map[string]interface{}
				var simpleText string
				allText := true
				for _, cp := range c {
					cpm, ok := cp.(map[string]interface{})
					if !ok {
						continue
					}
					text, _ := cpm["text"].(string)
					t, _ := cpm["type"].(string)
					if t == "input_text" || t == "output_text" || t == "text" {
						simpleText += text
						parts = append(parts, map[string]interface{}{"type": "text", "text": text})
					} else {
						allText = false
						parts = append(parts, cpm)
					}
				}
				if allText {
					messages = append(messages, map[string]interface{}{
						"role":    role,
						"content": simpleText,
					})
				} else {
					messages = append(messages, map[string]interface{}{
						"role":    role,
						"content": parts,
					})
				}
			}
		}
	}

	out["messages"] = messages

	// tools passthrough (already OpenAI format in Responses API)
	if tools, ok := req["tools"].([]interface{}); ok && len(tools) > 0 {
		out["tools"] = tools
	}

	return out, nil
}

// openAIToResponsesSync converts an OpenAI chat.completion response to Responses API format.
func openAIToResponsesSync(body []byte) ([]byte, error) {
	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return body, nil
	}

	id, _ := resp["id"].(string)
	model, _ := resp["model"].(string)

	var text string
	if choices, ok := resp["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if msg, ok := choice["message"].(map[string]interface{}); ok {
				text, _ = msg["content"].(string)
			}
		}
	}

	inputTokens := int64(0)
	outputTokens := int64(0)
	if usg, ok := resp["usage"].(map[string]interface{}); ok {
		if pt, ok := usg["prompt_tokens"].(float64); ok {
			inputTokens = int64(pt)
		}
		if ct, ok := usg["completion_tokens"].(float64); ok {
			outputTokens = int64(ct)
		}
	}

	out := map[string]interface{}{
		"id":         id,
		"object":     "response",
		"created_at": resp["created"],
		"model":      model,
		"status":     "completed",
		"output": []interface{}{
			map[string]interface{}{
				"type":   "message",
				"id":     id,
				"status": "completed",
				"role":   "assistant",
				"content": []interface{}{
					map[string]interface{}{
						"type": "output_text",
						"text": text,
					},
				},
			},
		},
		"usage": map[string]interface{}{
			"input_tokens":  inputTokens,
			"output_tokens": outputTokens,
		},
	}
	return json.Marshal(out)
}
