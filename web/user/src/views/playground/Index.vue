<template>
  <div class="playground">
    <el-card class="hero-card">
      <div class="hero-row">
        <div>
          <div class="eyebrow">Interactive API Lab</div>
          <h3>在线体验与任务调试</h3>
          <p>直接选择渠道并发起调用，支持 LLM 流式输出，以及图像/视频/音频任务轮询结果。</p>
        </div>
        <div class="hero-chip">实时调试</div>
      </div>
    </el-card>

    <!-- 配置栏 -->
    <el-card class="toolbar-card">
      <el-row :gutter="12">
        <el-col :span="12">
          <el-select
            v-model="selectedChannel"
            placeholder="选择渠道（channel_id）"
            style="width:100%"
            filterable
            @change="onChannelChange"
          >
            <el-option-group
              v-for="(group, type) in groupedChannels"
              :key="type"
              :label="type.toUpperCase()"
            >
              <el-option
                v-for="ch in group"
                :key="ch.id"
                :label="`${ch.name} — ${ch.price_display || ''}`"
                :value="ch.id"
              />
            </el-option-group>
          </el-select>
        </el-col>
        <el-col :span="8">
          <el-input v-model="apiKey" placeholder="API Key（sk-xxx）" show-password clearable />
        </el-col>
        <el-col :span="4">
          <el-button type="primary" style="width:100%" :loading="running" @click="runRequest">
            发送
          </el-button>
        </el-col>
      </el-row>
    </el-card>

    <!-- 输入 / 输出 -->
    <el-row :gutter="16">
      <el-col :span="12">
        <el-card class="editor-card">
          <template #header>
            <div class="card-head">请求 Body（JSON）</div>
          </template>
          <el-input
            v-model="requestBody"
            type="textarea"
            :rows="18"
            placeholder='{"model":"...", "messages":[{"role":"user","content":"你好"}]}'
            style="font-family:monospace;font-size:.85rem"
          />
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card class="result-card">
          <template #header>
            <div class="card-head status-head">
              <span>响应</span>
              <!-- 异步任务状态 -->
              <el-tag v-if="taskStatus === 'polling'" type="warning" size="small" effect="plain">
                轮询中… (task #{{ taskId }})
              </el-tag>
              <el-tag v-if="taskStatus === 'done'" type="success" size="small" effect="plain">
                完成
              </el-tag>
              <el-tag v-if="taskStatus === 'failed'" type="danger" size="small" effect="plain">
                失败
              </el-tag>
            </div>
          </template>

          <!-- 轮询进度条 -->
          <el-progress
            v-if="taskStatus === 'polling'"
            :percentage="pollProgress"
            :stroke-width="4"
            status="striped"
            striped-flow
            :duration="10"
            style="margin-bottom:10px"
          />

          <!-- 图片结果 -->
          <div v-if="resultImages.length" class="image-grid">
            <a
              v-for="(url, i) in resultImages"
              :key="i"
              :href="url"
              target="_blank"
              rel="noopener"
            >
              <img :src="url" class="result-img" :alt="`result-${i}`" />
            </a>
          </div>

          <!-- 视频 / 音频结果链接 -->
          <div v-if="resultUrl && !resultImages.length" class="result-link">
            <el-icon><Link /></el-icon>
            <a :href="resultUrl" target="_blank" rel="noopener">{{ resultUrl }}</a>
          </div>

          <!-- 文本 / JSON 响应 -->
          <pre class="response-pre" v-if="responseText">{{ responseText }}</pre>
          <pre class="response-pre placeholder" v-else-if="!resultImages.length && !resultUrl">（等待发送...）</pre>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { userApi } from '@/api'
import { ElMessage } from 'element-plus'
import { Link } from '@element-plus/icons-vue'

const channels = ref([])
const selectedChannel = ref(null)
const apiKey = ref(localStorage.getItem('playground_key') || '')
const requestBody = ref(JSON.stringify({
  model: 'claude-3-5-sonnet-20241022',
  messages: [{ role: 'user', content: '你好，请介绍一下你自己' }],
  max_tokens: 500
}, null, 2))
const responseText = ref('')
const running = ref(false)

// 异步任务轮询状态
const taskId = ref(null)
const taskStatus = ref('') // '' | 'polling' | 'done' | 'failed'
const pollProgress = ref(0)
const resultUrl = ref('')
const resultImages = ref([])
let pollTimer = null
let pollStart = 0
const POLL_TIMEOUT = 300_000 // 5 分钟

onMounted(async () => {
  const res = await userApi.listChannels()
  channels.value = res.channels ?? []
})

onUnmounted(() => stopPolling())

// 按 type 分组，方便渠道下拉分组显示
const groupedChannels = computed(() => {
  const groups = {}
  for (const ch of channels.value) {
    if (!groups[ch.type]) groups[ch.type] = []
    groups[ch.type].push(ch)
  }
  return groups
})

function onChannelChange() {
  const ch = channels.value.find(c => c.id === selectedChannel.value)
  if (!ch) return
  const templates = {
    llm: { model: 'claude-3-5-sonnet-20241022', messages: [{ role: 'user', content: '你好，请介绍一下你自己' }], max_tokens: 500 },
    image: { model: 'nano-banana-pro', prompt: '赛博朋克猫', size: '2k', aspect_ratio: '1:1', n: 1 },
    video: { model: 'video-gen-pro', prompt: '海浪拍打礁石', size: '1080p', aspect_ratio: '16:9', duration: 5 },
    audio: { model: 'music-gen', prompt: '一首轻快的爵士乐', style: 'jazz', duration: 30 },
  }
  if (templates[ch.type]) {
    requestBody.value = JSON.stringify(templates[ch.type], null, 2)
  }
}

function stopPolling() {
  if (pollTimer) { clearTimeout(pollTimer); pollTimer = null }
}

async function pollTask() {
  if (!taskId.value) return
  const elapsed = Date.now() - pollStart
  if (elapsed >= POLL_TIMEOUT) {
    taskStatus.value = 'failed'
    responseText.value = '任务超时（5 分钟未返回结果）'
    running.value = false
    return
  }
  pollProgress.value = Math.min(95, Math.round((elapsed / POLL_TIMEOUT) * 100))

  try {
    const res = await fetch(`/api/v1/tasks/${taskId.value}`, {
      headers: { 'X-API-Key': apiKey.value }
    })
    const data = await res.json()
    responseText.value = JSON.stringify(data, null, 2)

    if (data.code === 200) {
      taskStatus.value = 'done'
      pollProgress.value = 100
      running.value = false
      // 判断是图片（url 含常见图片扩展名或 channel type 是 image）
      if (data.url) {
        const ch = channels.value.find(c => c.id === selectedChannel.value)
        if (ch?.type === 'image' || /\.(png|jpg|jpeg|webp|gif)(\?|$)/i.test(data.url)) {
          resultImages.value = [data.url]
        } else {
          resultUrl.value = data.url
        }
      }
      return
    }
    if (data.code === 500) {
      taskStatus.value = 'failed'
      running.value = false
      return
    }
    // 还在进行中，继续轮询（间隔从 2s 逐渐增加到 5s）
    const interval = Math.min(5000, 2000 + Math.floor(elapsed / 30000) * 1000)
    pollTimer = setTimeout(pollTask, interval)
  } catch {
    const interval = Math.min(5000, 2000 + Math.floor((Date.now() - pollStart) / 30000) * 1000)
    pollTimer = setTimeout(pollTask, interval)
  }
}

async function runRequest() {
  if (!selectedChannel.value) return ElMessage.warning('请先选择渠道')
  if (!apiKey.value) return ElMessage.warning('请填写 API Key')
  let body
  try {
    body = JSON.parse(requestBody.value)
  } catch {
    return ElMessage.error('请求 Body 不是合法 JSON')
  }

  localStorage.setItem('playground_key', apiKey.value)
  running.value = true
  responseText.value = ''
  resultUrl.value = ''
  resultImages.value = []
  taskId.value = null
  taskStatus.value = ''
  pollProgress.value = 0
  stopPolling()

  try {
    const ch = channels.value.find(c => c.id === selectedChannel.value)
    const path = `/v1/${ch?.type ?? 'llm'}`
    const res = await fetch(`/api${path}?channel_id=${selectedChannel.value}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': apiKey.value,
      },
      body: JSON.stringify(body),
    })

    if (ch?.type === 'llm') {
      // SSE 流式输出
      const reader = res.body.getReader()
      const decoder = new TextDecoder()
      while (true) {
        const { done, value } = await reader.read()
        if (done) break
        responseText.value += decoder.decode(value, { stream: true })
      }
      running.value = false
    } else {
      // 异步任务（image / video / audio）
      const json = await res.json()
      responseText.value = JSON.stringify(json, null, 2)
      if (json.task_id) {
        taskId.value = json.task_id
        taskStatus.value = 'polling'
        pollStart = Date.now()
        pollTimer = setTimeout(pollTask, 2000)
        // running 保持 true 直到轮询结束
      } else {
        running.value = false
      }
    }
  } catch (e) {
    ElMessage.error('请求失败：' + e.message)
    running.value = false
  }
}
</script>

<style scoped>
.playground {
  max-width: 1320px;
}
.hero-card,
.toolbar-card {
  margin-bottom: 16px;
}
.hero-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}
.eyebrow {
  color: #1e66ff;
  font-size: .82rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: .08em;
}
.hero-row h3 {
  margin: 8px 0 10px;
  font-size: 1.55rem;
}
.hero-row p {
  margin: 0;
  color: #617086;
}
.hero-chip {
  padding: 10px 14px;
  border-radius: 999px;
  background: linear-gradient(90deg, #eef6ff, #eefcff);
  border: 1px solid #dce9ff;
  color: #155eef;
  font-weight: 700;
}
.card-head {
  font-weight: 700;
  color: #1a2b45;
}
.status-head {
  display: flex;
  align-items: center;
  gap: 10px;
}
.result-card :deep(.el-card__body) { padding: 14px 16px; }
.response-pre {
  white-space: pre-wrap; word-break: break-all; font-size: .82rem;
  font-family: monospace; min-height: 200px; color: #303133;
  background: #f7fafd; padding: 12px; border-radius: 10px; margin: 0;
  border: 1px solid #e4ecf7;
}
.response-pre.placeholder { color: #aaa; min-height: 400px; }
.image-grid {
  display: flex; flex-wrap: wrap; gap: 10px; margin-bottom: 12px;
}
.result-img {
  max-width: 100%; max-height: 320px; border-radius: 6px;
  object-fit: contain; border: 1px solid #e4e7ed;
  cursor: pointer; transition: opacity .2s;
}
.result-img:hover { opacity: .85; }
.result-link {
  display: flex; align-items: center; gap: 6px;
  padding: 10px 0; font-size: .9rem; margin-bottom: 8px;
  word-break: break-all;
}
.result-link a { color: #409eff; }

@media (max-width: 900px) {
  .hero-row {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
