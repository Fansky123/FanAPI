<template>
  <div class="playground-page">
    <!-- Mode tabs header -->
    <div class="aic-title-header">
      <div class="mode-tabs">
        <button
          v-for="tab in TABS"
          :key="tab.key"
          class="mode-tab"
          :class="{ active: activeTab === tab.key }"
          @click="switchTab(tab.key)"
        >
          <el-icon class="tab-icon"><component :is="tab.icon" /></el-icon>
          {{ tab.label }}
        </button>
      </div>
      <el-button v-if="activeTab === 'llm'" size="small" style="border-radius:6px" @click="clearChat">
        新对话
      </el-button>
    </div>

    <!-- Two-panel layout -->
    <div class="chat-layout">
      <!-- Left: Configuration panel -->
      <div class="left-panel">
        <!-- Shared: API Key -->
        <div class="field-label"><span class="req">*</span> API 密钥：</div>
        <el-select
          v-model="selectedKeyId"
          placeholder="选择 API 密钥"
          size="large"
          style="width:100%"
        >
          <el-option
            v-for="k in apiKeys"
            :key="k.id"
            :label="k.name || k.key_prefix + '...'"
            :value="k.id"
          />
        </el-select>

        <!-- Shared: Model selector (filtered by tab type) -->
        <div class="field-label" style="margin-top:12px">
          模型 <span class="opt">(选填)</span>
        </div>
        <el-select
          v-model="selectedModel"
          placeholder="从模型列表选择"
          filterable
          clearable
          size="large"
          style="width:100%"
        >
          <el-option
            v-for="ch in filteredChannels"
            :key="ch.id"
            :label="ch.name"
            :value="ch.routing_model || ch.name"
          />
        </el-select>

        <!-- ── LLM params ─────────────────────────── -->
        <template v-if="activeTab === 'llm'">
          <div class="field-label" style="margin-top:12px">
            系统提示词 <span class="opt">(选填)</span>
          </div>
          <el-input
            v-model="systemPrompt"
            type="textarea"
            :rows="3"
            placeholder="例：你是一个有帮助的AI助手"
            size="large"
            style="width:100%"
          />

          <div class="field-label" style="margin-top:12px">
            Max Tokens <span class="opt">(选填)</span>
          </div>
          <el-input-number
            v-model="maxTokens"
            :min="1"
            :max="128000"
            :step="256"
            size="large"
            style="width:100%"
            placeholder="默认不限制"
          />

          <div class="slider-field" style="margin-top:12px">
            <div class="slider-label-row">
              <span class="field-label-inline">Temperature</span>
              <span class="opt">(选填)</span>
              <el-switch v-model="useTemp" size="small" style="margin-left:auto" />
            </div>
            <el-slider v-if="useTemp" v-model="temperature" :min="0" :max="2" :step="0.01" show-input style="margin-top:4px" />
          </div>

          <div class="slider-field" style="margin-top:12px">
            <div class="slider-label-row">
              <span class="field-label-inline">Top P</span>
              <span class="opt">(选填)</span>
              <el-switch v-model="useTopP" size="small" style="margin-left:auto" />
            </div>
            <el-slider v-if="useTopP" v-model="topP" :min="0" :max="1" :step="0.01" show-input style="margin-top:4px" />
          </div>
        </template>

        <!-- ── Image params ───────────────────────── -->
        <template v-if="activeTab === 'image'">
          <div class="field-label" style="margin-top:12px"><span class="req">*</span> 提示词：</div>
          <el-input
            v-model="imgForm.prompt"
            type="textarea"
            :rows="5"
            placeholder="描述你想生成的图片内容..."
            size="large"
            style="width:100%"
          />

          <div class="field-label" style="margin-top:12px">分辨率档位</div>
          <el-select v-model="imgForm.size" size="large" style="width:100%">
            <el-option label="1k (1024px)" value="1k" />
            <el-option label="2k (2048px)" value="2k" />
            <el-option label="3k (3072px)" value="3k" />
            <el-option label="4k (4096px)" value="4k" />
          </el-select>

          <div class="field-label" style="margin-top:12px">宽高比</div>
          <el-select v-model="imgForm.aspect_ratio" size="large" style="width:100%">
            <el-option label="1:1 方图" value="1:1" />
            <el-option label="16:9 横版" value="16:9" />
            <el-option label="9:16 竖版" value="9:16" />
            <el-option label="4:3" value="4:3" />
            <el-option label="3:4" value="3:4" />
          </el-select>

          <el-button type="primary" size="large" style="width:100%;margin-top:16px;border-radius:8px" :loading="running" @click="sendImage">
            生成图片
          </el-button>
        </template>

        <!-- ── Video params ───────────────────────── -->
        <template v-if="activeTab === 'video'">
          <div class="field-label" style="margin-top:12px"><span class="req">*</span> 提示词：</div>
          <el-input
            v-model="videoForm.prompt"
            type="textarea"
            :rows="5"
            placeholder="描述你想生成的视频内容..."
            size="large"
            style="width:100%"
          />

          <div class="field-label" style="margin-top:12px">分辨率</div>
          <el-select v-model="videoForm.size" size="large" style="width:100%">
            <el-option label="720p" value="720p" />
            <el-option label="1080p" value="1080p" />
          </el-select>

          <div class="field-label" style="margin-top:12px">宽高比</div>
          <el-select v-model="videoForm.aspect_ratio" size="large" style="width:100%">
            <el-option label="16:9 横屏" value="16:9" />
            <el-option label="9:16 竖屏" value="9:16" />
            <el-option label="1:1 方形" value="1:1" />
          </el-select>

          <div class="field-label" style="margin-top:12px">时长（秒）</div>
          <el-select v-model="videoForm.duration" size="large" style="width:100%">
            <el-option label="5 秒" value="5" />
            <el-option label="10 秒" value="10" />
            <el-option label="15 秒" value="15" />
          </el-select>

          <el-button type="primary" size="large" style="width:100%;margin-top:16px;border-radius:8px" :loading="running" @click="sendVideo">
            生成视频
          </el-button>
        </template>

        <!-- ── Music params ───────────────────────── -->
        <template v-if="activeTab === 'music'">
          <div class="field-label" style="margin-top:12px">创作模式</div>
          <el-radio-group v-model="musicForm.input_type" size="large" style="width:100%">
            <el-radio-button value="10" style="width:50%">灵感模式</el-radio-button>
            <el-radio-button value="20" style="width:50%">自定义模式</el-radio-button>
          </el-radio-group>

          <!-- 灵感模式 -->
          <template v-if="musicForm.input_type === '10'">
            <div class="field-label" style="margin-top:12px"><span class="req">*</span> 灵感描述：</div>
            <el-input
              v-model="musicForm.gpt_description_prompt"
              type="textarea"
              :rows="5"
              placeholder="描述你想要的音乐风格，例如：一首轻快的电子流行乐，节奏明快"
              size="large"
              style="width:100%"
            />
          </template>

          <!-- 自定义模式 -->
          <template v-else>
            <div class="field-label" style="margin-top:12px">歌词 <span class="opt">(选填)</span></div>
            <el-input
              v-model="musicForm.prompt"
              type="textarea"
              :rows="4"
              placeholder="填写歌词内容，支持 [Verse]/[Chorus] 等标记"
              size="large"
              style="width:100%"
            />
            <div class="field-label" style="margin-top:10px">风格标签 <span class="opt">(选填)</span></div>
            <el-input v-model="musicForm.tags" placeholder="如：pop, female voice, upbeat" size="large" style="width:100%" />
            <div class="field-label" style="margin-top:10px">歌曲标题 <span class="opt">(选填)</span></div>
            <el-input v-model="musicForm.title" placeholder="歌曲名称" size="large" style="width:100%" />
          </template>

          <div class="field-label" style="margin-top:12px">
            <el-checkbox v-model="musicForm.make_instrumental">纯音乐（无人声）</el-checkbox>
          </div>

          <el-button type="primary" size="large" style="width:100%;margin-top:16px;border-radius:8px" :loading="running" @click="sendMusic">
            生成音乐
          </el-button>
        </template>
      </div>

      <!-- Center: Result panel -->
      <div class="center-panel">

        <!-- ── LLM Chat ─────────────────────────────── -->
        <template v-if="activeTab === 'llm'">
          <div class="chat-content" ref="chatBox">
            <template v-if="messages.length === 0">
              <div class="empty-chat">
                <el-icon :size="48" color="#c9cdd4"><ChatDotRound /></el-icon>
                <div style="margin-top:12px;color:#86909c;font-size:14px">开始一段对话吧</div>
              </div>
            </template>
            <template v-else>
              <div
                v-for="(msg, i) in messages"
                :key="i"
                class="msg-row"
                :class="msg.role"
              >
                <div class="msg-avatar">
                  <el-icon v-if="msg.role === 'assistant'"><Service /></el-icon>
                  <el-icon v-else><User /></el-icon>
                </div>
                <div class="msg-bubble" :class="msg.role">
                  <div
                    v-if="msg.role === 'assistant'"
                    class="msg-text markdown"
                    v-html="renderMarkdown(msg.content)"
                  ></div>
                  <div v-else class="msg-text">{{ msg.content }}</div>
                  <div class="msg-meta" v-if="msg.tokens">
                    <span class="token-info">{{ msg.tokens }} tokens</span>
                  </div>
                </div>
              </div>
              <div v-if="streaming" class="msg-row assistant">
                <div class="msg-avatar"><el-icon><Service /></el-icon></div>
                <div class="msg-bubble assistant">
                  <div class="msg-text" v-html="renderMarkdown(streamingText)"></div>
                  <span class="cursor-blink">|</span>
                </div>
              </div>
            </template>
          </div>
          <div class="send-area">
            <div class="send-toolbar">
              <el-button link size="small" @click="clearChat" style="color:#86909c">
                <el-icon><Delete /></el-icon> 清空对话
              </el-button>
              <el-tag v-if="running" type="warning" size="small" effect="plain">生成中...</el-tag>
            </div>
            <div class="send-row">
              <el-input
                v-model="inputText"
                type="textarea"
                :rows="4"
                placeholder="输入消息，按 Enter 发送，Shift+Enter 换行"
                class="send-textarea"
                @keydown.enter.exact.prevent="sendMessage"
                @keydown.shift.enter.exact="() => {}"
                :disabled="running"
              />
            </div>
            <div class="send-actions">
              <span class="send-hint">按 Enter 发送；按 Shift+Enter 换行</span>
              <el-button
                type="primary"
                :loading="running"
                :disabled="!inputText.trim()"
                class="send-btn"
                @click="sendMessage"
              >
                <el-icon><Promotion /></el-icon>
                发送
              </el-button>
            </div>
          </div>
        </template>

        <!-- ── Image / Video / Music result ──────────── -->
        <template v-else>
          <div class="result-content">
            <!-- Task status bar -->
            <div v-if="taskStatus" class="task-status-bar">
              <el-tag v-if="taskStatus === 'polling'" type="warning" effect="plain">生成中…</el-tag>
              <el-tag v-else-if="taskStatus === 'done'" type="success" effect="plain">完成</el-tag>
              <el-tag v-else-if="taskStatus === 'failed'" type="danger" effect="plain">失败</el-tag>
              <el-progress
                v-if="taskStatus === 'polling'"
                :percentage="pollProgress"
                :stroke-width="4"
                status="striped"
                striped-flow
                :duration="10"
                style="flex:1;margin-left:12px"
              />
            </div>

            <!-- Image results -->
            <div v-if="activeTab === 'image' && resultImages.length" class="image-grid">
              <a v-for="(url, i) in resultImages" :key="i" :href="url" target="_blank" rel="noopener noreferrer">
                <img :src="url" class="result-img" :alt="`result-${i}`" />
              </a>
            </div>

            <!-- Video result -->
            <div v-if="activeTab === 'video' && resultVideoUrl" class="video-result">
              <video :src="resultVideoUrl" controls class="result-video"></video>
              <div style="margin-top:8px">
                <a :href="resultVideoUrl" target="_blank" rel="noopener noreferrer" class="download-link">
                  <el-icon><Download /></el-icon> 下载视频
                </a>
              </div>
            </div>

            <!-- Music results -->
            <div v-if="activeTab === 'music' && musicItems.length" class="music-list">
              <div v-for="(item, i) in musicItems" :key="i" class="music-card">
                <div class="music-info">
                  <img v-if="item.image_url" :src="item.image_url" class="music-cover" alt="cover" />
                  <div class="music-meta">
                    <div class="music-title">{{ item.title || `第 ${i+1} 首` }}</div>
                    <div v-if="item.tags" class="music-tags">{{ item.tags }}</div>
                    <div v-if="item.duration" class="music-duration">时长：{{ item.duration }}s</div>
                  </div>
                </div>
                <audio v-if="item.audio_url" :src="item.audio_url" controls class="music-player"></audio>
              </div>
            </div>

            <!-- Error text -->
            <pre v-if="resultError" class="response-pre">{{ resultError }}</pre>

            <!-- Empty state -->
            <div v-if="!taskStatus && !resultImages.length && !resultVideoUrl && !musicItems.length && !resultError" class="empty-state">
              <el-icon class="empty-icon">
                <Picture v-if="activeTab === 'image'" />
                <VideoCamera v-else-if="activeTab === 'video'" />
                <Headset v-else />
              </el-icon>
              <p>{{ activeTab === 'image' ? '填写参数后点击「生成图片」' : activeTab === 'video' ? '填写参数后点击「生成视频」' : '填写参数后点击「生成音乐」' }}</p>
            </div>
          </div>
        </template>

      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { userApi, publicApi } from '@/api'
