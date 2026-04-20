-- Gemini LLM 官方接口渠道种子数据
-- 执行前请将 YOUR_GEMINI_KEY 替换为实际 API Key（可从 https://aistudio.google.com/apikey 获取）
--
-- 使用的上游接口（根据是否流式自动切换）：
--   非流式：POST https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent
--   流  式：POST https://generativelanguage.googleapis.com/v1beta/models/{model}:streamGenerateContent?alt=sse
--
-- URL 中的 {model} 由渠道 model 字段自动填充，{stream_action} 由系统根据请求自动切换。
--
-- 客户端调用示例（OpenAI 格式）：
--   POST /v1/chat/completions
--   Authorization: Bearer <your-api-key>
--   {
--     "model": "gemini-2.5-flash",   -- 即下面的 routing_model
--     "messages": [{"role": "user", "content": "你好"}],
--     "stream": true
--   }
--
-- 计费说明（token 计费）：
--   billing_config 中的价格单位为 credits/token
--   1 CNY = 1,000,000 credits
--   Gemini 2.5 Flash 参考定价（2025年）：
--     input : $0.15 / 1M tokens = 约 1.08 CNY / 1M tokens
--     output: $0.60 / 1M tokens = 约 4.32 CNY / 1M tokens
--   下面示例使用 input=2, output=8（credits/token）仅供参考，请根据实际汇率调整
--
-- 如需多个模型（如 gemini-2.0-flash、gemini-1.5-pro），复制此 INSERT 修改 name/model 即可

-- protocol=gemini：系统自动将 OpenAI 格式请求转换为 Gemini generateContent 格式
-- billing_type=token：按 prompt/completion token 计费
INSERT INTO channels (
    name, model, type, base_url, method, headers, timeout_ms,
    protocol, billing_type, billing_config, icon_url, description, is_active
) VALUES (
    'Gemini 2.5 Flash',
    'gemini-2.5-flash',
    'llm',
    'https://generativelanguage.googleapis.com/v1beta/models/{model}:{stream_action}',
    'POST',
    '{"Authorization": "Bearer YOUR_GEMINI_KEY"}',
    60000,
    'gemini',
    'token',
    '{"input_price": 2, "output_price": 8}',
    '',
    '',
    true
);

-- ──────────────────────────────────────────────────────────────────────
-- 如需添加其他 Gemini 模型，参考以下模板（取消注释后修改）：
-- ──────────────────────────────────────────────────────────────────────

-- Gemini 2.0 Flash（更快，价格更低）
-- INSERT INTO channels (
--     name, model, type, base_url, method, headers, timeout_ms,
--     protocol, billing_type, billing_config, icon_url, description, is_active
-- ) VALUES (
--     'Gemini 2.0 Flash',
--     'gemini-2.0-flash',
--     'llm',
--     'https://generativelanguage.googleapis.com/v1beta/models/{model}:{stream_action}',
--     'POST',
--     '{"Authorization": "Bearer YOUR_GEMINI_KEY"}',
--     60000,
--     'gemini',
--     'token',
--     '{"input_price": 1, "output_price": 4}',
--     '',
--     '',
--     true
-- );

-- Gemini 1.5 Pro（长上下文）
-- INSERT INTO channels (
--     name, model, type, base_url, method, headers, timeout_ms,
--     protocol, billing_type, billing_config, icon_url, description, is_active
-- ) VALUES (
--     'Gemini 1.5 Pro',
--     'gemini-1.5-pro',
--     'llm',
--     'https://generativelanguage.googleapis.com/v1beta/models/{model}:{stream_action}',
--     'POST',
--     '{"Authorization": "Bearer YOUR_GEMINI_KEY"}',
--     120000,
--     'gemini',
--     'token',
--     '{"input_price": 9, "output_price": 36}',
--     '',
--     '',
--     true
-- );
