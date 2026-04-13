package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// APIDocs 返回内嵌的 API 文档 HTML 页面，Base URL 根据请求自动推断
func APIDocs(c *gin.Context) {
	scheme := c.GetHeader("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "http"
	}
	baseURL := scheme + "://" + c.Request.Host
	html := strings.ReplaceAll(apiDocsHTML, "http://localhost:8080", baseURL)
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
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
pre{background:#13151f;padding:14px;border-radius:0 0 6px 6px;font-size:12px;overflow-x:auto;line-height:1.6;color:#a0c4ff;margin-top:0}
.note{background:#1a2233;border-left:3px solid #63b3ed;padding:10px 14px;border-radius:0 6px 6px 0;font-size:13px;color:#a0c4ff;margin-top:8px}
.code-tabs{border:1px solid #2d3748;border-radius:6px;margin-top:4px;overflow:hidden}
.tab-bar{display:flex;background:#13151f;border-bottom:1px solid #2d3748;gap:0}
.tab{background:none;border:none;color:#718096;padding:7px 14px;font-size:12px;cursor:pointer;border-bottom:2px solid transparent;margin-bottom:-1px}
.tab:hover{color:#e2e8f0}
.tab.active{color:#63b3ed;border-bottom-color:#63b3ed}
.code-pane{border-radius:0;border:none;margin-top:0}
</style>
</head>
<body>
<div class="wrap">
  <div class="page-title">FanAPI 开放接口文档</div>
  <div class="page-sub">LLM 对话 · 图片 / 视频 / 音频生成 · 任务查询</div>
  <div class="base-url">Base URL：http://localhost:8080</div>
  <div class="auth-tip">所有接口均需在 Header 中携带 API Key：<code>X-API-Key: YOUR_SK</code><br>
  <strong>渠道路由：</strong>将请求体中的 <code>model</code> 字段设为渠道名称（即 <code>GET /user/channels</code> 返回的 <code>routing_model</code> 值），服务端自动解析为真实上游模型名。也可使用 <code>?channel_id=X</code> 查询参数（向后兼容）。</div>

  <!-- LLM -->
  <div class="section">
    <div class="section-title">1 · LLM 对话</div>

    <!-- 1.1 OpenAI -->
    <div class="endpoint">
      <div class="ep-header" onclick="toggle(this)">
        <span class="method POST">POST</span>
        <span class="ep-path">/v1/chat/completions</span>
        <span class="ep-desc">OpenAI 标准格式（兼容 OpenAI SDK）</span>
      </div>
      <div class="ep-body open">
        <h4>Request Body</h4>
        <table>
          <tr><th>字段</th><th>类型</th><th></th><th>说明</th></tr>
          <tr><td>model</td><td>string</td><td><span class="req">必填</span></td><td>渠道名称</td></tr>
          <tr><td>messages</td><td>array</td><td><span class="req">必填</span></td><td><code>[{"role":"user","content":"..."}]</code></td></tr>
          <tr><td>stream</td><td>bool</td><td><span class="opt">可选</span></td><td>true = SSE 流式，默认 false</td></tr>
          <tr><td>max_tokens</td><td>int</td><td><span class="opt">可选</span></td><td>最大输出 token 数</td></tr>
        </table>
        <h4>示例</h4>
        <div class="code-tabs">
          <div class="tab-bar">
            <button class="tab active" onclick="switchLang(this,'curl')">cURL</button>
            <button class="tab" onclick="switchLang(this,'python')">Python</button>
            <button class="tab" onclick="switchLang(this,'go')">Go</button>
            <button class="tab" onclick="switchLang(this,'java')">Java</button>
            <button class="tab" onclick="switchLang(this,'node')">Node.js</button>
          </div>
          <pre class="code-pane" data-lang="curl">curl -X POST "http://localhost:8080/v1/chat/completions" \
  -H "X-API-Key: YOUR_SK" \
  -H "Content-Type: application/json" \
  -d '{"model":"渠道名称","messages":[{"role":"user","content":"你好"}],"stream":true,"max_tokens":500}'</pre>
          <pre class="code-pane" data-lang="python" hidden>from openai import OpenAI

client = OpenAI(
    api_key="YOUR_SK",
    base_url="http://localhost:8080/v1",
)

# 将 model 设为渠道名称（routing_model），服务端自动路由
# 流式
stream = client.chat.completions.create(
    model="渠道名称",
    messages=[{"role": "user", "content": "你好"}],
    stream=True,
    max_tokens=500,
)
for chunk in stream:
    print(chunk.choices[0].delta.content or "", end="", flush=True)

# 非流式
resp = client.chat.completions.create(
    model="渠道名称",
    messages=[{"role": "user", "content": "你好"}],
    max_tokens=500,
)
print(resp.choices[0].message.content)</pre>
          <pre class="code-pane" data-lang="go" hidden>package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	body, _ := json.Marshal(map[string]any{
		"model":      "渠道名称", // routing_model from /user/channels
		"messages":   []map[string]string{{"role": "user", "content": "你好"}},
		"stream":     false,
		"max_tokens": 500,
	})
	req, _ := http.NewRequest("POST",
		"http://localhost:8080/v1/chat/completions",
		bytes.NewReader(body))
	req.Header.Set("X-API-Key", "YOUR_SK")
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	fmt.Println(string(data))
}</pre>
          <pre class="code-pane" data-lang="java" hidden>import java.net.URI;
import java.net.http.*;

public class Main {
    public static void main(String[] args) throws Exception {
        String body = """
            {"model":"gpt-4o",
             "messages":[{"role":"user","content":"你好"}],
             "stream":false,"max_tokens":500}
            """;
        HttpRequest req = HttpRequest.newBuilder()
            .uri(URI.create("http://localhost:8080/v1/chat/completions"))
            .header("X-API-Key", "YOUR_SK")
            .header("Content-Type", "application/json")
            .POST(HttpRequest.BodyPublishers.ofString(body))
            .build();
        var resp = HttpClient.newHttpClient()
            .send(req, HttpResponse.BodyHandlers.ofString());
        System.out.println(resp.body());
    }
}</pre>
          <pre class="code-pane" data-lang="node" hidden>// 使用官方 openai SDK
import OpenAI from "openai";
const client = new OpenAI({
  apiKey: "YOUR_SK",
  baseURL: "http://localhost:8080/v1",
});

// 将 model 设为渠道名称（routing_model），服务端自动路由
// 流式
const stream = await client.chat.completions.create({
  model: "渠道名称",
  messages: [{ role: "user", content: "你好" }],
  stream: true,
  max_tokens: 500,
});
for await (const chunk of stream) {
  process.stdout.write(chunk.choices[0]?.delta?.content || "");
}

// 非流式
const resp = await client.chat.completions.create({
  model: "渠道名称",
  messages: [{ role: "user", content: "你好" }],
  max_tokens: 500,
});
console.log(resp.choices[0].message.content);</pre>
        </div>
        <div class="note">流式响应为 SSE 格式，每行 <code>data: {...}</code>，最后一行 <code>data: [DONE]</code>。完全兼容 OpenAI SDK，将 <code>base_url</code> 改为 <code>http://localhost:8080/v1</code>，<code>model</code> 填渠道名称即可直接使用。也支持 <code>?channel_id=X</code> 参数（向后兼容）。</div>
      </div>
    </div>

    <!-- 1.2 Claude -->
    <div class="endpoint">
      <div class="ep-header" onclick="toggle(this)">
        <span class="method POST">POST</span>
        <span class="ep-path">/v1/messages</span>
        <span class="ep-desc">Anthropic Claude 原生格式</span>
      </div>
      <div class="ep-body">
        <h4>Request Body</h4>
        <table>
          <tr><th>字段</th><th>类型</th><th></th><th>说明</th></tr>
          <tr><td>model</td><td>string</td><td><span class="req">必填</span></td><td>渠道名称如 <code>claude-3-5-sonnet</code></td></tr>
          <tr><td>messages</td><td>array</td><td><span class="req">必填</span></td><td><code>[{"role":"user","content":"..."}]</code>，不含 system</td></tr>
          <tr><td>system</td><td>string</td><td><span class="opt">可选</span></td><td>System prompt（顶层字段）</td></tr>
          <tr><td>max_tokens</td><td>int</td><td><span class="req">必填</span></td><td>最大输出 token 数</td></tr>
          <tr><td>stream</td><td>bool</td><td><span class="opt">可选</span></td><td>true = SSE 流式</td></tr>
        </table>
        <h4>示例</h4>
        <div class="code-tabs">
          <div class="tab-bar">
            <button class="tab active" onclick="switchLang(this,'curl')">cURL</button>
            <button class="tab" onclick="switchLang(this,'python')">Python</button>
            <button class="tab" onclick="switchLang(this,'go')">Go</button>
            <button class="tab" onclick="switchLang(this,'java')">Java</button>
            <button class="tab" onclick="switchLang(this,'node')">Node.js</button>
          </div>
          <pre class="code-pane" data-lang="curl">curl -X POST "http://localhost:8080/v1/messages" \
  -H "X-API-Key: YOUR_SK" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "渠道名称",
    "system": "你是一个助手",
    "messages": [{"role": "user", "content": "你好"}],
    "max_tokens": 500,
    "stream": true
  }'</pre>
          <pre class="code-pane" data-lang="python" hidden>import anthropic

client = anthropic.Anthropic(
    api_key="YOUR_SK",
    base_url="http://localhost:8080",
    default_headers={"X-API-Key": "YOUR_SK"},
)

# 流式
with client.messages.stream(
    model="渠道名称",  # routing_model from /user/channels
    system="你是一个助手",
    messages=[{"role": "user", "content": "你好"}],
    max_tokens=500,
) as stream:
    for text in stream.text_stream:
        print(text, end="", flush=True)

# 非流式
resp = client.messages.create(
    model="渠道名称",
    system="你是一个助手",
    messages=[{"role": "user", "content": "你好"}],
    max_tokens=500,
)
print(resp.content[0].text)</pre>
          <pre class="code-pane" data-lang="go" hidden>package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	body, _ := json.Marshal(map[string]any{
		"model":      "渠道名称",
		"system":     "你是一个助手",
		"messages":   []map[string]string{{"role": "user", "content": "你好"}},
		"max_tokens": 500,
		"stream":     false,
	})
	req, _ := http.NewRequest("POST",
		"http://localhost:8080/v1/messages",
		bytes.NewReader(body))
	req.Header.Set("X-API-Key", "YOUR_SK")
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	fmt.Println(string(data))
}</pre>
          <pre class="code-pane" data-lang="java" hidden>import java.net.URI;
import java.net.http.*;

public class Main {
    public static void main(String[] args) throws Exception {
        String body = """
            {"model":"渠道名称",
             "system":"你是一个助手",
             "messages":[{"role":"user","content":"你好"}],
             "max_tokens":500,"stream":false}
            """;
        HttpRequest req = HttpRequest.newBuilder()
            .uri(URI.create("http://localhost:8080/v1/messages"))
            .header("X-API-Key", "YOUR_SK")
            .header("Content-Type", "application/json")
            .POST(HttpRequest.BodyPublishers.ofString(body))
            .build();
        var resp = HttpClient.newHttpClient()
            .send(req, HttpResponse.BodyHandlers.ofString());
        System.out.println(resp.body());
    }
}</pre>
          <pre class="code-pane" data-lang="node" hidden>import Anthropic from "@anthropic-ai/sdk";

const client = new Anthropic({
  apiKey: "unused",
  baseURL: "http://localhost:8080",
  defaultHeaders: { "X-API-Key": "YOUR_SK" },
});

// 将 model 设为渠道名称（routing_model）
// 流式
const stream = client.messages.stream({
  model: "渠道名称",
  system: "你是一个助手",
  messages: [{ role: "user", content: "你好" }],
  max_tokens: 500,
});
stream.on("text", (text) => process.stdout.write(text));
await stream.finalMessage();

// 非流式
const resp = await client.messages.create({
  model: "渠道名称",
  system: "你是一个助手",
  messages: [{ role: "user", content: "你好" }],
  max_tokens: 500,
});
console.log(resp.content[0].text);</pre>
        </div>
        <div class="note">流式响应遵循 Claude SSE 协议：<code>event: message_start</code> / <code>content_block_delta</code> / <code>message_delta</code> / <code>message_stop</code>。</div>
      </div>
    </div>

    <!-- 1.3 Gemini -->
    <div class="endpoint">
      <div class="ep-header" onclick="toggle(this)">
        <span class="method POST">POST</span>
        <span class="ep-path">/v1/gemini</span>
        <span class="ep-desc">Google Gemini 原生格式</span>
      </div>
      <div class="ep-body">
        <h4>Request Body</h4>
        <table>
          <tr><th>字段</th><th>类型</th><th></th><th>说明</th></tr>
          <tr><td>contents</td><td>array</td><td><span class="req">必填</span></td><td><code>[{"role":"user","parts":[{"text":"..."}]}]</code></td></tr>
          <tr><td>generationConfig</td><td>object</td><td><span class="opt">可选</span></td><td><code>{"maxOutputTokens":500,"temperature":0.7}</code></td></tr>
          <tr><td>systemInstruction</td><td>object</td><td><span class="opt">可选</span></td><td><code>{"parts":[{"text":"你是助手"}]}</code></td></tr>
        </table>
        <h4>示例</h4>
        <div class="code-tabs">
          <div class="tab-bar">
            <button class="tab active" onclick="switchLang(this,'curl')">cURL</button>
            <button class="tab" onclick="switchLang(this,'python')">Python</button>
            <button class="tab" onclick="switchLang(this,'go')">Go</button>
            <button class="tab" onclick="switchLang(this,'java')">Java</button>
            <button class="tab" onclick="switchLang(this,'node')">Node.js</button>
          </div>
          <pre class="code-pane" data-lang="curl">curl -X POST "http://localhost:8080/v1/gemini" \
  -H "X-API-Key: YOUR_SK" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "渠道名称",
    "contents": [{"role":"user","parts":[{"text":"你好"}]}],
    "generationConfig": {"maxOutputTokens": 500}
  }'</pre>
          <pre class="code-pane" data-lang="python" hidden>import google.generativeai as genai
import google.auth.transport.requests as ga_requests
# 将 SDK 的 base_url 指向平台
# 方法：直接用 requests 原生调用更简单

import requests

resp = requests.post(
    "http://localhost:8080/v1/gemini",
    params={"channel_id": "1"},  # 可选，向后兼容；也可在 model 字段填渠道名称
    headers={"X-API-Key": "YOUR_SK", "Content-Type": "application/json"},
    json={
        "contents": [{"role": "user", "parts": [{"text": "你好"}]}],
        "generationConfig": {"maxOutputTokens": 500},
    },
)
print(resp.json())</pre>
          <pre class="code-pane" data-lang="go" hidden>package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	body, _ := json.Marshal(map[string]any{
		"contents": []map[string]any{
			{"role": "user", "parts": []map[string]string{{"text": "你好"}}},
		},
		"generationConfig": map[string]any{"maxOutputTokens": 500},
	})
	req, _ := http.NewRequest("POST",
		"http://localhost:8080/v1/gemini",
		bytes.NewReader(body))
	req.Header.Set("X-API-Key", "YOUR_SK")
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	fmt.Println(string(data))
}</pre>
          <pre class="code-pane" data-lang="java" hidden>import java.net.URI;
import java.net.http.*;

public class Main {
    public static void main(String[] args) throws Exception {
        String body = """
            {"model":"渠道名称",
             "contents":[{"role":"user","parts":[{"text":"你好"}]}],
             "generationConfig":{"maxOutputTokens":500}}
            """;
        HttpRequest req = HttpRequest.newBuilder()
            .uri(URI.create("http://localhost:8080/v1/gemini"))
            .header("X-API-Key", "YOUR_SK")
            .header("Content-Type", "application/json")
            .POST(HttpRequest.BodyPublishers.ofString(body))
            .build();
        var resp = HttpClient.newHttpClient()
            .send(req, HttpResponse.BodyHandlers.ofString());
        System.out.println(resp.body());
    }
}</pre>
          <pre class="code-pane" data-lang="node" hidden>const resp = await fetch("http://localhost:8080/v1/gemini", {
  method: "POST",
  headers: { "X-API-Key": "YOUR_SK", "Content-Type": "application/json" },
  body: JSON.stringify({
    model: "渠道名称",
    contents: [{ role: "user", parts: [{ text: "你好" }] }],
    generationConfig: { maxOutputTokens: 500 },
  }),
});
const data = await resp.json();
console.log(data.candidates[0].content.parts[0].text);</pre>
        </div>
        <div class="note">流式响应（SSE）中每片均携带 <code>usageMetadata</code>，以最后一片为准用于计费。</div>
      </div>
    </div>
  </div>

  <!-- IMAGE -->
  <div class="section">
    <div class="section-title">2 · 图片生成</div>
    <div class="endpoint">
      <div class="ep-header" onclick="toggle(this)">
        <span class="method POST">POST</span>
        <span class="ep-path">/v1/image</span>
        <span class="ep-desc">创建图片生成任务（异步，返回 task_id）</span>
      </div>
      <div class="ep-body open">
        <h4>Request Body</h4>
        <table>
          <tr><th>字段</th><th>类型</th><th></th><th>说明</th></tr>
          <tr><td>model</td><td>string</td><td><span class="req">必填</span></td><td>渠道名称如 <code>nano-banana-pro</code></td></tr>
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
        <div class="code-tabs">
          <div class="tab-bar">
            <button class="tab active" onclick="switchLang(this,'curl')">cURL</button>
            <button class="tab" onclick="switchLang(this,'python')">Python</button>
            <button class="tab" onclick="switchLang(this,'go')">Go</button>
            <button class="tab" onclick="switchLang(this,'java')">Java</button>
            <button class="tab" onclick="switchLang(this,'node')">Node.js</button>
          </div>
          <pre class="code-pane" data-lang="curl">curl -X POST "http://localhost:8080/v1/image" \
  -H "X-API-Key: YOUR_SK" \
  -H "Content-Type: application/json" \
  -d '{"model":"渠道名称","prompt":"赛博朋克猫","size":"4k","aspect_ratio":"9:16"}'</pre>
          <pre class="code-pane" data-lang="python" hidden>import requests

resp = requests.post(
    "http://localhost:8080/v1/image",
    headers={"X-API-Key": "YOUR_SK"},
    json={
        "model": "渠道名称",
        "prompt": "赛博朋克猫",
        "size": "4k",
        "aspect_ratio": "9:16"
    }
)
task_id = resp.json()["task_id"]
print("Task ID:", task_id)</pre>
          <pre class="code-pane" data-lang="go" hidden>package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	body, _ := json.Marshal(map[string]any{
		"model":        "渠道名称",
		"prompt":       "赛博朋克猫",
		"size":         "4k",
		"aspect_ratio": "9:16",
	})
	req, _ := http.NewRequest("POST",
		"http://localhost:8080/v1/image",
		bytes.NewReader(body))
	req.Header.Set("X-API-Key", "YOUR_SK")
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	fmt.Println(string(data)) // {"task_id":1}
}</pre>
          <pre class="code-pane" data-lang="java" hidden>import java.net.URI;
import java.net.http.*;

public class Main {
    public static void main(String[] args) throws Exception {
        String body = """
            {"model":"渠道名称","prompt":"赛博朋克猫",
             "size":"4k","aspect_ratio":"9:16"}
            """;
        HttpRequest req = HttpRequest.newBuilder()
            .uri(URI.create("http://localhost:8080/v1/image"))
            .header("X-API-Key", "YOUR_SK")
            .header("Content-Type", "application/json")
            .POST(HttpRequest.BodyPublishers.ofString(body))
            .build();
        var resp = HttpClient.newHttpClient()
            .send(req, HttpResponse.BodyHandlers.ofString());
        System.out.println(resp.body()); // {"task_id":1}
    }
}</pre>
          <pre class="code-pane" data-lang="node" hidden>const resp = await fetch("http://localhost:8080/v1/image", {
  method: "POST",
  headers: { "X-API-Key": "YOUR_SK", "Content-Type": "application/json" },
  body: JSON.stringify({
    model: "渠道名称",
    prompt: "赛博朋克猫",
    size: "4k",
    aspect_ratio: "9:16"
  })
});
const { task_id } = await resp.json();
console.log("Task ID:", task_id);</pre>
        </div>
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
        <span class="ep-path">/v1/video</span>
        <span class="ep-desc">创建视频生成任务（异步，返回 task_id）</span>
      </div>
      <div class="ep-body">
        <h4>Request Body</h4>
        <table>
          <tr><th>字段</th><th>类型</th><th></th><th>说明</th></tr>
          <tr><td>model</td><td>string</td><td><span class="req">必填</span></td><td>渠道名称如 <code>video-gen-pro</code></td></tr>
          <tr><td>prompt</td><td>string</td><td><span class="req">必填</span></td><td>提示词</td></tr>
          <tr><td>size</td><td>string</td><td><span class="opt">可选</span></td><td><code>720p</code> / <code>1080p</code> / <code>2k</code> / <code>4k</code></td></tr>
          <tr><td>aspect_ratio</td><td>string</td><td><span class="opt">可选</span></td><td><code>16:9</code> / <code>9:16</code></td></tr>
          <tr><td>duration</td><td>int</td><td><span class="opt">可选</span></td><td>时长（秒）</td></tr>
          <tr><td>refer_images</td><td>[]string</td><td><span class="opt">可选</span></td><td>参考图（首帧/尾帧）</td></tr>
        </table>
        <h4>示例</h4>
        <div class="code-tabs">
          <div class="tab-bar">
            <button class="tab active" onclick="switchLang(this,'curl')">cURL</button>
            <button class="tab" onclick="switchLang(this,'python')">Python</button>
            <button class="tab" onclick="switchLang(this,'go')">Go</button>
            <button class="tab" onclick="switchLang(this,'java')">Java</button>
            <button class="tab" onclick="switchLang(this,'node')">Node.js</button>
          </div>
          <pre class="code-pane" data-lang="curl">curl -X POST "http://localhost:8080/v1/video" \
  -H "X-API-Key: YOUR_SK" \
  -H "Content-Type: application/json" \
  -d '{"model":"渠道名称","prompt":"海浪拍打礁石","size":"1080p","aspect_ratio":"16:9","duration":5}'</pre>
          <pre class="code-pane" data-lang="python" hidden>import requests

resp = requests.post(
    "http://localhost:8080/v1/video",
    headers={"X-API-Key": "YOUR_SK"},
    json={
        "model": "渠道名称",
        "prompt": "海浪拍打礁石",
        "size": "1080p",
        "aspect_ratio": "16:9",
        "duration": 5
    }
)
task_id = resp.json()["task_id"]
print("Task ID:", task_id)</pre>
          <pre class="code-pane" data-lang="go" hidden>package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	body, _ := json.Marshal(map[string]any{
		"model":        "渠道名称",
		"prompt":       "海浪拍打礁石",
		"size":         "1080p",
		"aspect_ratio": "16:9",
		"duration":     5,
	})
	req, _ := http.NewRequest("POST",
		"http://localhost:8080/v1/video",
		bytes.NewReader(body))
	req.Header.Set("X-API-Key", "YOUR_SK")
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	fmt.Println(string(data))
}</pre>
          <pre class="code-pane" data-lang="java" hidden>import java.net.URI;
import java.net.http.*;

public class Main {
    public static void main(String[] args) throws Exception {
        String body = """
            {"model":"渠道名称","prompt":"海浪拍打礁石",
             "size":"1080p","aspect_ratio":"16:9","duration":5}
            """;
        HttpRequest req = HttpRequest.newBuilder()
            .uri(URI.create("http://localhost:8080/v1/video"))
            .header("X-API-Key", "YOUR_SK")
            .header("Content-Type", "application/json")
            .POST(HttpRequest.BodyPublishers.ofString(body))
            .build();
        var resp = HttpClient.newHttpClient()
            .send(req, HttpResponse.BodyHandlers.ofString());
        System.out.println(resp.body());
    }
}</pre>
          <pre class="code-pane" data-lang="node" hidden>const resp = await fetch("http://localhost:8080/v1/video", {
  method: "POST",
  headers: { "X-API-Key": "YOUR_SK", "Content-Type": "application/json" },
  body: JSON.stringify({
    model: "渠道名称",
    prompt: "海浪拍打礁石",
    size: "1080p",
    aspect_ratio: "16:9",
    duration: 5
  })
});
const { task_id } = await resp.json();
console.log("Task ID:", task_id);</pre>
        </div>
        <h4>Response</h4>
        <pre>{"task_id": 5}</pre>
      </div>
    </div>
  </div>

  <!-- AUDIO -->
  <div class="section">
    <div class="section-title">4 · 音频生成（TTS / 语音合成）</div>
    <div class="endpoint">
      <div class="ep-header" onclick="toggle(this)">
        <span class="method POST">POST</span>
        <span class="ep-path">/v1/audio</span>
        <span class="ep-desc">创建语音合成 / TTS 任务（异步，返回 task_id）</span>
      </div>
      <div class="ep-body">
        <p class="note">此接口用于 <b>文字转语音（TTS）</b> 等音频合成场景。如需 Suno AI 音乐创作，请使用第 5 节的 <code>/v1/music</code> 接口。</p>
        <h4>Request Body</h4>
        <table>
          <tr><th>字段</th><th>类型</th><th></th><th>说明</th></tr>
          <tr><td>model</td><td>string</td><td><span class="req">必填</span></td><td>渠道名称，如 <code>tts-gen</code></td></tr>
          <tr><td>input</td><td>string</td><td><span class="req">必填</span></td><td>待合成的文本内容（TTS 场景）</td></tr>
          <tr><td>voice</td><td>string</td><td><span class="opt">可选</span></td><td>发音人 / 音色，如 <code>alloy</code></td></tr>
          <tr><td>duration</td><td>int</td><td><span class="opt">可选</span></td><td>目标时长（秒），用于计费预扣</td></tr>
        </table>
        <h4>示例</h4>
        <div class="code-tabs">
          <div class="tab-bar">
            <button class="tab active" onclick="switchLang(this,'curl')">cURL</button>
            <button class="tab" onclick="switchLang(this,'python')">Python</button>
            <button class="tab" onclick="switchLang(this,'go')">Go</button>
            <button class="tab" onclick="switchLang(this,'java')">Java</button>
            <button class="tab" onclick="switchLang(this,'node')">Node.js</button>
          </div>
          <pre class="code-pane" data-lang="curl">curl -X POST "http://localhost:8080/v1/audio" \
  -H "X-API-Key: YOUR_SK" \
  -H "Content-Type: application/json" \
  -d '{"model":"渠道名称","prompt":"一首轻快的爵士乐","style":"jazz","duration":30}'</pre>
          <pre class="code-pane" data-lang="python" hidden>import requests

resp = requests.post(
    "http://localhost:8080/v1/audio",
    headers={"X-API-Key": "YOUR_SK"},
    json={
        "model": "渠道名称",
        "prompt": "一首轻快的爵士乐",
        "style": "jazz",
        "duration": 30
    }
)
task_id = resp.json()["task_id"]
print("Task ID:", task_id)</pre>
          <pre class="code-pane" data-lang="go" hidden>package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	body, _ := json.Marshal(map[string]any{
		"model":    "渠道名称",
		"prompt":   "一首轻快的爵士乐",
		"style":    "jazz",
		"duration": 30,
	})
	req, _ := http.NewRequest("POST",
		"http://localhost:8080/v1/audio",
		bytes.NewReader(body))
	req.Header.Set("X-API-Key", "YOUR_SK")
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	fmt.Println(string(data))
}</pre>
          <pre class="code-pane" data-lang="java" hidden>import java.net.URI;
