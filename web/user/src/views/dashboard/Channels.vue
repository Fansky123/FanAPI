<template>
  <div class="models-page">
    <!-- 页面标题 -->
    <div class="page-title">模型列表</div>

    <!-- 过滤器：类别 -->
    <div class="filter-row">
      <span class="filter-label">类别：</span>
      <div class="filter-groups">
        <a
          v-for="t in typeOptions"
          :key="t.value"
          class="filter-tag"
          :class="{ active: filterType === t.value }"
          @click="filterType = t.value"
        >{{ t.label }}</a>
      </div>
    </div>

    <!-- 过滤器：供应商 -->
    <div class="filter-row" v-if="vendors.length > 0">
      <span class="filter-label">供应商：</span>
      <div class="filter-groups">
        <a
          class="filter-tag"
          :class="{ active: filterVendor === '' }"
          @click="filterVendor = ''"
        >全部</a>
        <a
          v-for="v in vendors"
          :key="v"
          class="filter-tag"
          :class="{ active: filterVendor === v }"
          @click="filterVendor = v"
        >{{ v }}</a>
      </div>
    </div>

    <!-- 搜索框 -->
    <div class="search-row">
      <el-input
        v-model="filterName"
        placeholder="搜索模型名称..."
        clearable
        style="width: 260px"
        size="default"
      >
        <template #prefix><el-icon><Search /></el-icon></template>
      </el-input>
      <span class="count-label">共 <b>{{ filteredChannels.length }}</b> 个模型</span>
    </div>

    <!-- 加载 -->
    <div class="grid-loading" v-if="loading">
      <el-skeleton :rows="3" animated v-for="i in 6" :key="i" />
    </div>

    <!-- 模型卡片网格 -->
    <div class="model-grid" v-else>
      <div
        class="model-card"
        v-for="ch in filteredChannels"
        :key="ch.id"
        @click="openDoc(ch)"
      >
        <!-- 顶部：图标 + 名称 + 复制 -->
        <div class="card-header">
          <img
            v-if="ch.icon_url"
            :src="ch.icon_url"
            class="card-icon"
            alt=""
            @error="(e) => e.target.style.display='none'"
          />
          <div class="card-icon card-icon-fallback" v-else>
            {{ (ch.name || '?').charAt(0).toUpperCase() }}
          </div>
          <div class="card-meta">
            <div class="card-name-row">
              <span class="card-name">{{ ch.name }}</span>
              <el-icon class="copy-icon" @click.stop="copyModel(ch.name)" title="复制模型名"><CopyDocument /></el-icon>
            </div>
            <div class="card-cost" v-if="ch.price_display">
              <div v-for="(line, i) in ch.price_display.split('\n')" :key="i">
                <span v-if="i === 0">积分消耗：{{ line }}</span>
                <span v-else style="color:#909399;font-size:11px">{{ line }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 描述（最多2行） -->
        <div class="card-desc" v-if="ch.description">{{ ch.description }}</div>

        <!-- 底部：类型标签 + 状态 -->
        <div class="card-footer">
          <span class="card-type-tag" :class="'type-' + (ch.type || 'llm')">{{ typeLabel(ch.type) }}</span>
          <span class="card-status">
            <span class="status-dot"></span>可用
          </span>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <el-empty v-if="!loading && filteredChannels.length === 0" description="暂无可用模型" style="padding: 60px 0" />

    <!-- 模型文档抽屉（右侧滑出） -->
    <el-drawer
      v-model="docVisible"
      :title="docMode === 'balance' ? 'API：查询账户余额' : docMode === 'task' ? 'API：查询任务结果' : docChannel?.name"
      direction="rtl"
      size="600px"
      destroy-on-close
      @closed="docMode = 'channel'; langTab = 'curl'; sunoMode = 'inspire'"
    >
      <!-- 余额接口文档 -->
      <template v-if="docMode === 'balance'">
        <div class="doc-info-grid">
          <div class="doc-info-item">
            <span class="doc-label">方法</span>
            <span class="method-badge" style="background:#00b42a">GET</span>
          </div>
          <div class="doc-info-item">
            <span class="doc-label">认证</span>
            <el-tag size="small" type="info">Bearer Token</el-tag>
          </div>
        </div>

        <div class="doc-section-title">接口地址</div>
        <div class="doc-endpoint-row">
          <span class="method-badge" style="background:#00b42a">GET</span>
          <code class="doc-endpoint-text">/user/balance</code>
          <el-icon class="doc-copy-icon" @click.stop="copyText(`GET ${window.location.origin}/user/balance`)"><CopyDocument /></el-icon>
        </div>

        <div class="doc-section-title">说明</div>
        <div style="font-size:13px;color:#606266;line-height:1.8;margin-bottom:4px">
          返回当前 API Key 对应账户的余额。<br>
          <code style="background:#f7f8fa;padding:1px 5px;border-radius:3px">balance_credits</code> 为内部精度值（÷ 1,000,000 = 积分数）；
          <code style="background:#f7f8fa;padding:1px 5px;border-radius:3px">balance_cny</code> 为等值积分数。
        </div>

        <div class="doc-section-title">请求头</div>
        <pre class="doc-pre">Authorization: Bearer YOUR_API_KEY</pre>

        <div class="doc-section-title" style="display:flex;align-items:center;justify-content:space-between">
          <span>代码示例</span>
          <el-button size="small" plain type="primary" @click.stop="copyText(balanceCode(langTab))">
            <el-icon><CopyDocument /></el-icon> 复制
          </el-button>
        </div>
        <el-tabs v-model="langTab" size="small" style="margin:0 0 4px">
          <el-tab-pane label="cURL" name="curl" />
          <el-tab-pane label="Python" name="python" />
          <el-tab-pane label="PHP" name="php" />
          <el-tab-pane label="Go" name="go" />
          <el-tab-pane label="Java" name="java" />
        </el-tabs>
        <pre class="doc-pre doc-pre-scroll">{{ balanceCode(langTab) }}</pre>

        <div class="doc-section-title">响应示例</div>
        <pre class="doc-pre">{{ balanceResponse() }}</pre>
      </template>

      <!-- 任务结果查询文档 -->
      <template v-else-if="docMode === 'task'">
        <div class="doc-section-title">接口地址</div>
        <div class="doc-endpoint-row">
          <span class="method-badge" style="background:#00b42a">GET</span>
          <code class="doc-endpoint-text">/v1/tasks/{id}</code>
          <el-icon class="doc-copy-icon" @click.stop="copyText(`${window.location.origin}/v1/tasks/YOUR_TASK_ID`)"><CopyDocument /></el-icon>
        </div>

        <div class="doc-section-title">说明</div>
        <div style="font-size:13px;color:#606266;line-height:1.8;margin-bottom:4px">
          轮询图片 / 视频 / 音频 / 音乐任务的执行结果。<br>
          <code style="background:#f7f8fa;padding:1px 5px;border-radius:3px">code=150</code> 进行中，
          <code style="background:#f7f8fa;padding:1px 5px;border-radius:3px">code=200</code> 成功，
          <code style="background:#f7f8fa;padding:1px 5px;border-radius:3px">code=500</code> 失败。<br>
          建议间隔 2～5 秒轮询，成功后读取 <code style="background:#f7f8fa;padding:1px 5px;border-radius:3px">url</code>（单结果）或 <code style="background:#f7f8fa;padding:1px 5px;border-radius:3px">items</code>（多结果，如音乐）。
        </div>

        <div class="doc-section-title">请求头</div>
        <pre class="doc-pre">Authorization: Bearer YOUR_API_KEY</pre>

        <div class="doc-section-title" style="display:flex;align-items:center;justify-content:space-between">
          <span>代码示例</span>
          <el-button size="small" plain type="primary" @click.stop="copyText(taskCode(langTab))">
            <el-icon><CopyDocument /></el-icon> 复制
          </el-button>
        </div>
        <el-tabs v-model="langTab" size="small" style="margin:0 0 4px">
          <el-tab-pane label="cURL" name="curl" />
          <el-tab-pane label="Python" name="python" />
          <el-tab-pane label="PHP" name="php" />
          <el-tab-pane label="Go" name="go" />
          <el-tab-pane label="Java" name="java" />
        </el-tabs>
        <pre class="doc-pre doc-pre-scroll">{{ taskCode(langTab) }}</pre>

        <div class="doc-section-title">响应示例</div>
        <pre class="doc-pre">{{ taskResponse() }}</pre>
      </template>

      <!-- 渠道文档 -->
      <template v-else-if="docChannel">
        <!-- 基本信息 + 图标 + 余额按钮 -->
        <div style="display:flex;align-items:flex-start;gap:12px;padding:12px 0 16px;border-bottom:1px solid #f0f1f5;margin-bottom:4px">
          <div style="flex:1;display:flex;flex-wrap:wrap;gap:12px 32px">
            <div class="doc-info-item">
              <span class="doc-label">类型</span>
              <span class="card-type-tag" :class="'type-' + docChannel.type">{{ typeLabel(docChannel.type) }}</span>
            </div>
            <div class="doc-info-item">
              <span class="doc-label">计费</span>
              <span style="color:#165dff;font-weight:600">
                <span v-for="(line, i) in (docChannel.price_display || '—').split('\n')" :key="i">
                  <br v-if="i > 0" />{{ line }}
                </span>
              </span>
              <span v-if="docChannel.group_price" style="color:#00b42a;font-size:12px;margin-left:6px">（专属：{{ docChannel.group_price }}）</span>
            </div>
            <div class="doc-info-item">
              <span class="doc-label">协议</span>
              <el-tag size="small" type="info">{{ docChannel.protocol || 'openai' }}</el-tag>
            </div>
          </div>
          <div style="display:flex;flex-direction:column;align-items:center;gap:8px;flex-shrink:0">
            <img v-if="docChannel.icon_url" :src="docChannel.icon_url" style="width:40px;height:40px;border-radius:8px;object-fit:contain;border:1px solid #e4e7ed" alt="" />
            <div v-else class="card-icon card-icon-fallback" style="width:40px;height:40px;font-size:18px;border-radius:8px">{{ (docChannel.name||'?').charAt(0).toUpperCase() }}</div>
            <el-button size="small" type="primary" plain @click="openBalanceDocs">余额</el-button>
            <el-button size="small" type="success" plain @click="openTaskDocs">状态查询</el-button>
          </div>
        </div>

        <!-- 接口地址 -->
        <div class="doc-section-title">接口地址</div>
        <div class="doc-endpoint-row">
          <span class="method-badge">POST</span>
          <code class="doc-endpoint-text">{{ docEndpoint(docChannel) }}</code>
          <el-icon class="doc-copy-icon" @click.stop="copyText(docEndpoint(docChannel))" title="复制"><CopyDocument /></el-icon>
        </div>

        <!-- 模型名称 -->
        <div class="doc-section-title">model 字段值</div>
        <div class="doc-endpoint-row">
          <code class="doc-endpoint-text" style="font-size:14px">{{ docChannel.routing_model || docChannel.name }}</code>
          <el-icon class="doc-copy-icon" @click.stop="copyModel(docChannel.routing_model || docChannel.name)" title="复制"><CopyDocument /></el-icon>
        </div>

        <!-- 请求头 -->
        <div class="doc-section-title">请求头</div>
        <pre class="doc-pre">Authorization: Bearer YOUR_API_KEY</pre>

        <!-- Suno 模式选择 -->
        <template v-if="docChannel.type === 'music'">
          <div class="doc-section-title">生成模式</div>
          <el-radio-group v-model="sunoMode" size="small" style="margin-bottom:4px;flex-wrap:wrap;display:flex;gap:4px 0">
            <el-radio-button label="inspire" value="inspire">灵感模式</el-radio-button>
            <el-radio-button label="custom" value="custom">自定义模式</el-radio-button>
            <el-radio-button label="extend" value="extend">续写模式</el-radio-button>
            <el-radio-button label="overpainting" value="overpainting">添加人声</el-radio-button>
            <el-radio-button label="underpainting" value="underpainting">添加伴奏</el-radio-button>
          </el-radio-group>
        </template>

        <!-- 请求体 -->
        <div class="doc-section-title" style="display:flex;align-items:center;justify-content:space-between">
          <span>请求体示例</span>
          <el-button size="small" plain type="primary" @click.stop="copyText(docRequestBody(docChannel))">
            <el-icon><CopyDocument /></el-icon> 复制
          </el-button>
        </div>
        <pre class="doc-pre doc-pre-scroll">{{ docRequestBody(docChannel) }}</pre>

        <!-- 多语言代码示例 -->
        <div class="doc-section-title" style="display:flex;align-items:center;justify-content:space-between">
          <span>代码示例</span>
          <el-button size="small" plain type="primary" @click.stop="copyText(docCode(docChannel, langTab))">
            <el-icon><CopyDocument /></el-icon> 复制
          </el-button>
        </div>
        <el-tabs v-model="langTab" size="small" style="margin:0 0 4px">
          <el-tab-pane label="cURL" name="curl" />
          <el-tab-pane label="Python" name="python" />
          <el-tab-pane label="PHP" name="php" />
          <el-tab-pane label="Go" name="go" />
          <el-tab-pane label="Java" name="java" />
        </el-tabs>
        <pre class="doc-pre doc-pre-scroll">{{ docCode(docChannel, langTab) }}</pre>

        <!-- 响应示例 -->
        <div class="doc-section-title" style="display:flex;align-items:center;justify-content:space-between">
          <span>响应示例</span>
          <el-button size="small" plain type="primary" @click.stop="copyText(docResponse(docChannel))">
            <el-icon><CopyDocument /></el-icon> 复制
          </el-button>
        </div>
        <pre class="doc-pre doc-pre-scroll">{{ docResponse(docChannel) }}</pre>
      </template>
    </el-drawer>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, CopyDocument } from '@element-plus/icons-vue'
