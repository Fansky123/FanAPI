package protocol

import "testing"

func TestResponsesToOpenAIChatCompletionsMessagesCompatible(t *testing.T) {
	req := map[string]interface{}{
		"model":             "gpt-4o",
		"stream":            true,
		"max_output_tokens": 128,
		"temperature":       0.3,
		"top_p":             0.9,
		"tool_choice":       "auto",
		"instructions":      "You are helpful.",
		"messages": []interface{}{
			map[string]interface{}{"role": "system", "content": "You are helpful."},
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{"type": "text", "text": "look"},
					map[string]interface{}{"type": "image_url", "image_url": map[string]interface{}{"url": "https://example.com/a.png"}},
				},
			},
		},
		"tools": []interface{}{
			map[string]interface{}{
				"type":        "function",
				"name":        "weather",
				"description": "Get weather",
				"parameters":  map[string]interface{}{"type": "object"},
			},
		},
	}

	out, err := responsesToOpenAI(req)
	if err != nil {
		t.Fatalf("responsesToOpenAI returned error: %v", err)
	}

	messages, ok := out["messages"].([]interface{})
	if !ok || len(messages) != 2 {
		t.Fatalf("expected 2 messages from passthrough, got %#v", out["messages"])
	}

	user, _ := messages[1].(map[string]interface{})
	parts, ok := user["content"].([]interface{})
	if !ok || len(parts) != 2 {
		t.Fatalf("expected user multimodal content preserved, got %#v", user["content"])
	}
	image, _ := parts[1].(map[string]interface{})
	if image["type"] != "image_url" {
		t.Fatalf("expected image_url part preserved, got %#v", image)
	}

	if out["max_tokens"] != 128 {
		t.Fatalf("expected max_output_tokens mapped to max_tokens, got %#v", out["max_tokens"])
	}
	if out["tool_choice"] != "auto" {
		t.Fatalf("expected tool_choice passthrough, got %#v", out["tool_choice"])
	}

	tools, ok := out["tools"].([]map[string]interface{})
	if !ok || len(tools) != 1 {
		t.Fatalf("expected converted tools, got %#v", out["tools"])
	}
	fn, _ := tools[0]["function"].(map[string]interface{})
	if fn["name"] != "weather" {
		t.Fatalf("expected nested function tool conversion, got %#v", tools[0])
	}
}

func TestResponsesToOpenAINativeInputStillWorks(t *testing.T) {
	req := map[string]interface{}{
		"model":        "gpt-4o",
		"instructions": "native system",
		"input": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{"type": "input_text", "text": "hello"},
				},
			},
		},
	}

	out, err := responsesToOpenAI(req)
	if err != nil {
		t.Fatalf("responsesToOpenAI returned error: %v", err)
	}

	messages, ok := out["messages"].([]interface{})
	if !ok || len(messages) != 2 {
		t.Fatalf("expected 2 messages from native input conversion, got %#v", out["messages"])
	}
	sys, _ := messages[0].(map[string]interface{})
	if sys["role"] != "system" || sys["content"] != "native system" {
		t.Fatalf("expected leading system message from instructions, got %#v", sys)
	}
	user, _ := messages[1].(map[string]interface{})
	if user["content"] != "hello" {
		t.Fatalf("expected input_text collapsed to string, got %#v", user["content"])
	}
}

func TestOpenAIToClaudeConvertsDataURIImageURL(t *testing.T) {
	req := map[string]interface{}{
		"model": "claude-3-5-sonnet",
		"messages": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{"type": "text", "text": "describe"},
					map[string]interface{}{"type": "image_url", "image_url": map[string]interface{}{"url": "data:image/png;base64,QUJD"}},
				},
			},
		},
	}

	out, err := openAIToClaude(req)
	if err != nil {
		t.Fatalf("openAIToClaude returned error: %v", err)
	}

	msgs, _ := out["messages"].([]map[string]interface{})
	content, _ := msgs[0]["content"].([]map[string]interface{})
	image := content[1]
	source, _ := image["source"].(map[string]interface{})
	if image["type"] != "image" || source["type"] != "base64" {
		t.Fatalf("expected image_url data URI converted to base64 image block, got %#v", image)
	}
	if source["media_type"] != "image/png" || source["data"] != "QUJD" {
		t.Fatalf("expected media/data extracted from data URI, got %#v", source)
	}
}