import java.net.http.*;

public class Main {
    public static void main(String[] args) throws Exception {
        String body = """
            {"model":"渠道名称","prompt":"一首轻快的爵士乐",
             "style":"jazz","duration":30}
            """;
        HttpRequest req = HttpRequest.newBuilder()
            .uri(URI.create("http://localhost:8080/v1/audio"))
            .header("X-API-Key", "YOUR_SK")
            .header("Content-Type", "application/json")
            .POST(HttpRequest.BodyPublishers.ofString(body))
            .build();
        var resp = HttpClient.newHttpClient()
            .send(req, HttpResponse.BodyHandlers.ofString());
        System.out.println(resp.body());
    }
}</pre>
          <pre class="code-pane" data-lang="node" hidden>const resp = await fetch("http://localhost:8080/v1/audio", {
  method: "POST",
  headers: { "X-API-Key": "YOUR_SK", "Content-Type": "application/json" },
  body: JSON.stringify({
    model: "渠道名称",
    prompt: "一首轻快的爵士乐",
    style: "jazz",
    duration: 30
  })
});
const { task_id } = await resp.json();
console.log("Task ID:", task_id);</pre>
        </div>
        <h4>Response</h4>
        <pre>{"task_id": 10}</pre>
      </div>
    </div>
  </div>

  <!-- MUSIC -->
  <div class="section">
    <div class="section-title">5 · 音乐生成（Suno）</div>
    <div class="endpoint">
      <div class="ep-header" onclick="toggle(this)">
        <span class="method POST">POST</span>
        <span class="ep-path">/v1/music</span>
        <span class="ep-desc">提交 Suno AI 音乐创作任务（异步，返回 task_id）</span>
      </div>
      <div class="ep-body">
        <p class="note">每次创作同时生成 <b>2 首</b>歌曲，完成后通过 <code>/v1/tasks/:id</code> 获取结果（<code>items</code> 数组）。</p>
        <h4>Request Body</h4>
        <table>
          <tr><th>字段</th><th>类型</th><th></th><th>说明</th></tr>
          <tr><td>model</td><td>string</td><td><span class="req">必填</span></td><td>渠道名称，如 <code>suno-music</code></td></tr>
          <tr><td>input_type</td><td>string</td><td><span class="req">必填</span></td><td><code>"10"</code> 灵感模式 &nbsp;|&nbsp; <code>"20"</code> 自定义/歌词模式</td></tr>
          <tr><td>mv_version</td><td>string</td><td><span class="opt">可选</span></td><td>模型版本，默认 <code>chirp-v5</code>；可选：<code>chirp-v4-5+</code> / <code>chirp-v4-5</code> / <code>chirp-v4</code> / <code>chirp-v3-5</code></td></tr>
          <tr><td>make_instrumental</td><td>bool</td><td><span class="opt">可选</span></td><td>是否纯音乐（无人声），默认 <code>false</code></td></tr>
          <tr><td>gpt_description_prompt</td><td>string</td><td><span class="opt">灵感模式必填</span></td><td>用自然语言描述想要的歌曲，Suno 自动生成歌词</td></tr>
          <tr><td>prompt</td><td>string</td><td><span class="opt">歌词模式必填</span></td><td>完整歌词内容，支持 <code>[Verse]</code>/<code>[Chorus]</code> 等结构标记</td></tr>
          <tr><td>tags</td><td>string</td><td><span class="opt">可选</span></td><td>音乐风格，如 <code>pop,female voice</code></td></tr>
          <tr><td>title</td><td>string</td><td><span class="opt">可选</span></td><td>歌曲名称</td></tr>
          <tr><td>continue_clip_id</td><td>string</td><td><span class="opt">可选</span></td><td>续写：已有歌曲的 clipId 或 MP3 URL（扩展/续写模式）</td></tr>
          <tr><td>continue_at</td><td>string</td><td><span class="opt">可选</span></td><td>续写起始时间（秒），如 <code>"70"</code> 表示从 1:10 开始</td></tr>
          <tr><td>cover_clip_id</td><td>string</td><td><span class="opt">可选</span></td><td>Cover 翻唱：参考音频 MP3 URL</td></tr>
          <tr><td>task</td><td>string</td><td><span class="opt">可选</span></td><td><code>underpainting</code> 给人声添加伴奏 &nbsp;|&nbsp; <code>overpainting</code> 给伴奏添加人声</td></tr>
          <tr><td>metadata_params</td><td>object</td><td><span class="opt">可选</span></td><td>underpainting/overpainting 附加参数：<code>underpainting_clip_id</code>/<code>overpainting_clip_id</code>、<code>_start_s</code>、<code>_end_s</code></td></tr>
          <tr><td>callback_url</td><td>string</td><td><span class="opt">可选</span></td><td>任务状态变更回调 URL（不填则不推送）</td></tr>
        </table>
        <h4>示例（灵感模式）</h4>
        <div class="code-tabs">
          <div class="tab-bar">
            <button class="tab active" onclick="switchLang(this,'curl')">cURL</button>
            <button class="tab" onclick="switchLang(this,'python')">Python</button>
            <button class="tab" onclick="switchLang(this,'node')">Node.js</button>
          </div>
          <pre class="code-pane" data-lang="curl">curl -X POST "http://localhost:8080/v1/music" \
  -H "X-API-Key: YOUR_SK" \
  -H "Content-Type: application/json" \
  -d &#39;{"model":"suno-music","input_type":"10","mv_version":"chirp-v5","make_instrumental":false,"gpt_description_prompt":"一首关于兄弟情义的歌"}&#39;</pre>
          <pre class="code-pane" data-lang="python" hidden>import requests