import { publicApi } from '@/api/index'

// ─── 数据 ───────────────────────────────────────────────
const loading = ref(true)
const channels = ref([])
const filterType = ref('')
const filterVendor = ref('')
const filterName = ref('')

const typeOptions = [
  { label: '全部', value: '' },
  { label: '文本对话', value: 'llm' },
  { label: '图片生成', value: 'image' },
  { label: '视频生成', value: 'video' },
  { label: '语音合成', value: 'audio' },
  { label: '音乐生成', value: 'music' },
]

const vendors = computed(() => {
  const set = new Set()
  channels.value.forEach(ch => { if (ch.vendor) set.add(ch.vendor) })
  return [...set].sort()
})

const filteredChannels = computed(() => {
  return channels.value.filter(ch => {
    if (filterType.value && ch.type !== filterType.value) return false
    if (filterVendor.value && ch.vendor !== filterVendor.value) return false
    if (filterName.value) {
      const q = filterName.value.toLowerCase()
      if (!(ch.name || '').toLowerCase().includes(q) &&
          !(ch.routing_model || '').toLowerCase().includes(q)) return false
    }
    return true
  })
})

onMounted(async () => {
  try {
    const res = await publicApi.listChannels()
    channels.value = res?.channels ?? res ?? []
  } catch {
    ElMessage.error('获取模型列表失败')
  } finally {
    loading.value = false
  }
})