import { useUserStore } from '@/stores/user'
import { useSiteStore } from '@/stores/site'
import { ElMessage } from 'element-plus'
import {
  ChatDotRound, User, Promotion, Delete, Service,
  Picture, VideoCamera, Headset, Download
} from '@element-plus/icons-vue'

const store = useUserStore()
const site = useSiteStore()

// ── Tab config ───────────────────────────────────────────────
const TABS = [
  { key: 'llm',   label: '文本对话', icon: ChatDotRound },
  { key: 'image', label: '图片生成', icon: Picture },
  { key: 'video', label: '视频生成', icon: VideoCamera },
  { key: 'music', label: '音乐生成', icon: Headset },
]
const activeTab = ref('llm')

function switchTab(tab) {
  if (activeTab.value === tab) return
  activeTab.value = tab
  selectedModel.value = ''
  stopPolling()
  taskStatus.value = ''
  resultImages.value = []
  resultVideoUrl.value = ''
  musicItems.value = []
  resultError.value = ''
}

// ── Shared state ─────────────────────────────────────────────
const apiKeys = ref([])
const selectedKeyId = ref(null)
const selectedModel = ref('')
const channels = ref([])
const running = ref(false)

// ── LLM state ────────────────────────────────────────────────
const systemPrompt = ref('')
const maxTokens = ref(null)
const temperature = ref(0.7)
const topP = ref(1)
const useTemp = ref(false)
const useTopP = ref(false)
const messages = ref([])
const inputText = ref('')
const streaming = ref(false)
const streamingText = ref('')
const chatBox = ref(null)