resp = requests.post(
    "http://localhost:8080/v1/music",
    headers={"X-API-Key": "YOUR_SK"},
    json={
        "model": "suno-music",
        "input_type": "10",
        "mv_version": "chirp-v5",
        "make_instrumental": False,
        "gpt_description_prompt": "一首关于兄弟情义的歌"
    }
)
task_id = resp.json()["task_id"]
print("Task ID:", task_id)</pre>
          <pre class="code-pane" data-lang="node" hidden>const resp = await fetch("http://localhost:8080/v1/music", {
  method: "POST",
  headers: { "X-API-Key": "YOUR_SK", "Content-Type": "application/json" },
  body: JSON.stringify({
    model: "suno-music",
    input_type: "10",
    mv_version: "chirp-v5",
    make_instrumental: false,
    gpt_description_prompt: "一首关于兄弟情义的歌"
  })
});
const { task_id } = await resp.json();
console.log("Task ID:", task_id);</pre>
        </div>
        <h4>示例（自定义歌词模式）</h4>
        <div class="code-tabs">
          <div class="tab-bar">
            <button class="tab active" onclick="switchLang(this,'curl2')">cURL</button>
            <button class="tab" onclick="switchLang(this,'node2')">Node.js</button>
          </div>
          <pre class="code-pane" data-lang="curl2">curl -X POST "http://localhost:8080/v1/music" \
  -H "X-API-Key: YOUR_SK" \
  -H "Content-Type: application/json" \
  -d &#39;{"model":"suno-music","input_type":"20","mv_version":"chirp-v5","make_instrumental":false,"prompt":"[主歌]\n星光灿烂 夜色朦胧\n勇敢前行 不问结果\n[副歌]\n在路上 勇敢追梦","tags":"folk,male voice","title":"在路上"}&#39;</pre>
          <pre class="code-pane" data-lang="node2" hidden>const resp = await fetch("http://localhost:8080/v1/music", {
  method: "POST",
  headers: { "X-API-Key": "YOUR_SK", "Content-Type": "application/json" },
  body: JSON.stringify({
    model: "suno-music",
    input_type: "20",
    mv_version: "chirp-v5",
    make_instrumental: false,
    prompt: "[主歌]\n星光灿烂 夜色朦胧\n勇敢前行 不问结果\n[副歌]\n在路上 勇敢追梦",
    tags: "folk,male voice",
    title: "在路上"
  })
});
const { task_id } = await resp.json();</pre>
        </div>
        <h4>Response（提交成功）</h4>
        <pre>{"task_id": 42}</pre>
        <h4>完成后 GET /v1/tasks/42 返回示例</h4>
        <pre>{"code":200,"status":2,"msg":"创作完成","task_id":42,"items":[
  {"id":"1874135948011253762","clip_id":"fa03ad72-b5ef-4ad4-a981-bcfa458a6e29",
   "title":"在路上","tags":"folk,male voice","duration":139,
   "audio_url":"https://cdn1.suno.ai/fa03ad72-b5ef-4ad4-a981-bcfa458a6e29.mp3",
   "image_url":"https://cdn2.suno.ai/image_b639f4e1-549b-4147-bfb8-a9564505951d.jpeg"},
  {"id":"1874135948011253763","clip_id":"71359d0d-1d92-4d88-93a0-df7bccf85727",
   "title":"在路上","tags":"folk,male voice","duration":143,
   "audio_url":"https://cdn1.suno.ai/71359d0d-1d92-4d88-93a0-df7bccf85727.mp3",
   "image_url":"https://cdn2.suno.ai/image_a6cd0b65-0b99-4e23-9476-17bdcd90d226.jpeg"}
]}</pre>
      </div>
    </div>
  </div>

  <!-- TASKS -->
  <div class="section">
    <div class="section-title">6 · 任务查询</div>
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
        <h4>示例（轮询）</h4>
        <div class="code-tabs">
          <div class="tab-bar">
            <button class="tab active" onclick="switchLang(this,'curl')">cURL</button>
            <button class="tab" onclick="switchLang(this,'python')">Python</button>
            <button class="tab" onclick="switchLang(this,'go')">Go</button>
            <button class="tab" onclick="switchLang(this,'java')">Java</button>
            <button class="tab" onclick="switchLang(this,'node')">Node.js</button>
          </div>
          <pre class="code-pane" data-lang="curl">curl "http://localhost:8080/v1/tasks/1" -H "X-API-Key: YOUR_SK"</pre>
          <pre class="code-pane" data-lang="python" hidden>import requests, time