function typeLabel(type) {
  const map = { llm: '文本对话', image: '图片生成', video: '视频生成', audio: '语音合成', music: '音乐生成' }
  return map[type] || type || 'LLM'
}

function copyModel(name) {
  navigator.clipboard.writeText(name).then(() => {
    ElMessage({ message: '模型名已复制', type: 'success', duration: 1200 })
  })
}

// ─── 文档弹窗
const docVisible = ref(false)
const docChannel = ref(null)
const docMode = ref('channel')  // 'channel' | 'balance'
const langTab = ref('curl')
const sunoMode = ref('inspire') // 'inspire' | 'custom'

function openDoc(ch) {
  docChannel.value = ch
  docMode.value = 'channel'
  langTab.value = 'curl'
  sunoMode.value = 'inspire'
  docVisible.value = true
}

function openBalanceDocs() {
  langTab.value = 'curl'
  docMode.value = 'balance'
  docVisible.value = true
}

function openTaskDocs() {
  langTab.value = 'curl'
  docMode.value = 'task'
  docVisible.value = true
}

function taskCode(lang) {
  const origin = window.location.origin
  if (lang === 'curl') {
    return `curl -X GET "${origin}/v1/tasks/YOUR_TASK_ID" \\\n  -H "Authorization: Bearer YOUR_API_KEY"`
  }
  if (lang === 'python') {
    return `import requests\n\nurl = "${origin}/v1/tasks/YOUR_TASK_ID"\nheaders = {"Authorization": "Bearer YOUR_API_KEY"}\n\nresponse = requests.get(url, headers=headers)\nprint(response.json())`
  }
  if (lang === 'php') {
    return `<?php\n$url = "${origin}/v1/tasks/YOUR_TASK_ID";\n\n$ch = curl_init($url);\ncurl_setopt_array($ch, [\n    CURLOPT_RETURNTRANSFER => true,\n    CURLOPT_HTTPHEADER     => ['Authorization: Bearer YOUR_API_KEY'],\n]);\n\necho curl_exec($ch);\ncurl_close($ch);`
  }
  if (lang === 'go') {
    return `package main\n\nimport (\n\t"fmt"\n\t"io"\n\t"net/http"\n)\n\nfunc main() {\n\treq, _ := http.NewRequest("GET", "${origin}/v1/tasks/YOUR_TASK_ID", nil)\n\treq.Header.Set("Authorization", "Bearer YOUR_API_KEY")\n\n\tresp, _ := (&http.Client{}).Do(req)\n\tdefer resp.Body.Close()\n\tdata, _ := io.ReadAll(resp.Body)\n\tfmt.Println(string(data))\n}`
  }
  if (lang === 'java') {
    return `import java.net.http.*;\nimport java.net.URI;\n\npublic class Main {\n    public static void main(String[] args) throws Exception {\n        var request = HttpRequest.newBuilder()\n            .uri(URI.create("${origin}/v1/tasks/YOUR_TASK_ID"))\n            .header("Authorization", "Bearer YOUR_API_KEY")\n            .GET()\n            .build();\n\n        var response = HttpClient.newHttpClient()\n            .send(request, HttpResponse.BodyHandlers.ofString());\n        System.out.println(response.body());\n    }\n}`
  }
  return ''
}