// ── Image state ───────────────────────────────────────────────
const imgForm = ref({ prompt: '', size: '1k', aspect_ratio: '1:1' })
const resultImages = ref([])

// ── Video state ───────────────────────────────────────────────
const videoForm = ref({ prompt: '', size: '720p', aspect_ratio: '16:9', duration: '5' })
const resultVideoUrl = ref('')

// ── Music state ───────────────────────────────────────────────
const musicForm = ref({
  input_type: '10',
  gpt_description_prompt: '',
  prompt: '',
  tags: '',
  title: '',
  make_instrumental: false,
})
const musicItems = ref([])

// ── Polling state ─────────────────────────────────────────────
const taskStatus = ref('')
const pollProgress = ref(0)
const resultError = ref('')
let pollTimer = null
let pollStart = 0
let currentTaskId = null
let currentApiKey = ''

// ── Computed: filtered channels by tab type ───────────────────
const filteredChannels = computed(() => {
  const typeMap = { llm: 'llm', image: 'image', video: 'video', music: 'music' }
  const t = typeMap[activeTab.value] || 'llm'
  return channels.value.filter(ch => ch.type === t)
})

onMounted(async () => {
  try {
    const res = await userApi.listAPIKeys()
    apiKeys.value = res.api_keys ?? []
    if (apiKeys.value.length > 0) selectedKeyId.value = apiKeys.value[0].id
  } catch {}

  try {
    const fn = store.isLoggedIn ? userApi.listChannels : publicApi.listChannels
    const res = await fn({ page: 1, per_page: 100 })
    channels.value = res.channels ?? []
  } catch {}
})