func TestOpenAIToClaudeConvertsRemoteImageURL(t *testing.T) {
	req := map[string]interface{}{
		"model": "claude-3-5-sonnet",
		"messages": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{"type": "image_url", "image_url": map[string]interface{}{"url": "https://example.com/image.png"}},
				},
			},
		},
	}

	out, err := openAIToClaude(req)
	if err != nil {
		t.Fatalf("openAIToClaude returned error: %v", err)
	}

	msgs, _ := out["messages"].([]map[string]interface{})
	content, _ := msgs[0]["content"].([]map[string]interface{})
	image := content[0]
	source, _ := image["source"].(map[string]interface{})
	if image["type"] != "image" || source["type"] != "url" {
		t.Fatalf("expected remote image_url converted to Claude url image block, got %#v", image)
	}
	if source["url"] != "https://example.com/image.png" {
		t.Fatalf("expected URL preserved, got %#v", source)
	}
}

func TestClaudeRequestToOpenAIPreservesImageURLParts(t *testing.T) {
	req := map[string]interface{}{
		"messages": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{"type": "image_url", "image_url": map[string]interface{}{"url": "https://example.com/a.png"}},
					map[string]interface{}{"type": "text", "text": "what is this"},
				},
			},
		},
	}

	out, err := claudeRequestToOpenAI(req)
	if err != nil {
		t.Fatalf("claudeRequestToOpenAI returned error: %v", err)
	}

	messages, _ := out["messages"].([]interface{})
	user, _ := messages[0].(map[string]interface{})
	parts, ok := user["content"].([]map[string]interface{})
	if !ok || len(parts) != 2 {
		t.Fatalf("expected rich content with image_url preserved, got %#v", user["content"])
	}
	if parts[0]["type"] != "image_url" {
		t.Fatalf("expected first part image_url to be preserved, got %#v", parts[0])
	}
}

func TestOpenAIToResponsesRequestConvertsRemoteImageURLPart(t *testing.T) {
	req := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{"type": "text", "text": "look"},
					map[string]interface{}{"type": "image_url", "image_url": map[string]interface{}{"url": "https://example.com/a.png"}},
				},
			},
		},
	}

	out, err := openAIToResponsesRequest(req)
	if err != nil {
		t.Fatalf("openAIToResponsesRequest returned error: %v", err)
	}

	input, _ := out["input"].([]interface{})
	user, _ := input[0].(map[string]interface{})
	content, _ := user["content"].([]interface{})
	image, _ := content[1].(map[string]interface{})
	if image["type"] != "input_image" || image["image_url"] != "https://example.com/a.png" {
		t.Fatalf("expected image_url part converted to input_image string url, got %#v", image)
	}
}

func TestOpenAIToResponsesRequestConvertsDataURIImageURLPart(t *testing.T) {
	req := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{"type": "image_url", "image_url": map[string]interface{}{"url": "data:image/png;base64,QUJD"}},
				},
			},
		},
	}

	out, err := openAIToResponsesRequest(req)
	if err != nil {
		t.Fatalf("openAIToResponsesRequest returned error: %v", err)
	}

	input, _ := out["input"].([]interface{})
	user, _ := input[0].(map[string]interface{})
	content, _ := user["content"].([]interface{})
	image, _ := content[0].(map[string]interface{})
	if image["type"] != "input_image" || image["image_url"] != "data:image/png;base64,QUJD" {
		t.Fatalf("expected data URI image_url converted to input_image string url, got %#v", image)
	}
}

func TestOpenAIToResponsesRequestPreservesInputImagePart(t *testing.T) {
	req := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{"type": "input_image", "image_url": "https://example.com/already-native.png"},
				},
			},
		},
	}

	out, err := openAIToResponsesRequest(req)
	if err != nil {
		t.Fatalf("openAIToResponsesRequest returned error: %v", err)
	}

	input, _ := out["input"].([]interface{})
	user, _ := input[0].(map[string]interface{})
	content, _ := user["content"].([]interface{})
	image, _ := content[0].(map[string]interface{})
	if image["type"] != "input_image" || image["image_url"] != "https://example.com/already-native.png" {
		t.Fatalf("expected input_image part preserved, got %#v", image)
	}
}
