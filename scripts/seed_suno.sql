-- Suno AI 音乐生成渠道种子数据
-- 执行前将 YOUR_SUNO_BASE_URL 替换为实际上游 API 域名（如 https://api.example.com）
--         YOUR_SUNO_KEY 替换为实际 API Key
-- 1 CNY = 1,000,000 credits；Suno 每次创作固定扣 1 次（同时生成 2 首）

-- ============================================================
-- 说明：所有创作模式（灵感/歌词/续写/cover/加伴奏/加人声）共用
--       同一接口 POST /_open/suno/music/generate，由 input_type
--       和附加字段控制行为，因此只需建一个渠道，通过 request_script
--       把我们平台的统一格式转换为 Suno 所需格式即可。
--
-- 轮询：提交成功后上游返回 data.taskBatchId，作为轮询 ID；
--       每 5 秒调用 GET /_open/suno/music/getState?taskBatchId={id}
--       直到 data.taskStatus == "finished"。
--
-- 任务结果：完成后每次生成 2 首歌曲，挂在 TaskResult.items 数组里，
--           每首包含 audio_url、image_url、duration、title、tags。
-- ============================================================

INSERT INTO channels (
    name, model, type, base_url, method, headers, timeout_ms,
    request_script, response_script,
    query_url, query_method, query_timeout_ms, query_script,
    billing_type, billing_config,
    key_pool_id, protocol, is_active
) VALUES (
    'Suno V5 音乐生成',
    'suno-music',
    'music',
    'YOUR_SUNO_BASE_URL/_open/suno/music/generate',
    'POST',
    '{"Authorization": "Bearer YOUR_SUNO_KEY", "Content-Type": "application/json"}',
    120000,

    -- =========================================================
    -- request_script：平台统一格式 → Suno API 格式
    --
    -- 平台字段            Suno 字段
    -- ---------------    -----------------
    -- input_type         inputType
    -- mv_version         mvVersion         (默认 chirp-v5)
    -- make_instrumental  makeInstrumental
    -- gpt_description_prompt  gptDescriptionPrompt  (灵感模式)
    -- prompt             prompt            (歌词模式)
    -- tags               tags
    -- title              title
    -- continue_clip_id   continueClipId    (续写)
    -- continue_at        continueAt
    -- cover_clip_id      coverClipId       (cover)
    -- task               task              (underpainting/overpainting)
    -- metadata_params    metadataParams
    -- callback_url       callbackUrl
    -- =========================================================
    $request_script$
function mapRequest(input) {
    var b = {
        mvVersion:         input.mv_version || 'chirp-v5',
        inputType:         input.input_type  || '10',
        makeInstrumental:  input.make_instrumental !== undefined ? input.make_instrumental : false,
        callbackUrl:       input.callback_url || ''
    };

    if (b.inputType === '10') {
        // 灵感模式：只需 gptDescriptionPrompt
        b.gptDescriptionPrompt = input.gpt_description_prompt || '';
    } else {
        // 自定义/歌词/续写/cover/underpainting/overpainting 模式
        b.prompt = input.prompt || '';
        b.tags   = input.tags   || '';
        b.title  = input.title  || '';

        if (input.continue_clip_id) { b.continueClipId = input.continue_clip_id; }
        if (input.continue_at)      { b.continueAt     = input.continue_at; }
        if (input.cover_clip_id)    { b.coverClipId    = input.cover_clip_id; }

        // 添加伴奏 / 添加人声特殊任务类型
        if (input.task) {
            b.task           = input.task;
            b.metadataParams = input.metadata_params || {};
        }
    }

    return b;
}
    $request_script$,

    -- =========================================================
    -- response_script：解析提交创作后的响应
    -- 上游成功响应示例：
    -- {
    --   "code": 200,
    --   "msg": "您已成功提交创作任务...",
    --   "data": {
    --     "taskBatchId": "1067069768483733505",
    --     "items": [{"id": "...", "status": 0, "progressMsg": "排队中..."}]
    --   }
    -- }
    -- 返回 upstream_task_id → 触发轮询流程
    -- =========================================================
    $response_script$
function mapResponse(output) {
    if (!output || output.code !== 200) {
        var errMsg = (output && output.msg) ? output.msg : '提交任务失败';
        return { status: 3, msg: errMsg };
    }
    var taskBatchId = output.data && output.data.taskBatchId;
    if (!taskBatchId) {
        return { status: 3, msg: '上游未返回 taskBatchId' };
    }
    return {
        status:           1,
        upstream_task_id: String(taskBatchId),
        msg:              '生成中'
    };
}
    $response_script$,

    -- 轮询 URL：{id} 会被替换为 upstream_task_id（即 taskBatchId）
    'YOUR_SUNO_BASE_URL/_open/suno/music/getState?taskBatchId={id}',
    'GET',
    30000,

    -- =========================================================
    -- query_script：解析轮询状态接口的响应
    -- data.taskStatus:
    --   "create"     → 刚提交
    --   "processing" → 生成中
    --   "finished"   → 全部完成（含成功和失败皆有可能）
    --
    -- data.items[].status:
    --   10 = 排队等待
    --   20 = 正在执行
    --   30 = 成功
    --   40 = 失败
    --
    -- 返回 status=2 表示成功，items 含所有已成功的歌曲；
    -- 返回 status=1 表示仍在进行；
    -- 返回 status=3 表示失败。
    -- =========================================================
    $query_script$
function mapResponse(output) {
    if (!output || output.code !== 200) {
        var errMsg = (output && output.msg) ? output.msg : '查询失败';
        return { status: 3, msg: errMsg };
    }

    var data       = output.data || {};
    var taskStatus = data.taskStatus || '';
    var items      = data.items || [];

    // taskStatus 不为 finished 时继续等待
    if (taskStatus !== 'finished') {
        // 计算整体进度（所有 item 的平均值）
        var totalProgress = 0;
        for (var i = 0; i < items.length; i++) {
            totalProgress += (items[i].progress || 0);
        }
        var avgProgress = items.length > 0 ? Math.round(totalProgress / items.length) : 0;
        return { status: 1, msg: '生成中', progress: avgProgress };
    }

    // finished：收集所有成功的歌曲
    var successItems = [];
    for (var j = 0; j < items.length; j++) {
        var it = items[j];
        if (it.status === 30) {
            successItems.push({
                id:        it.id        || '',
                clip_id:   it.clipId    || '',
                title:     it.title     || '',
                tags:      it.tags      || '',
                prompt:    it.prompt    || '',
                duration:  it.duration  || 0,
                audio_url: it.cld2AudioUrl  || '',
                image_url: it.cld2ImageUrl  || '',
                progress_msg: it.progressMsg || ''
            });
        }
    }

    if (successItems.length === 0) {
        // 全部失败
        var failMsg = (items[0] && items[0].progressMsg) ? items[0].progressMsg : '创作失败';
        return { status: 3, code: 500, msg: failMsg };
    }

    return {
        status:  2,
        code:    200,
        msg:     '创作完成',
        items:   successItems
    };
}
    $query_script$,

    -- =========================================================
    -- 计费配置：count 类型，按次计费
    -- 每次调用固定扣 base_price credits，无论生成几首歌。
    -- 根据实际成本修改 base_price（默认 5000000 = 5 CNY）
    -- =========================================================
    'count',
    '{"base_price": 5000000}',

    0,       -- key_pool_id：不使用号池，直接用 headers 中的静态 key
    'openai', -- protocol（不影响 music 任务，保持默认）
    true
);
