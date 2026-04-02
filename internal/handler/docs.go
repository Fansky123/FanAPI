package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIDocs 返回内嵌的 API 文档 HTML 页面
func APIDocs(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, apiDocsHTML)
}

const apiDocsHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>FanAPI 接口文档</title>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;background:#0f1117;color:#e2e8f0;line-height:1.6;padding:40px 0}
.wrap{max-width:860px;margin:0 auto;padding:0 24px}
.page-title{font-size:26px;font-weight:700;color:#fff;margin-bottom:6px}
.page-sub{color:#718096;margin-bottom:24px;font-size:14px}
.base-url{background:#1a1d27;border:1px solid #2d3748;border-radius:8px;padding:10px 16px;font-family:monospace;color:#68d391;margin-bottom:32px;font-size:13px}
.auth-tip{background:#1a2233;border-left:3px solid #63b3ed;padding:10px 14px;border-radius:0 6px 6px 0;font-size:13px;color:#a0c4ff;margin-bottom:32px}
.section{margin-bottom:40px}
.section-title{font-size:13px;font-weight:600;color:#718096;text-transform:uppercase;letter-spacing:.08em;margin-bottom:10px}
.endpoint{background:#1a1d27;border:1px solid #2d3748;border-radius:8px;margin-bottom:10px;overflow:hidden}
.ep-header{display:flex;align-items:center;padding:13px 16px;cursor:pointer;gap:12px;user-select:none}
.ep-header:hover{background:#202336}
.method{padding:3px 10px;border-radius:4px;font-size:11px;font-weight:700;font-family:monospace;min-width:50px;text-align:center}
.GET{background:#1a3a2a;color:#68d391;border:1px solid #276749}
.POST{background:#1a2a3a;color:#63b3ed;border:1px solid #2b6cb0}
.ep-path{font-family:monospace;font-size:14px;color:#e2e8f0;flex:1}
.ep-desc{font-size:13px;color:#718096}
.ep-body{padding:16px;border-top:1px solid #2d3748;display:none}
.ep-body.open{display:block}
.ep-body h4{font-size:11px;color:#718096;text-transform:uppercase;letter-spacing:.08em;margin-bottom:8px;margin-top:14px}
.ep-body h4:first-child{margin-top:0}
table{width:100%;border-collapse:collapse;font-size:13px}
th{text-align:left;padding:7px 10px;background:#13151f;color:#718096;font-weight:500}
td{padding:7px 10px;border-top:1px solid #1e2232;vertical-align:top}
.req{color:#f6ad55;font-size:11px}.opt{color:#4a5568;font-size:11px}
code{background:#13151f;padding:2px 6px;border-radius:3px;font-family:monospace;font-size:12px;color:#f6e05e}
pre{background:#13151f;padding:14px;border-radius:6px;font-size:12px;overflow-x:auto;line-height:1.6;color:#a0c4ff;margin-top:4px}
.note{background:#1a2233;border-left:3px solid #63b3ed;padding:10px 14px;border-radius:0 6px 6px 0;font-size:13px;color:#a0c4ff;margin-top:8px}
</style>
</head>
<body>
<div class="wrap">
  <div class="page-title">FanAPI 开放接口文档</div>
  <div class="page-sub">LLM 对话 · 图片 / 视频 / 音频生成 · 任务查询</div>
  <div class="base-url">Base URL：http://localhost:8080</div>
  <div class="auth-tip">所有接口均需在 Header 中携带 API Key：<code>X-API-Key: YOUR_SK</code><br>
  调用接口时通过 Query 参数 <code>channel_id</code> 指定渠道，渠道列表见登录后 <code>GET /user/channels</code>。</div>

  <!-- LLM -->
  <div class="section">
    <div class="section-title">1 · LLM 对话</div>
    <div class="endpoint">
      <div class="ep-header" onclick="toggle(this)">
        <span class="method POST">POST</span>
        <span class="ep-path">/v1/llm?channel_id={id}</span>
        <span class="ep-desc">流式 / 非流式对话（OpenAI 兼容）</span>
      </div>
      <div class="ep-body open">
        <h4>Request Body</h4>
        <table>
          <tr><th>字段</th><th>类型</th><th></th><th>说明</th></tr>
          <tr><td>model</td><td>string</td><td><span class="req">必填</span></td><td>模型名，如 <code>claude-3-5-sonnet-20241022</code></td></tr>
          <tr><td>messages</td><td>array</td><td><span class="req">必填</span></td><td><code>[{"role":"user","content":"..."}]</code></td></tr>
          <tr><td>stream</td><td>bool</td><td><span class="opt">可选</span></td><td>true = SSE 流式，默认 false</td></tr>
          <tr><td>max_tokens</td><td>int</td><td><span class="opt">可选</span></td><td>最大输出 token 数</td></tr>
        </table>
        <h4>示例</h4>
        <pre>curl -X POST "http://localhost:8080/v1/llm?channel_id=1" \
  -H "X-API-Key: YOUR_SK" \
  -H "Content-Type: application/json" \
  -d '{"model":"claude-3-5-sonnet-20241022","messages":[{"role":"user","content":"你好"}],"stream":true,"max_tokens":500}'</pre>
        <div class="note">流式响应为 SSE 格式，每行 <code>data: {...}</code>，最后一行 <code>data: [DONE]</code></div>
      </div>
    </div>
  </div>

  <!-- IMAGE -->
  <div class="section">
    <div class="section-title">2 · 图片生成</div>
    <div class="endpoint">
      <div class="ep-header" onclick="toggle(this)">
        <span class="method POST">POST</span>
        <span class="ep-path">/v1/image?channel_id={id}</span>
        <span class="ep-desc">创建图片生成任务（异步，返回 task_id）</span>
      </div>
      <div class="ep-body open">
        <h4>Request Body</h4>
        <table>
          <tr><th>字段</th><th>类型</th><th></th><th>说明</th></tr>
          <tr><td>model</td><td>string</td><td><span class="req">必填</span></td><td>模型，如 <code>nano-banana-pro</code></td></tr>
          <tr><td>prompt</td><td>string</td><td><span class="req">必填</span></td><td>正向提示词</td></tr>
          <tr><td>negative_prompt</td><td>string</td><td><span class="opt">可选</span></td><td>反向提示词</td></tr>
          <tr><td>size</td><td>string</td><td><span class="opt">可选</span></td><td>分辨率档位：<code>1k</code> / <code>2k</code> / <code>3k</code> / <code>4k</code>，默认 2k</td></tr>
          <tr><td>aspect_ratio</td><td>string</td><td><span class="opt">可选</span></td><td>宽高比：<code>16:9</code> / <code>9:16</code> / <code>1:1</code> / <code>4:3</code>，默认 1:1</td></tr>
          <tr><td>refer_images</td><td>[]string</td><td><span class="opt">可选</span></td><td>参考图 URL 列表</td></tr>
          <tr><td>n</td><td>int</td><td><span class="opt">可选</span></td><td>生成数量，默认 1</td></tr>
        </table>
        <h4>size × aspect_ratio 像素对照</h4>
        <table>
          <tr><th>size \ ratio</th><th>16:9</th><th>9:16</th><th>1:1</th><th>4:3</th></tr>
          <tr><td>1k</td><td>1024×576</td><td>576×1024</td><td>1024×1024</td><td>1024×768</td></tr>
          <tr><td>2k</td><td>2048×1152</td><td>1152×2048</td><td>2048×2048</td><td>2048×1536</td></tr>
          <tr><td>3k</td><td>3072×1728</td><td>1728×3072</td><td>3072×3072</td><td>3072×2304</td></tr>
          <tr><td>4k</td><td>4096×2304</td><td>2304×4096</td><td>4096×4096</td><td>4096×3072</td></tr>
        </table>
        <h4>示例</h4>
        <pre>curl -X POST "http://localhost:8080/v1/image?channel_id=2" \
  -H "X-API-Key: YOUR_SK" \
  -H "Content-Type: application/json" \
  -d '{"model":"nano-banana-pro","prompt":"赛博朋克猫","size":"4k","aspect_ratio":"9:16"}'</pre>
        <h4>Response</h4>
        <pre>{"task_id": 1}</pre>
      </div>
    </div>
  </div>

  <!-- VIDEO -->
  <div class="section">
    <div class="section-title">3 · 视频生成</div>
    <div class="endpoint">
      <div class="ep-header" onclick="toggle(this)">
        <span class="method POST">POST</span>
        <span class="ep-path">/v1/video?channel_id={id}</span>
        <span class="ep-desc">创建视频生成任务（异步，返回 task_id）</span>
      </div>
      <div class="ep-body">
        <h4>Request Body</h4>
        <table>
          <tr><th>字段</th><th>类型</th><th></th><th>说明</th></tr>
          <tr><td>model</td><td>string</td><td><span class="req">必填</span></td><td>视频模型名称</td></tr>
          <tr><td>prompt</td><td>string</td><td><span class="req">必填</span></td><td>提示词</td></tr>
          <tr><td>size</td><td>string</td><td><span class="opt">可选</span></td><td><code>720p</code> / <code>1080p</code> / <code>2k</code> / <code>4k</code></td></tr>
          <tr><td>aspect_ratio</td><td>string</td><td><span class="opt">可选</span></td><td><code>16:9</code> / <code>9:16</code></td></tr>
          <tr><td>duration</td><td>int</td><td><span class="opt">可选</span></td><td>时长（秒）</td></tr>
          <tr><td>refer_images</td><td>[]string</td><td><span class="opt">可选</span></td><td>参考图（首帧/尾帧）</td></tr>
        </table>
        <h4>Response</h4>
        <pre>{"task_id": 5}</pre>
      </div>
    </div>
  </div>

  <!-- AUDIO -->
  <div class="section">
    <div class="section-title">4 · 音频生成</div>
    <div class="endpoint">
      <div class="ep-header" onclick="toggle(this)">
        <span class="method POST">POST</span>
        <span class="ep-path">/v1/audio?channel_id={id}</span>
        <span class="ep-desc">创建音频生成任务（异步，返回 task_id）</span>
      </div>
      <div class="ep-body">
        <h4>Request Body</h4>
        <table>
          <tr><th>字段</th><th>类型</th><th></th><th>说明</th></tr>
          <tr><td>model</td><td>string</td><td><span class="req">必填</span></td><td>模型名称</td></tr>
          <tr><td>prompt</td><td>string</td><td><span class="req">必填</span></td><td>歌词 / 描述</td></tr>
          <tr><td>style</td><td>string</td><td><span class="opt">可选</span></td><td>风格描述</td></tr>
          <tr><td>duration</td><td>int</td><td><span class="opt">可选</span></td><td>时长（秒）</td></tr>
        </table>
        <h4>Response</h4>
        <pre>{"task_id": 10}</pre>
      </div>
    </div>
  </div>

  <!-- TASKS -->
  <div class="section">
    <div class="section-title">5 · 任务查询</div>
    <div class="endpoint">
      <div class="ep-header" onclick="toggle(this)">
        <span class="method GET">GET</span>
        <span class="ep-path">/v1/tasks/:id</span>
        <span class="ep-desc">轮询图片 / 视频 / 音频任务结果</span>
      </div>
      <div class="ep-body open">
        <h4>Response</h4>
        <table>
          <tr><th>字段</th><th>类型</th><th>说明</th></tr>
          <tr><td>code</td><td>int</td><td><code>150</code> 进行中 &nbsp;|&nbsp; <code>200</code> 成功 &nbsp;|&nbsp; <code>500</code> 失败</td></tr>
          <tr><td>status</td><td>int</td><td>0 排队 · 1 生成中 · 2 成功 · 3 失败</td></tr>
          <tr><td>url</td><td>string</td><td>结果文件 URL（成功时）</td></tr>
          <tr><td>msg</td><td>string</td><td>错误描述（失败时）</td></tr>
        </table>
        <pre>// 进行中
{"code":150,"url":"","status":1,"msg":""}

// 成功
{"code":200,"url":"https://cdn.example.com/output.png","status":2,"msg":""}

// 失败
{"code":500,"url":"","status":3,"msg":"upstream error"}</pre>
        <h4>示例（每 3 秒轮询）</h4>
        <pre>curl "http://localhost:8080/v1/tasks/1" -H "X-API-Key: YOUR_SK"</pre>
        <div class="note">建议轮询策略：前 10s 每 2s 一次，之后每 5s 一次，5 分钟未完成视为超时。</div>
      </div>
    </div>
  </div>
</div>

<script>
function toggle(el){
  var b=el.nextElementSibling;
  b.classList.toggle('open');
}
</script>
</body>
</html>`
