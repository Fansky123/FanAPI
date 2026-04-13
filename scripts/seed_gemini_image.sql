-- Gemini 图像生成渠道种子数据
-- 执行前请将 YOUR_GEMINI_KEY 替换为实际 API Key
--
-- 使用的上游接口：
--   POST https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-04-17:generateContent
--   Authorization: Bearer YOUR_GEMINI_KEY
--
-- 客户端调用示例：
--   POST /v1/image
--   {
--     "model": "gemini-2.5-flash-image",
--     "prompt": "生成一张简洁的产品宣传图，主体是一台银色笔记本电脑",
--     "aspect_ratio": "16:9"
--   }
--
-- 响应流程：
--   1. API Server 扣费 + 写 Task 记录
--   2. 发布到 NATS（task.image.<channelID>）
--   3. fanapi-script Worker 调用 Gemini API（同步等待，最长 3 分钟）
--   4. response_script 提取 base64 图片 → 返回标准格式
--   5. 客户端轮询 GET /v1/tasks/:id 拿结果
--
-- 注意：
--   - timeout_ms 设为 180000（3 分钟），Gemini 图像生成约需 1-2 分钟
--   - billing_type 使用 count（按次收费），base_price 单位为 credits
--   - 1 CNY = 1,000,000 credits；示例价格 500000 = 0.5 CNY / 次

INSERT INTO channels (
    name, model, type, base_url, method, headers, timeout_ms,
    request_script, response_script,
    billing_type, billing_config, is_active
) VALUES (
    'Gemini 2.5 Flash Image',
    'gemini-2.5-flash-image',
    'image',
    'https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-04-17:generateContent',
    'POST',
    '{"Authorization": "Bearer YOUR_GEMINI_KEY", "Content-Type": "application/json"}',
    180000,

    -- ── request_script ──────────────────────────────────────────
    -- 将平台标准格式转换为 Gemini generateContent 原生格式。
    -- 输入：{"model":"gemini-2.5-flash-image","prompt":"...","aspect_ratio":"16:9"}
    -- 输出：{"contents":[...],"generationConfig":{...}}
    'function mapRequest(input) {
    var genConfig = {
        responseModalities: ["TEXT", "IMAGE"]
    };
    if (input.aspect_ratio) {
        genConfig.aspectRatio = input.aspect_ratio;
    }
    return {
        contents: [{
            role: "user",
            parts: [{ text: input.prompt || "" }]
        }],
        generationConfig: genConfig
    };
}',


    -- ── response_script ─────────────────────────────────────────
    -- 将 Gemini generateContent 响应映射为平台标准图片格式。
    -- 输出标准格式：{"code":200,"status":2,"url":"data:image/png;base64,...","msg":""}
    'function mapResponse(input) {
    // 检查顶层 error
    if (input.error) {
        return { code: 500, status: 3, msg: input.error.message || JSON.stringify(input.error), url: "" };
    }
    var candidates = input.candidates;
    if (!candidates || candidates.length === 0) {
        return { code: 500, status: 3, msg: "no candidates in response", url: "" };
    }
    var parts = candidates[0].content && candidates[0].content.parts;
    if (!parts || parts.length === 0) {
        return { code: 500, status: 3, msg: "no parts in content", url: "" };
    }
    for (var i = 0; i < parts.length; i++) {
        var p = parts[i];
        if (p.inlineData && p.inlineData.data) {
            var mime = p.inlineData.mimeType || "image/png";
            return { code: 200, status: 2, msg: "", url: "data:" + mime + ";base64," + p.inlineData.data };
        }
    }
    return { code: 500, status: 3, msg: "no image data in response", url: "" };
}',


    -- billing_type=count：按次收费，每次固定扣除 price_per_call credits
    'count',
    '{
        "price_per_call": 500000,
        "metric_paths": {
            "count": "request.n"
        }
    }',
    true
);