function taskResponse() {
  return JSON.stringify({
    task_id: 12345,
    task_type: 'music',
    status: 1,
    code: 200,
    msg: 'success',
    url: '',
    items: [
      { audio_url: 'https://cdn.suno.ai/example1.mp3', title: '为你歌唱', duration: 180 },
      { audio_url: 'https://cdn.suno.ai/example2.mp3', title: '为你歌唱', duration: 182 },
    ],
    credits_charged: 3600,
  }, null, 2)
}

const endpointMap = {
  llm: '/v1/chat/completions',
  image: '/v1/image',
  video: '/v1/video',
  audio: '/v1/audio',
  music: '/v1/music',
}

function docEndpoint(ch) {
  return endpointMap[ch.type] || '/v1/chat/completions'
}

function docRequestBody(ch) {
  const model = ch.routing_model || ch.name
  if (ch.type === 'llm') {
    return JSON.stringify({
      model,
      messages: [{ role: 'user', content: '你好，请介绍一下自己' }],
      stream: false,
    }, null, 2)
  }
  if (ch.type === 'image') {
    return JSON.stringify({ model, prompt: '一只可爱的橘猫坐在阳光下', size: '1k', aspect_ratio: '1:1', n: 1 }, null, 2)
  }
  if (ch.type === 'video') {
    return JSON.stringify({ model, prompt: '海浪拍打岸边，夕阳西下', size: '720p', aspect_ratio: '16:9', duration: '5' }, null, 2)
  }
  if (ch.type === 'audio') {
    return JSON.stringify({ model, input: '你好，欢迎使用语音合成服务', voice: 'alloy' }, null, 2)
  }
  if (ch.type === 'music') {
    if (sunoMode.value === 'custom') {
      return JSON.stringify({
        model,
        input_type: '20',
        prompt: '[主歌]\n周四的阳光晒脸庞\n微风轻轻吹过窗\n\n[副歌]\n周四快乐不散场\n欢声笑语满心房',
        title: '周四快乐',
        tags: 'pop,female voice',
        mv_version: 'chirp-v5',
        make_instrumental: false,
      }, null, 2)
    }
    if (sunoMode.value === 'extend') {
      return JSON.stringify({
        model,
        input_type: '20',
        prompt: '[Verse 1]\n小狗汪汪叫\n尾巴甩甩跳\n\n[Chorus]\n汪汪汪谁在听\n汪汪汪快乐行',
        title: '为你歌唱',
        tags: '',
        mv_version: 'chirp-v5',
        make_instrumental: false,
        continue_clip_id: 'https://cdn1.suno.ai/7c395650-62f2-4c4f-8b68-cf55b874c96c.mp3',
        continue_at: '27',
      }, null, 2)
    }
    if (sunoMode.value === 'overpainting') {
      return JSON.stringify({
        model,
        input_type: '20',
        prompt: '[Verse 1]\nUsah lepas kau pergi\nMungkin curhatku keliru\n\n[Chorus]\nAku bertanya pada diriku',
        title: 'Hi,melancholic',
        tags: 'pop,female voice',
        mv_version: 'chirp-v4-5+',
        make_instrumental: false,
        task: 'overpainting',
        metadata_params: {
          overpainting_clip_id: 'https://cdn1.suno.ai/21ae9c64-86ab-435a-b810-ed62727caf0a.mp3',
          overpainting_start_s: 0,
          overpainting_end_s: 57.9,
        },
      }, null, 2)
    }
    if (sunoMode.value === 'underpainting') {
      return JSON.stringify({
        model,
        input_type: '20',
        prompt: '',
        title: 'Hi,melancholic',
        tags: 'pop,female voice',
        mv_version: 'chirp-v4-5+',
        make_instrumental: true,
        task: 'underpainting',
        metadata_params: {
          underpainting_clip_id: 'https://cdn1.suno.ai/21ae9c64-86ab-435a-b810-ed62727caf0a.mp3',
          underpainting_start_s: 0,
          underpainting_end_s: 57.9,
        },
      }, null, 2)
    }
    // 灵感模式 input_type=10
    return JSON.stringify({
      model,
      input_type: '10',
      gpt_description_prompt: '轻快的爵士乐，适合咖啡馆氛围，女声演唱',
      mv_version: 'chirp-v5',
      make_instrumental: false,
    }, null, 2)
  }
  return JSON.stringify({ model, prompt: '...' }, null, 2)
}