onUnmounted(() => stopPolling())

// ── Helper: get current API key string ───────────────────────
function getApiKey() {
  const key = apiKeys.value.find(k => k.id === selectedKeyId.value)
  return key?.raw_key || ''
}

// ── LLM ──────────────────────────────────────────────────────
async function sendMessage() {
  if (!inputText.value.trim() || running.value) return

  const userMsg = inputText.value.trim()
  inputText.value = ''
  messages.value.push({ role: 'user', content: userMsg })
  scrollToBottom()

  running.value = true
  streaming.value = true
  streamingText.value = ''

  try {
    const apiKeyStr = getApiKey()
    if (!apiKeyStr) {
      ElMessage.warning('请先选择 API 密钥（需要创建密钥后才能使用）')
      messages.value.pop()
      return
    }

    const msgs = []
    if (systemPrompt.value.trim()) {
      msgs.push({ role: 'system', content: systemPrompt.value.trim() })
    }
    messages.value.forEach(m => msgs.push({ role: m.role, content: m.content }))

    const body = {
      model: selectedModel.value || 'gpt-3.5-turbo',
      messages: msgs,
      stream: true
    }
    if (maxTokens.value) body.max_tokens = maxTokens.value
    if (useTemp.value) body.temperature = temperature.value
    if (useTopP.value) body.top_p = topP.value

    const baseUrl = site.apiBase || '/v1'
    const resp = await fetch(`${baseUrl}/chat/completions`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${apiKeyStr}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(body)
    })

    if (!resp.ok) {
      const errText = await resp.text()
      throw new Error(errText || `HTTP ${resp.status}`)
    }

    const reader = resp.body.getReader()
    const decoder = new TextDecoder()
    let accum = ''

    while (true) {
      const { value, done } = await reader.read()
      if (done) break
      const chunk = decoder.decode(value)
      const lines = chunk.split('\n')
      for (const line of lines) {
        if (!line.startsWith('data: ')) continue
        const data = line.slice(6).trim()
        if (data === '[DONE]') break
        try {
          const parsed = JSON.parse(data)
          const delta = parsed.choices?.[0]?.delta?.content || ''
          accum += delta
          streamingText.value = accum
          scrollToBottom()
        } catch {}
      }
    }

    messages.value.push({ role: 'assistant', content: accum })
    streaming.value = false
    streamingText.value = ''
  } catch (e) {
    streaming.value = false
    streamingText.value = ''
    ElMessage.error('请求失败：' + (e?.message || '未知错误'))
    if (messages.value[messages.value.length - 1]?.role === 'user') {
      messages.value.pop()
    }
  } finally {
    running.value = false
    scrollToBottom()
  }
}

