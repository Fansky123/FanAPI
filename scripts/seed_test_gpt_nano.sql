-- 测试用渠道种子数据（chatfire.cn GPT 系列 + Nano 图片生成）
-- 执行前将 YOUR_CHATFIRE_KEY 替换为实际 API Key，
-- 或执行后在管理后台 / 直接 UPDATE 语句替换。
--
-- 计费单位：credits（1 CNY ≈ 1,000,000 credits）
-- GPT-4o    input  ≈ 18 CNY/1M tokens  → 18,000,000 credits/1M tokens
--           output ≈ 72 CNY/1M tokens  → 72,000,000 credits/1M tokens
-- GPT-4o-mini input  ≈ 1.1 CNY/1M    → 1,100,000 credits/1M tokens
--             output ≈ 4.4 CNY/1M    → 4,400,000 credits/1M tokens

-- ============================================================
-- 1. LLM 渠道：gpt-4o（完全兼容 OpenAI，无需映射脚本）
-- ============================================================
INSERT INTO channels (
    name, model, type, base_url, method, headers, timeout_ms,
    request_script, response_script,
    query_url, query_method,
    billing_type, billing_config,
    key_pool_id, protocol, is_active
) VALUES (
    'ChatFire - GPT-4o',
    'gpt-4o',
    'llm',
    'https://api.*****.cn/v1/chat/completions',
    'POST',
    '{"Authorization": "Bearer YOUR_CHATFIRE_KEY", "Content-Type": "application/json"}',
    60000,
    '',
    '',
    '',
    'GET',
    'token',
    '{
        "input_price_per_1m_tokens": 18000000,
        "output_price_per_1m_tokens": 72000000,
        "input_from_response": true,
        "metric_paths": {
            "input_tokens":  "response.usage.prompt_tokens",
            "output_tokens": "response.usage.completion_tokens"
        }
    }',
    0,
    'openai',
    true
);

-- ============================================================
-- 2. LLM 渠道：gpt-4o-mini（完全兼容 OpenAI，无需映射脚本）
-- ============================================================
INSERT INTO channels (
    name, model, type, base_url, method, headers, timeout_ms,
    request_script, response_script,
    query_url, query_method,
    billing_type, billing_config,
    key_pool_id, protocol, is_active
) VALUES (
    'ChatFire - GPT-4o-mini',
    'gpt-4o-mini',
    'llm',
    'https://api.chatfire.cn/v1/chat/completions',
    'POST',
    '{"Authorization": "Bearer YOUR_CHATFIRE_KEY", "Content-Type": "application/json"}',
    60000,
    '',
    '',
    '',
    'GET',
    'token',
    '{
        "input_price_per_1m_tokens": 1100000,
        "output_price_per_1m_tokens": 4400000,
        "input_from_response": true,
        "metric_paths": {
            "input_tokens":  "response.usage.prompt_tokens",
            "output_tokens": "response.usage.completion_tokens"
        }
    }',
    0,
    'openai',
    true
);

-- ============================================================
-- 3. 图片渠道：nano-banana-pro（chatfire.cn 异步图片生成）
--
-- 平台标准入参（/v1/image）：
--   { "model": "nano-banana-pro",
--     "prompt": "...",
--     "size": "4k",          ← 可选，默认 4k
--     "refer_images": ["https://..."] }   ← 可选，参考图
--
-- chatfire 上游要求格式：
--   { "model": "nano-banana-pro_4k",
--     "prompt": "...",
--     "image": "https://..." }   ← 单个 URL 字符串
--
-- request_script 完成：
--   1. model → model + "_" + size（如 "nano-banana-pro_4k"）
--   2. refer_images[0] → image（取第一张参考图，改字段名）
--
-- response_script 完成：
--   { "data": [{ "url": "..." }] }  →  { "code": 200, "url": "...", "status": 2 }
-- ============================================================
INSERT INTO channels (
    name, model, type, base_url, method, headers, timeout_ms,
    request_script, response_script,
    query_url, query_method,
    billing_type, billing_config,
    key_pool_id, protocol, is_active
) VALUES (
    'ChatFire - Nano Banana Pro',
    'nano-banana-pro',
    'image',
    'https://api.chatfire.cn/v1/images/generations',
    'POST',
    '{"Authorization": "Bearer YOUR_CHATFIRE_KEY", "Content-Type": "application/json"}',
    120000,

    -- request_script（JavaScript）
    $js$
function mapRequest(input) {
    var out = {};
    out.prompt = input.prompt;

    // model 名称拼接档位后缀（未填 size 时默认 4k）
    var size = input.size && input.size !== '' ? input.size : '4k';
    out.model = (input.model || 'nano-banana-pro') + '_' + size;

    // refer_images[0] → image（chatfire 接受单个 URL 字符串）
    var refs = input.refer_images;
    if (refs && refs.length > 0) {
        out.image = refs[0];
    }

    return out;
}
    $js$,

    -- response_script（JavaScript）
    $js$
function mapResponse(input) {
    var out = { code: 200, status: 2, msg: '' };
    if (input.data && input.data.length > 0) {
        out.url = input.data[0].url;
    }
    return out;
}
    $js$,

    '',
    'GET',
    'image',
    '{
        "base_price": 10000,
        "metric_paths": {
            "size":         "request.size",
            "aspect_ratio": "request.aspect_ratio",
            "count":        "request.n"
        }
    }',
    0,
    'openai',
    true
);