function docCode(ch, lang) {
  const origin = window.location.origin
  const endpoint = docEndpoint(ch)
  const body = docRequestBody(ch)
  if (lang === 'curl') {
    return `curl -X POST "${origin}${endpoint}" \\\n  -H "Content-Type: application/json" \\\n  -H "Authorization: Bearer YOUR_API_KEY" \\\n  -d '${body}'`
  }
  if (lang === 'python') {
    return `import requests\nimport json\n\nurl = "${origin}${endpoint}"\nheaders = {\n    "Authorization": "Bearer YOUR_API_KEY",\n    "Content-Type": "application/json"\n}\nbody = json.loads('''${body}''')\n\nresponse = requests.post(url, headers=headers, json=body)\nprint(response.json())`
  }
  if (lang === 'php') {
    const safeBody = body.replace(/'/g, "\\'")
    return `<?php\n$url = "${origin}${endpoint}";\n$body = '${safeBody}';\n\n$ch = curl_init($url);\ncurl_setopt_array($ch, [\n    CURLOPT_RETURNTRANSFER => true,\n    CURLOPT_POST           => true,\n    CURLOPT_HTTPHEADER     => [\n        'Authorization: Bearer YOUR_API_KEY',\n        'Content-Type: application/json',\n    ],\n    CURLOPT_POSTFIELDS     => $body,\n]);\n\n$response = curl_exec($ch);\ncurl_close($ch);\necho $response;`
  }
  if (lang === 'go') {
    return `package main\n\nimport (\n\t"bytes"\n\t"fmt"\n\t"io"\n\t"net/http"\n)\n\nfunc main() {\n\tbody := []byte(\`${body}\`)\n\n\treq, _ := http.NewRequest("POST", "${origin}${endpoint}", bytes.NewBuffer(body))\n\treq.Header.Set("Authorization", "Bearer YOUR_API_KEY")\n\treq.Header.Set("Content-Type", "application/json")\n\n\tresp, _ := (&http.Client{}).Do(req)\n\tdefer resp.Body.Close()\n\tdata, _ := io.ReadAll(resp.Body)\n\tfmt.Println(string(data))\n}`
  }
  if (lang === 'java') {
    const escapedBody = body.replace(/\\/g, '\\\\').replace(/"/g, '\\"').replace(/\n/g, '\\n')
    return `import java.net.http.*;\nimport java.net.URI;\n\npublic class Main {\n    public static void main(String[] args) throws Exception {\n        String body = "${escapedBody}";\n\n        var request = HttpRequest.newBuilder()\n            .uri(URI.create("${origin}${endpoint}"))\n            .header("Authorization", "Bearer YOUR_API_KEY")\n            .header("Content-Type", "application/json")\n            .POST(HttpRequest.BodyPublishers.ofString(body))\n            .build();\n\n        var response = HttpClient.newHttpClient()\n            .send(request, HttpResponse.BodyHandlers.ofString());\n        System.out.println(response.body());\n    }\n}`
  }
  return ''
}

function docResponse(ch) {
  if (ch.type === 'llm') {
    return JSON.stringify({
      id: 'chatcmpl-abc123',
      object: 'chat.completion',
      model: ch.routing_model || ch.name,
      choices: [{
        index: 0,
        message: { role: 'assistant', content: '你好！我是一个人工智能助手，很高兴认识你。请问有什么我可以帮助你的吗？' },
        finish_reason: 'stop',
      }],
      usage: { prompt_tokens: 12, completion_tokens: 34, total_tokens: 46 },
    }, null, 2)
  }
  if (ch.type === 'image' || ch.type === 'video' || ch.type === 'music' || ch.type === 'audio') {
    return JSON.stringify({
      task_id: 'task_abc1234xyz',
      status: 'pending',
    }, null, 2)
  }
  return JSON.stringify({ task_id: 'task_abc1234xyz', status: 'pending' }, null, 2)
}

function balanceCode(lang) {
  const origin = window.location.origin
  if (lang === 'curl') {
    return `curl -X GET "${origin}/user/balance" \\\n  -H "Authorization: Bearer YOUR_API_KEY"`
  }
  if (lang === 'python') {
    return `import requests\n\nurl = "${origin}/user/balance"\nheaders = {"Authorization": "Bearer YOUR_API_KEY"}\n\nresponse = requests.get(url, headers=headers)\nprint(response.json())`
  }
  if (lang === 'php') {
    return `<?php\n$url = "${origin}/user/balance";\n\n$ch = curl_init($url);\ncurl_setopt_array($ch, [\n    CURLOPT_RETURNTRANSFER => true,\n    CURLOPT_HTTPHEADER     => ['Authorization: Bearer YOUR_API_KEY'],\n]);\n\necho curl_exec($ch);\ncurl_close($ch);`
  }
  if (lang === 'go') {
    return `package main\n\nimport (\n\t"fmt"\n\t"io"\n\t"net/http"\n)\n\nfunc main() {\n\treq, _ := http.NewRequest("GET", "${origin}/user/balance", nil)\n\treq.Header.Set("Authorization", "Bearer YOUR_API_KEY")\n\n\tresp, _ := (&http.Client{}).Do(req)\n\tdefer resp.Body.Close()\n\tdata, _ := io.ReadAll(resp.Body)\n\tfmt.Println(string(data))\n}`
  }
  if (lang === 'java') {
    return `import java.net.http.*;\nimport java.net.URI;\n\npublic class Main {\n    public static void main(String[] args) throws Exception {\n        var request = HttpRequest.newBuilder()\n            .uri(URI.create("${origin}/user/balance"))\n            .header("Authorization", "Bearer YOUR_API_KEY")\n            .GET()\n            .build();\n\n        var response = HttpClient.newHttpClient()\n            .send(request, HttpResponse.BodyHandlers.ofString());\n        System.out.println(response.body());\n    }\n}`
  }
  return ''
}

function balanceResponse() {
  return JSON.stringify({
    balance_credits: 1971573,
    balance_cny: 1.971573,
  }, null, 2)
}

function copyText(text) {
  navigator.clipboard.writeText(text).then(() => {
    ElMessage({ message: '已复制', type: 'success', duration: 1200 })
  })
}
</script>
<style scoped>
.models-page {
  padding-bottom: 60px;
}

/* 页面标题 */
.page-title {
  font-size: 24px;
  font-weight: 600;
  color: rgb(26, 27, 28);
  line-height: 32px;
  padding-bottom: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

/* 过滤器行 */
.filter-row {
  display: flex;
  align-items: flex-start;
  min-height: 32px;
  font-size: 16px;
  margin-bottom: 14px;
}
.filter-label {
  color: rgb(134, 147, 171);
  width: 70px;
  line-height: 32px;
  flex-shrink: 0;
  font-size: 16px;
}
.filter-groups {
  display: flex;
  flex-wrap: wrap;
  gap: 0;
  flex: 1;
}
.filter-tag {
  display: inline-flex;
  align-items: center;
  padding: 0 10px;
  height: 32px;
  border-radius: 6px;
  margin-right: 10px;
  margin-bottom: 6px;
  cursor: pointer;
  color: rgb(27, 35, 55);
  font-size: 14px;
  transition: all 0.15s;
  user-select: none;
  text-decoration: none;
}
.filter-tag:hover {
  background: rgba(22, 93, 255, 0.08);
  color: rgb(22, 93, 255);
}
.filter-tag.active {
  background: white;
  color: rgb(22, 93, 255);
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
}

/* 搜索行 */
.search-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}
.count-label {
  font-size: 13px;
  color: #86909c;
}
.count-label b {
  color: #165dff;
}

/* 网格 */
.model-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 14px 16px;
  padding-top: 5px;
}

/* 卡片 */
.model-card {
  background: white;
  border-radius: 8px;
  padding: 15px;
  cursor: pointer;
  border: 1px solid #f0f1f5;
  transition: box-shadow 0.15s, border-color 0.15s;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.model-card:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
  border-color: #d0e2ff;
}

/* 卡片头部 */
.card-header {
  display: flex;
  align-items: flex-start;
  gap: 8px;
}
.card-icon {
  width: 48px;
  height: 48px;
  border-radius: 6px;
  object-fit: contain;
  flex-shrink: 0;
  background: #f0f4ff;
}
.card-icon-fallback {
  display: flex;
  align-items: center;
  justify-content: center;
  background: #165dff;
  color: white;
  font-size: 20px;
  font-weight: 700;
  border-radius: 6px;
}
.card-meta {
  flex: 1;
  min-width: 0;
}
.card-name-row {
  display: flex;
  align-items: center;
  gap: 6px;
}
.card-name {
  font-size: 16px;
  font-weight: 700;
  color: #1d2129;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}
.copy-icon {
  font-size: 14px;
  color: #86909c;
  cursor: pointer;
  flex-shrink: 0;
  transition: color 0.15s;
}
.copy-icon:hover {
  color: #165dff;
}
.card-cost {
  font-size: 13px;
  color: rgb(100, 116, 139);
  margin-top: 6px;
}

/* 描述 */
.card-desc {
  font-size: 14px;
  color: rgb(100, 116, 139);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 20px;
  height: 40px;
}

/* 底部 */
.card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: auto;
}
.card-type-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  background: #f0f4ff;
  color: #165dff;
}
.card-type-tag.type-image { background: #e8f7e8; color: #00b42a; }
.card-type-tag.type-video { background: #fff4e5; color: #ff7d00; }
.card-type-tag.type-audio { background: #f0f4ff; color: #165dff; }
.card-status {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #86909c;
}
.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #00b42a;
  display: inline-block;
}

/* 加载骨架 */
.grid-loading {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 14px 16px;
}
.grid-loading .el-skeleton {
  background: white;
  border-radius: 8px;
  padding: 15px;
  border: 1px solid #f0f1f5;
}

/* 文档弹窗内容 */
.doc-info-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 32px;
  padding: 12px 0 16px;
  border-bottom: 1px solid #f0f1f5;
  margin-bottom: 4px;
}
.doc-info-item {
  display: flex;
  align-items: center;
  gap: 8px;
}
.doc-label {
  font-size: 13px;
  color: #86909c;
  flex-shrink: 0;
}
.doc-section-title {
  font-size: 13px;
  font-weight: 600;
  color: #1d2129;
  margin: 16px 0 8px;
}
.doc-endpoint-row {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #f7f8fa;
  border-radius: 6px;
  padding: 9px 12px;
}
.method-badge {
  background: #165dff;
  color: white;
  font-size: 11px;
  font-weight: 700;
  padding: 2px 7px;
  border-radius: 4px;
  flex-shrink: 0;
  letter-spacing: 0.5px;
}
.doc-endpoint-text {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  color: #1d2129;
  flex: 1;
  word-break: break-all;
}
.doc-copy-icon {
  font-size: 15px;
  color: #86909c;
  cursor: pointer;
  flex-shrink: 0;
  transition: color 0.15s;
}
.doc-copy-icon:hover { color: #165dff; }
.doc-pre {
  background: #1e1e2e;
  color: #cdd6f4;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12.5px;
  line-height: 1.6;
  border-radius: 6px;
  padding: 12px 14px;
  margin: 0;
  white-space: pre;
  overflow-x: auto;
}
.doc-pre-scroll {
  max-height: 200px;
  overflow-y: auto;
}
</style>