function clearChat() {
  messages.value = []
  streamingText.value = ''
  streaming.value = false
}

function scrollToBottom() {
  nextTick(() => {
    if (chatBox.value) chatBox.value.scrollTop = chatBox.value.scrollHeight
  })
}

function renderMarkdown(text) {
  if (!text) return ''
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/```([\s\S]*?)```/g, '<pre><code>$1</code></pre>')
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.+?)\*/g, '<em>$1</em>')
    .replace(/\n/g, '<br>')
}

// ── Image ─────────────────────────────────────────────────────
async function sendImage() {
  if (!imgForm.value.prompt.trim()) return ElMessage.warning('请输入提示词')
  const apiKeyStr = getApiKey()
  if (!apiKeyStr) return ElMessage.warning('请先选择 API 密钥')
  if (!selectedModel.value) return ElMessage.warning('请选择图片模型')

  running.value = true
  resultImages.value = []
  taskStatus.value = ''
  resultError.value = ''
  stopPolling()

  const body = {
    model: selectedModel.value,
    prompt: imgForm.value.prompt,
    size: imgForm.value.size,
    aspect_ratio: imgForm.value.aspect_ratio,
  }

  try {
    const resp = await fetch('/v1/image', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${apiKeyStr}` },
      body: JSON.stringify(body),
    })
    const data = await resp.json()
    if (data.task_id) {
      currentTaskId = data.task_id
      currentApiKey = apiKeyStr
      taskStatus.value = 'polling'
      pollProgress.value = 5
      pollStart = Date.now()
      startPolling()
    } else if (data.data) {
      resultImages.value = data.data.map(d => d.url).filter(Boolean)
      taskStatus.value = 'done'
      running.value = false
    } else {
      resultError.value = JSON.stringify(data, null, 2)
      taskStatus.value = resp.ok ? '' : 'failed'
      running.value = false
    }
  } catch (e) {
    ElMessage.error('请求失败：' + (e?.message || '未知错误'))
    running.value = false
  }
}

// ── Video ─────────────────────────────────────────────────────
async function sendVideo() {
  if (!videoForm.value.prompt.trim()) return ElMessage.warning('请输入提示词')
  const apiKeyStr = getApiKey()
  if (!apiKeyStr) return ElMessage.warning('请先选择 API 密钥')
  if (!selectedModel.value) return ElMessage.warning('请选择视频模型')

  running.value = true
  resultVideoUrl.value = ''
  taskStatus.value = ''
  resultError.value = ''
  stopPolling()

  const body = {
    model: selectedModel.value,
    prompt: videoForm.value.prompt,
    size: videoForm.value.size,
    aspect_ratio: videoForm.value.aspect_ratio,
    duration: videoForm.value.duration,
  }

  try {
    const resp = await fetch('/v1/video', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${apiKeyStr}` },
      body: JSON.stringify(body),
    })
    const data = await resp.json()
    if (data.task_id) {
      currentTaskId = data.task_id
      currentApiKey = apiKeyStr
      taskStatus.value = 'polling'
      pollProgress.value = 5
      pollStart = Date.now()
      startPolling()
    } else {
      resultError.value = JSON.stringify(data, null, 2)
      taskStatus.value = resp.ok ? '' : 'failed'
      running.value = false
    }
  } catch (e) {
    ElMessage.error('请求失败：' + (e?.message || '未知错误'))
    running.value = false
  }
}

