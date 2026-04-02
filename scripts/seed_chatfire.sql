-- chatfire.cn 渠道种子数据
-- 执行前请将 YOUR_CHATFIRE_KEY 替换为实际 API Key
-- 1 CNY = 1,000,000 credits；价格单位：credits / 1k tokens 或 credits / 次

-- ============================================================
-- 1. LLM 渠道：claude-3-5-sonnet-20241022
--    接口格式完全兼容 OpenAI，无需 request_script 映射
--    input_from_response=true：输入 token 数从响应 usage.prompt_tokens 取
-- ============================================================
INSERT INTO channels (
    name, model, type, base_url, method, headers, timeout_ms,
    request_script, response_script,
    billing_type, billing_config, is_active
) VALUES (
    'ChatFire - Claude 3.5 Sonnet',
    'claude-3-5-sonnet-20241022',
    'llm',
    'https://api.chatfire.cn/v1/chat/completions',
    'POST',
    '{"Authorization": "Bearer YOUR_CHATFIRE_KEY", "Content-Type": "application/json"}',
    60000,
    '',  -- 入参符合 OpenAI 标准，无需映射
    '',  -- 响应符合 OpenAI 标准，无需映射
    'token',
    '{
        "input_price_per_1k_tokens": 15000,
        "output_price_per_1k_tokens": 60000,
        "input_from_response": true,
        "metric_paths": {
            "input_tokens":  "response.usage.prompt_tokens",
            "output_tokens": "response.usage.completion_tokens",
            "max_tokens":    "request.max_tokens"
        }
    }',
    true
);

-- ============================================================
-- 2. 图片渠道：nano-banana-pro（chatfire 版模型名带 _4k 后缀）
--
-- 我们的标准入参：
--   { "model": "nano-banana-pro", "prompt": "...",
--     "refer_images": [...], "size": "4k", "aspect_ratio": "9:16" }
--
-- chatfire 要求的格式：
--   { "model": "nano-banana-pro_4k", "prompt": "...",
--     "image": [...], "size": "9x16" }
--
-- request_script 负责做如下转换：
--   1. model 拼接 size 档位后缀（_1k/_2k/_3k/_4k）
--   2. refer_images → image（字段改名）
--   3. size + aspect_ratio → chatfire 的 "WxH" 格式（如 "9x16"）
-- ============================================================
INSERT INTO channels (
    name, model, type, base_url, method, headers, timeout_ms,
    request_script, response_script,
    billing_type, billing_config, is_active
) VALUES (
    'ChatFire - Nano Banana Pro',
    'nano-banana-pro',
    'image',
    'https://api.chatfire.cn/v1/images/generations',
    'POST',
    '{"Authorization": "Bearer YOUR_CHATFIRE_KEY", "Content-Type": "application/json"}',
    120000,

    -- request_script：将平台标准格式转换为 chatfire nano 所需的格式
    -- 规则：
    --   model        → model + "_" + size（如 "nano-banana-pro_4k"）
    --   refer_images → image（chatfire 字段名）
    --   size         → 删除（已融入 model 名）
    --   aspect_ratio → size（chatfire 用 "WxH" 宽高比格式，如 "9x16"）
    'package main

import "strings"

func MapRequest(input map[string]interface{}) map[string]interface{} {
    out := make(map[string]interface{})

    // 透传 prompt
    if v, ok := input["prompt"]; ok {
        out["prompt"] = v
    }

    // model 名称拼接 size 档位后缀（如 "nano-banana-pro" + "_4k"）
    modelName, _ := input["model"].(string)
    size, _ := input["size"].(string)
    if size != "" {
        out["model"] = modelName + "_" + size
    } else {
        out["model"] = modelName
    }

    // refer_images → image（chatfire 使用 image 字段名）
    if imgs, ok := input["refer_images"]; ok {
        out["image"] = imgs
    }

    // aspect_ratio "9:16" → chatfire size "9x16"（冒号换成 x）
    if ar, ok := input["aspect_ratio"].(string); ok && ar != "" {
        out["size"] = strings.ReplaceAll(ar, ":", "x")
    }

    return out
}',

    -- response_script：将 chatfire nano 原始响应映射为平台标准格式
    -- 输入示例：{"data":[{"url":"https://v3.fal.media/files/...output.png"}],"created":1756272317}
    -- 输出格式：{"code":200,"url":"https://...","status":2,"msg":""}
    'package main

func MapResponse(input map[string]interface{}) map[string]interface{} {
    out := map[string]interface{}{
        "code":   200,
        "status": 2,
        "msg":    "",
    }

    // 从 data[0].url 提取生成图片 URL
    if data, ok := input["data"]; ok {
        if arr, ok := data.([]interface{}); ok && len(arr) > 0 {
            if item, ok := arr[0].(map[string]interface{}); ok {
                if u, ok := item["url"].(string); ok {
                    out["url"] = u
                }
            }
        }
    }

    return out
}',

    'image',
    '{
        "base_price": 10000,
        "resolution_tiers": [
            {"max_pixels": 1048576,  "multiplier": 1.0},
            {"max_pixels": 4194304,  "multiplier": 2.0},
            {"max_pixels": 9437184,  "multiplier": 3.0},
            {"max_pixels": 16777216, "multiplier": 4.0}
        ],
        "metric_paths": {
            "size":         "request.size",
            "aspect_ratio": "request.aspect_ratio",
            "count":        "request.n"
        }
    }',
    true
);