def poll_task(task_id, api_key, timeout=300):
    url = f"http://localhost:8080/v1/tasks/{task_id}"
    headers = {"X-API-Key": api_key}
    deadline = time.time() + timeout
    interval = 2
    while time.time() < deadline:
        r = requests.get(url, headers=headers).json()
        if r["code"] == 200:
            return r["url"]          # 成功，返回结果 URL
        if r["code"] == 500:
            raise Exception(r["msg"]) # 失败
        time.sleep(interval)
        interval = min(interval + 1, 5)  # 逐步增加间隔，最大 5s
    raise TimeoutError("task timeout")

url = poll_task(1, "YOUR_SK")
print("Result:", url)</pre>
          <pre class="code-pane" data-lang="go" hidden>package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type TaskResp struct {
	Code   int    ` + "`" + `json:"code"` + "`" + `
	URL    string ` + "`" + `json:"url"` + "`" + `
	Status int    ` + "`" + `json:"status"` + "`" + `
	Msg    string ` + "`" + `json:"msg"` + "`" + `
}

func pollTask(taskID int, apiKey string) (string, error) {
	url := fmt.Sprintf("http://localhost:8080/v1/tasks/%d", taskID)
	interval := 2 * time.Second
	deadline := time.Now().Add(5 * time.Minute)
	for time.Now().Before(deadline) {
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("X-API-Key", apiKey)
		resp, _ := http.DefaultClient.Do(req)
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		var t TaskResp
		json.Unmarshal(body, &t)
		if t.Code == 200 {
			return t.URL, nil
		}
		if t.Code == 500 {
			return "", fmt.Errorf(t.Msg)
		}
		time.Sleep(interval)
		if interval < 5*time.Second {
			interval += time.Second
		}
	}
	return "", fmt.Errorf("timeout")
}