// ── Music ─────────────────────────────────────────────────────
async function sendMusic() {
  const apiKeyStr = getApiKey()
  if (!apiKeyStr) return ElMessage.warning('请先选择 API 密钥')
  if (!selectedModel.value) return ElMessage.warning('请选择音乐模型')
  if (musicForm.value.input_type === '10' && !musicForm.value.gpt_description_prompt.trim()) {
    return ElMessage.warning('请输入灵感描述')
  }

  running.value = true
  musicItems.value = []
  taskStatus.value = ''
  resultError.value = ''
  stopPolling()

  const body = {
    model: selectedModel.value,
    input_type: musicForm.value.input_type,
    make_instrumental: musicForm.value.make_instrumental,
  }
  if (musicForm.value.input_type === '10') {
    body.gpt_description_prompt = musicForm.value.gpt_description_prompt
  } else {
    body.prompt = musicForm.value.prompt
    body.tags = musicForm.value.tags
    body.title = musicForm.value.title
  }

  try {
    const resp = await fetch('/v1/music', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${apiKeyStr}` },
      body: JSON.stringify(body),
    })
    const data = await resp.json()
    if (data.task_id) {
      currentTaskId = data.task_id
      currentApiKey = apiKeyStr
      taskStatus.value = 'polling'
      pollProgress.value = 5
      pollStart = Date.now()
      startPolling()
    } else {
      resultError.value = JSON.stringify(data, null, 2)
      taskStatus.value = resp.ok ? '' : 'failed'
      running.value = false
    }
  } catch (e) {
    ElMessage.error('请求失败：' + (e?.message || '未知错误'))
    running.value = false
  }
}

// ── Polling ───────────────────────────────────────────────────
function startPolling() {
  pollTimer = setInterval(async () => {
    if (!currentTaskId) return
    const elapsed = (Date.now() - pollStart) / 1000
    pollProgress.value = Math.min(90, 5 + elapsed * 1.2)
    try {
      const resp = await fetch(`/v1/tasks/${currentTaskId}`, {
        headers: { Authorization: `Bearer ${currentApiKey}` }
      })
      if (!resp.ok) return
      const data = await resp.json()
      // code 150 = in progress
      if (data.code === 200 || data.status === 2) {
        stopPolling()
        pollProgress.value = 100
        taskStatus.value = 'done'
        running.value = false
        // extract results
        if (activeTab.value === 'image') {
          // result.data or result.url
          if (data.result?.data) {
            resultImages.value = data.result.data.map(d => d.url).filter(Boolean)
          } else if (data.url) {
            resultImages.value = [data.url]
          } else if (data.result?.url) {
            resultImages.value = [data.result.url]
          }
        } else if (activeTab.value === 'video') {
          resultVideoUrl.value = data.url || data.result?.url || ''
        } else if (activeTab.value === 'music') {
          if (data.items && Array.isArray(data.items)) {
            musicItems.value = data.items
          } else if (data.result?.items) {
            musicItems.value = data.result.items
          }
        }
      } else if (data.code >= 400 || data.status === 3) {
        stopPolling()
        taskStatus.value = 'failed'
        running.value = false
        resultError.value = data.msg || '生成失败'
      }
    } catch {}
  }, 3000)
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}
</script>

<style scoped>
.playground-page {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 100px);
  gap: 12px;
  min-height: 600px;
}

/* ── Header / tab bar ─────────────────────────────────── */
.aic-title-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: white;
  border-radius: 10px;
  padding: 10px 20px;
  border: 1px solid #e5e6eb;
  flex-shrink: 0;
}

.mode-tabs {
  display: flex;
  gap: 4px;
}

.mode-tab {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 7px 16px;
  border: 1px solid transparent;
  border-radius: 8px;
  background: transparent;
  color: #4e5969;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}
.mode-tab:hover {
  background: #f2f3f5;
  color: #1d2129;
}
.mode-tab.active {
  background: #e8f0ff;
  border-color: #bedaff;
  color: #165dff;
}
.tab-icon {
  font-size: 15px;
}

/* ── Two-panel layout ─────────────────────────────────── */
.chat-layout {
  display: flex;
  gap: 12px;
  flex: 1;
  min-height: 0;
}

.left-panel {
  width: 300px;
  flex-shrink: 0;
  background: white;
  border-radius: 10px;
  padding: 15px;
  border: 1px solid #e5e6eb;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.field-label {
  font-size: 13px;
  color: #1d2129;
  margin-bottom: 4px;
  font-weight: 500;
}
.field-label-inline {
  font-size: 13px;
  color: #1d2129;
  font-weight: 500;
}
.req { color: #f53f3f; margin-right: 2px; }
.opt { color: #86909c; font-size: 12px; }
.slider-field { display: flex; flex-direction: column; }
.slider-label-row {
  display: flex;
  align-items: center;
  gap: 4px;
}

/* ── Center panel ─────────────────────────────────────── */
.center-panel {
  flex: 1;
  background: white;
  border-radius: 10px;
  border: 1px solid #e5e6eb;
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
}

/* LLM: chat content */
.chat-content {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.empty-chat {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  padding: 60px 0;
}
.msg-row {
  display: flex;
  gap: 10px;
  align-items: flex-start;
}
.msg-row.user { flex-direction: row-reverse; }
.msg-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: #f0f4ff;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #165dff;
  flex-shrink: 0;
  font-size: 16px;
}
.msg-row.user .msg-avatar { background: #165dff; color: white; }
.msg-bubble {
  max-width: 80%;
  padding: 10px 14px;
  border-radius: 10px;
  font-size: 14px;
  line-height: 1.6;
  word-break: break-word;
}
.msg-bubble.assistant {
  background: #f5f7fa;
  color: #1d2129;
  border-radius: 2px 10px 10px 10px;
}
.msg-bubble.user {
  background: #165dff;
  color: white;
  border-radius: 10px 2px 10px 10px;
}
.msg-meta { margin-top: 4px; }
.token-info { font-size: 11px; color: #c9cdd4; }
.cursor-blink {
  animation: blink 1s step-end infinite;
  color: #165dff;
  font-weight: bold;
}
@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}
.markdown :deep(pre) {
  background: #1e1e2e;
  color: #e2e8f0;
  padding: 12px;
  border-radius: 6px;
  overflow-x: auto;
  font-size: 13px;
  margin: 8px 0;
}
.markdown :deep(code) {
  background: rgba(0,0,0,0.1);
  padding: 1px 5px;
  border-radius: 3px;
  font-size: 13px;
}
.markdown :deep(strong) { font-weight: 600; }

/* LLM: send area */
.send-area {
  border-top: 1px solid #f0f1f5;
  padding: 10px 14px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.send-toolbar { display: flex; align-items: center; gap: 8px; }
.send-textarea :deep(textarea) {
  background: #f5f7fa;
  border: 1px solid #e5e6eb;
  border-radius: 8px;
  font-size: 14px;
  resize: none;
}
.send-actions { display: flex; align-items: center; justify-content: space-between; }
.send-hint { font-size: 12px; color: #c9cdd4; }
.send-btn {
  border-radius: 20px;
  padding: 0 20px;
  height: 36px;
  font-size: 14px;
}

/* ── Result panel (image / video / music) ─────────────── */
.result-content {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.task-status-bar {
  display: flex;
  align-items: center;
  gap: 8px;
}

/* Image */
.image-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 12px;
}
.result-img {
  width: 100%;
  border-radius: 8px;
  display: block;
  cursor: zoom-in;
  transition: transform 0.2s;
}
.result-img:hover { transform: scale(1.02); }

/* Video */
.video-result { display: flex; flex-direction: column; gap: 8px; }
.result-video {
  width: 100%;
  max-width: 720px;
  border-radius: 10px;
  background: #000;
}
.download-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: #165dff;
  text-decoration: none;
}
.download-link:hover { text-decoration: underline; }

/* Music */
.music-list { display: flex; flex-direction: column; gap: 12px; }
.music-card {
  background: #f7f8fa;
  border-radius: 10px;
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.music-info { display: flex; gap: 12px; align-items: flex-start; }
.music-cover {
  width: 60px;
  height: 60px;
  border-radius: 8px;
  object-fit: cover;
  flex-shrink: 0;
}
.music-meta { display: flex; flex-direction: column; gap: 3px; }
.music-title { font-size: 14px; font-weight: 600; color: #1d2129; }
.music-tags { font-size: 12px; color: #86909c; }
.music-duration { font-size: 12px; color: #86909c; }
.music-player { width: 100%; }

/* Empty / Error */
.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: #c0c4cc;
  min-height: 300px;
}
.empty-icon { font-size: 48px; }
.response-pre {
  background: #fff8f8;
  border: 1px solid #fcc;
  border-radius: 6px;
  padding: 12px;
  font-size: 12px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
  color: #f53f3f;
}
</style>
