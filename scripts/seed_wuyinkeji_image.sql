-- 无音科技 (wuyinkeji) 图片生成渠道种子数据
-- 接口文档：https://api.wuyinkeji.com/api/async/image_nanoBanana2
-- 执行前将以下占位符替换为实际值：
--   YOUR_WUYINKEJI_KEY  → 无音科技 API Key（同时作为请求体 key 和轮询 key）
--
-- 提交接口：POST https://api.wuyinkeji.com/api/async/image_nanoBanana2
-- 请求体：
--   { "key": "xxx", "prompt": "...", "size": "...", "aspectRatio": "16:9", "urls": ["..."] }
-- 响应：
--   { "code": 200, "data": { "id": "image_xxx", "count": "1" } }
--
-- 轮询接口：GET https://api.wuyinkeji.com/api/async/detail?key=xxx&id=image_xxx
--   返回 data.status 字段判断任务状态，data.url 或 data.urls 存放结果图片。

INSERT INTO channels (
    name, model, type, base_url, method, headers, timeout_ms,
    request_script, response_script,
    query_url, query_method, query_timeout_ms, query_script,
    billing_type, billing_config,
    is_active
) VALUES (
    '无音科技 - nanoBanana2 图片生成',
    'nano-banana2',
    'image',
    'https://api.*****.com/api/async/image_nanoBanana2',
    'POST',
    '{"Content-Type": "application/json"}',
    120000,

    -- =========================================================
    -- request_script：平台标准格式 → 无音科技 API 格式
    --
    -- 平台字段          无音科技字段
    -- ------------      -----------
    -- prompt            prompt
    -- size              size
    -- aspect_ratio      aspectRatio（冒号保留，如 "16:9"）
    -- refer_images[]    urls（图片数组）
    -- （注入）          key（从 script 硬编码注入）
    -- =========================================================
    $request_script$
function mapRequest(input) {
    var out = {
        key:    'YOUR_WUYINKEJI_KEY',
        prompt: input.prompt || ''
    };

    if (input.size)         { out.size        = input.size; }
    if (input.aspect_ratio) { out.aspectRatio = input.aspect_ratio; }
    if (input.refer_images && input.refer_images.length > 0) {
        out.urls = input.refer_images;
    }

    return out;
}
    $request_script$,

    -- =========================================================
    -- response_script：解析提交响应，提取 upstream_task_id
    -- 成功响应：{ "code": 200, "data": { "id": "image_xxx", "count": "1" } }
    -- =========================================================
    $response_script$
function mapResponse(output) {
    if (!output || output.code !== 200) {
        var errMsg = (output && output.msg) ? output.msg : '提交任务失败';
        return { status: 3, msg: errMsg };
    }
    var taskId = output.data && output.data.id;
    if (!taskId) {
        return { status: 3, msg: '上游未返回任务 id' };
    }
    return {
        status:           1,
        upstream_task_id: String(taskId),
        msg:              '生成中'
    };
}
    $response_script$,

    -- =========================================================
    -- 轮询配置：GET /api/async/detail?key=xxx&id={id}
    -- {id} 会自动替换为 upstream_task_id
    -- =========================================================
    'https://api.wuyinkeji.com/api/async/detail?key=YOUR_WUYINKEJI_KEY&id={id}',
    'GET',
    30000,

    -- =========================================================
    -- query_script：解析轮询状态响应
    --
    -- 响应结构（参考）：
    -- {
    --   "code": 200,
    --   "data": {
    --     "status": "processing" | "success" | "failed",
    --     "url": "https://...",        // 单图
    --     "urls": ["https://..."],     // 多图（优先）
    --     "msg": "..."
    --   }
    -- }
    -- =========================================================
    $query_script$
function mapResponse(output) {
    if (!output || output.code !== 200) {
        var errMsg = (output && output.msg) ? output.msg : '查询失败';
        return { status: 3, msg: errMsg };
    }

    var data = output.data || {};
    var st   = data.status;   // 数字：1=处理中，2=成功，3=失败

    if (st === 3) {
        return { status: 3, msg: data.message || '生成失败' };
    }

    if (st !== 2) {
        return { status: 1, msg: '生成中' };
    }

    // status=2 成功，图片在 data.result[]
    var urls = data.result || [];
    if (urls.length === 0) {
        return { status: 3, msg: '上游未返回图片地址' };
    }

    return {
        status: 2,
        code:   200,
        msg:    '生成完成',
        url:    urls[0],
        urls:   urls
    };
}
    $query_script$,

    -- =========================================================
    -- 计费配置：image 类型按次计费
    -- base_price 单位 credits，默认 5000000 = 5 CNY，按实际成本调整
    -- =========================================================
    'image',
    '{
        "base_price": 5000000,
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