func main() {
	url, err := pollTask(1, "YOUR_SK")
	if err != nil {
		panic(err)
	}
	fmt.Println("Result:", url)
}</pre>
          <pre class="code-pane" data-lang="java" hidden>import java.net.URI;
import java.net.http.*;

public class Main {
    public static void main(String[] args) throws Exception {
        int taskId = 1;
        String apiKey = "YOUR_SK";
        HttpClient client = HttpClient.newHttpClient();
        int interval = 2000;
        long deadline = System.currentTimeMillis() + 300_000;
        while (System.currentTimeMillis() < deadline) {
            HttpRequest req = HttpRequest.newBuilder()
                .uri(URI.create("http://localhost:8080/v1/tasks/" + taskId))
                .header("X-API-Key", apiKey)
                .GET().build();
            var resp = client.send(req, HttpResponse.BodyHandlers.ofString());
            String body = resp.body();
            if (body.contains("\"code\":200")) {
                System.out.println("Done: " + body);
                return;
            }
            if (body.contains("\"code\":500")) {
                System.err.println("Failed: " + body);
                return;
            }
            Thread.sleep(interval);
            if (interval < 5000) interval += 1000;
        }
        System.err.println("Timeout");
    }
}</pre>
          <pre class="code-pane" data-lang="node" hidden>async function pollTask(taskId, apiKey, timeout = 300000) {
  const url = ` + "`" + `http://localhost:8080/v1/tasks/${taskId}` + "`" + `;
  const headers = { "X-API-Key": apiKey };
  const deadline = Date.now() + timeout;
  let interval = 2000;
  while (Date.now() < deadline) {
    const r = await fetch(url, { headers }).then(res => res.json());
    if (r.code === 200) return r.url;
    if (r.code === 500) throw new Error(r.msg);
    await new Promise(res => setTimeout(res, interval));
    if (interval < 5000) interval += 1000;
  }
  throw new Error("timeout");
}

const resultUrl = await pollTask(1, "YOUR_SK");
console.log("Result:", resultUrl);</pre>
        </div>
        <div class="note">建议轮询策略：前 10s 每 2s 一次，之后每 5s 一次，5 分钟未完成视为超时。</div>
      </div>
    </div>
  </div>

  <!-- GROUP PRICING & CHANNELS -->
  <div class="section">
    <div class="section-title">7 · 用户分组定价 &amp; 渠道列表</div>
    <div class="endpoint">
      <div class="ep-header" onclick="toggle(this)">
        <span class="method GET">GET</span>
        <span class="ep-path">/user/channels</span>
        <span style="margin-left:8px;color:#555">获取当前用户可用渠道列表（含专属定价）</span>
      </div>
      <div class="ep-body">
        <p>登录用户调用此接口时，如果其所在定价分组对某渠道设置了优惠价，响应中的 <code>group_price</code> 字段会返回该专属单价；否则省略该字段，以 <code>price</code> 为准。</p>
        <p>匿名用户请使用 <span class="ep-path">/public/channels</span>（无需认证，不含分组价）。</p>
        <table>
          <tr><th>字段</th><th>类型</th><th>说明</th></tr>
          <tr><td>id</td><td>int</td><td>渠道 ID</td></tr>
          <tr><td>name</td><td>string</td><td>渠道显示名</td></tr>
          <tr><td>routing_model</td><td>string</td><td>请求时填入 <code>model</code> 的路由标识</td></tr>
          <tr><td>price</td><td>float64</td><td>默认单价（每 1K tokens，CNY）</td></tr>
          <tr><td>group_price</td><td>float64</td><td>（可选）当前分组专属单价；存在时优先生效</td></tr>
          <tr><td>type</td><td>string</td><td>渠道类型，如 <code>openai</code> / <code>claude</code> / <code>gemini</code></td></tr>
        </table>
      </div>
    </div>
    <p class="note" style="margin-top:12px">
      <b>定价分组配置方式：</b>在渠道的 <code>billing_config</code> JSON 中添加 <code>pricing_groups</code> 对象，
      键为分组名（如 <code>vip</code>），值为该分组的覆盖价格：
      <pre>{"input_price": 0.002, "output_price": 0.008, "pricing_groups": {"vip": {"input_price": 0.001, "output_price": 0.004}}}</pre>
    </p>
  </div>

  <!-- LOAD BALANCING -->
  <div class="section">
    <div class="section-title">8 · 渠道权重 &amp; 负载均衡</div>
    <p>
      系统根据每条渠道的 <strong>priority（优先级）</strong>和 <strong>weight（权重）</strong>自动选路：
    </p>
    <ol>
      <li>优先使用 <code>priority</code> 最高的渠道组（同值为一组）。</li>
      <li>同组内按 <code>weight</code> 比例加权随机选取。</li>
      <li>若某渠道 5 分钟内请求次数 ≥ 5 且错误率 ≥ 50%，视为不健康，自动跳过（降级到下一优先级组）。</li>
      <li>最多重试 3 条渠道，全部失败则返回 503。</li>
    </ol>
    <p>渠道字段说明：</p>
    <table>
      <tr><th>字段</th><th>默认值</th><th>说明</th></tr>
      <tr><td>priority</td><td>0</td><td>数值越高越优先；相同 priority 的渠道同组竞争</td></tr>
      <tr><td>weight</td><td>1</td><td>同组内的流量权重比例，如 3:1 表示前者承接 75% 流量</td></tr>
    </table>
  </div>

  <!-- AUTH TYPES -->
  <div class="section">
    <div class="section-title">9 · 渠道认证类型（auth_type）</div>
    <p>创建/更新渠道时可指定 <code>auth_type</code>，系统据此自动构造上游鉴权头：</p>
    <table>
      <tr><th>auth_type</th><th>说明</th><th>附加字段</th></tr>
      <tr><td>bearer <em>（默认）</em></td><td>在 HTTP 头中发送 <code>Authorization: Bearer &lt;key&gt;</code></td><td>—</td></tr>
      <tr><td>query_param</td><td>将 key 附加到请求 URL 的查询参数中</td><td><code>auth_param_name</code>：参数名，如 <code>api_key</code></td></tr>
      <tr><td>basic</td><td>HTTP Basic Auth，key 格式为 <code>user:pass</code></td><td>—</td></tr>
      <tr><td>sigv4</td><td>AWS Signature Version 4（Bedrock 等服务）</td><td><code>auth_region</code>：AWS 区域；<code>auth_service</code>：服务名，如 <code>bedrock</code></td></tr>
    </table>
    <p class="note">sigv4 模式下 key 字段存储 <code>ACCESS_KEY_ID:SECRET_ACCESS_KEY</code>。</p>
  </div>

  <!-- BALANCE -->
  <div class="section">
    <div class="section-title">10 · 余额查询</div>
        <span class="ep-desc">查询当前 API Key 对应账户的剩余余额</span>
      </div>
      <div class="ep-body open">
        <h4>Response</h4>
        <table>
          <tr><th>字段</th><th>类型</th><th>说明</th></tr>
          <tr><td>balance_credits</td><td>int64</td><td>剩余 credits（1 CNY = 1,000,000 credits）</td></tr>
          <tr><td>balance_cny</td><td>float64</td><td>折合人民币金额</td></tr>
        </table>
        <pre>{"balance_credits": 5000000, "balance_cny": 5.0}</pre>
        <h4>示例</h4>
        <div class="code-tabs">
          <div class="tab-bar">
            <button class="tab active" onclick="switchLang(this,'curl')">cURL</button>
            <button class="tab" onclick="switchLang(this,'python')">Python</button>
            <button class="tab" onclick="switchLang(this,'go')">Go</button>
            <button class="tab" onclick="switchLang(this,'java')">Java</button>
            <button class="tab" onclick="switchLang(this,'node')">Node.js</button>
          </div>
          <pre class="code-pane" data-lang="curl">curl "http://localhost:8080/user/balance" -H "X-API-Key: YOUR_SK"</pre>
          <pre class="code-pane" data-lang="python" hidden>import requests

resp = requests.get(
    "http://localhost:8080/user/balance",
    headers={"X-API-Key": "YOUR_SK"},
)
data = resp.json()
print(f"余额：{data['balance_credits']} credits（{data['balance_cny']:.4f} CNY）")</pre>
          <pre class="code-pane" data-lang="go" hidden>package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	req, _ := http.NewRequest("GET", "http://localhost:8080/user/balance", nil)
	req.Header.Set("X-API-Key", "YOUR_SK")
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	fmt.Printf("余额：%.0f credits（%.4f CNY）\n",
		result["balance_credits"], result["balance_cny"])
}</pre>
          <pre class="code-pane" data-lang="java" hidden>import java.net.URI;
import java.net.http.*;

public class Main {
    public static void main(String[] args) throws Exception {
        HttpRequest req = HttpRequest.newBuilder()
            .uri(URI.create("http://localhost:8080/user/balance"))
            .header("X-API-Key", "YOUR_SK")
            .GET().build();
        var resp = HttpClient.newHttpClient()
            .send(req, HttpResponse.BodyHandlers.ofString());
        System.out.println(resp.body());
    }
}</pre>
          <pre class="code-pane" data-lang="node" hidden>const resp = await fetch("http://localhost:8080/user/balance", {
  headers: { "X-API-Key": "YOUR_SK" },
});
const { balance_credits, balance_cny } = await resp.json();
console.log(` + "`" + `余额：${balance_credits} credits（${balance_cny.toFixed(4)} CNY）` + "`" + `);</pre>
        </div>
      </div>
    </div>
  </div>
</div>

<script>
function toggle(el){
  var b=el.nextElementSibling;
  b.classList.toggle('open');
}
function switchLang(btn, lang) {
  var box = btn.closest('.code-tabs');
  box.querySelectorAll('.tab').forEach(function(t){ t.classList.remove('active'); });
  box.querySelectorAll('.code-pane').forEach(function(p){ p.hidden = true; });
  btn.classList.add('active');
  box.querySelector('.code-pane[data-lang="' + lang + '"]').hidden = false;
}
</script>
</body>
</html>`
