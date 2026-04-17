<template>
  <div class="gen-page">
    <div class="page-title-header">
      <h2>视频生成</h2>
    </div>

    <div class="gen-layout">
      <div class="gen-panel content-card">
        <div class="form-group">
          <label><span class="req">*</span> API 密钥</label>
          <el-select v-model="selectedKeyId" placeholder="选择 API 密钥" style="width:100%">
            <el-option
              v-for="k in apiKeys"
              :key="k.id"
              :label="k.name || k.key_prefix + '...'"
              :value="k.id"
            />
          </el-select>
        </div>

        <div class="form-group">
          <label><span class="req">*</span> 选择模型</label>
          <el-select v-model="selectedModel" placeholder="选择视频模型" style="width:100%" filterable @change="onModelChange">
            <el-option v-for="ch in videoChannels" :key="ch.id" :label="`${ch.name}${ch.price_display ? '  —  ' + ch.price_display : ''}`" :value="ch.routing_model || ch.name" />
          </el-select>
        </div>

        <div class="form-group">
          <label>提示词 <span class="hint">（Prompt）</span></label>
          <el-input v-model="form.prompt" type="textarea" :rows="5" placeholder="描述你想生成的视频内容，例如：海浪拍打礁石，慢动作镜头" />
        </div>

        <div class="form-group">
          <label>分辨率</label>
          <el-select v-model="form.size" style="width:100%">
            <el-option label="720p" value="720p" />
            <el-option label="1080p" value="1080p" />
            <el-option label="4K" value="4k" />
          </el-select>
        </div>

        <div class="form-group">
          <label>比例</label>
          <el-select v-model="form.aspect_ratio" style="width:100%">
            <el-option label="16:9（横屏）" value="16:9" />
            <el-option label="9:16（竖屏）" value="9:16" />
            <el-option label="1:1（方形）" value="1:1" />
          </el-select>
        </div>

        <div class="form-group">
          <label>时长（秒）</label>
          <el-input-number v-model="form.duration" :min="3" :max="60" controls-position="right" style="width:100%" />
        </div>

        <el-button type="primary" style="width:100%;margin-top:8px" :loading="running" @click="generate">
          <el-icon><VideoCamera /></el-icon>
          生成视频
        </el-button>
      </div>

      <div class="gen-result content-card">
        <div class="result-header">
          <span class="result-title">生成结果</span>
          <el-tag v-if="taskStatus === 'polling'" type="warning" size="small">生成中…</el-tag>
          <el-tag v-if="taskStatus === 'done'" type="success" size="small">完成</el-tag>
          <el-tag v-if="taskStatus === 'failed'" type="danger" size="small">失败</el-tag>
          <span v-if="taskId && taskStatus === 'polling'" style="font-size:12px;color:#86909c">Task #{{ taskId }}</span>
        </div>

        <el-progress v-if="taskStatus === 'polling'" :percentage="pollProgress" :stroke-width="4" status="striped" striped-flow :duration="10" style="margin-bottom:16px" />

        <div class="video-result" v-if="resultUrl">
          <video :src="resultUrl" controls class="result-video"></video>
          <div class="result-link">
            <el-icon><Link /></el-icon>
            <a :href="resultUrl" target="_blank" rel="noopener">{{ resultUrl }}</a>
          </div>
        </div>

        <div class="empty-state" v-else-if="taskStatus !== 'failed'">
          <el-icon class="empty-icon"><VideoCamera /></el-icon>
          <p>填写参数后点击「生成视频」</p>
          <p class="empty-sub">视频生成通常需要30秒~5分钟，请耐心等待</p>
        </div>

        <pre class="response-pre" v-if="responseText">{{ responseText }}</pre>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { userApi } from '@/api'
import { ElMessage } from 'element-plus'
import { VideoCamera, Link } from '@element-plus/icons-vue'

const channels = ref([])
const apiKeys = ref([])
const selectedKeyId = ref(null)
const selectedModel = ref(null)
const running = ref(false)
const responseText = ref('')
const resultUrl = ref('')
const taskStatus = ref('')
const pollProgress = ref(0)
let pollTimer = null
let pollStart = 0
let currentTaskId = null
let currentApiKey = null

const form = ref({ prompt: '', size: '1080p', aspect_ratio: '16:9', duration: 5 })

const videoChannels = computed(() => channels.value.filter(c => c.type === 'video'))

const getApiKey = () => {
  const k = apiKeys.value.find(k => k.id === selectedKeyId.value)
  return k?.raw_key || null
}

function onModelChange() {
  responseText.value = ''
  resultUrl.value = ''
  taskStatus.value = ''
}

async function generate() {
  if (!selectedModel.value) return ElMessage.warning('请选择视频模型')
  if (!form.value.prompt.trim()) return ElMessage.warning('请输入提示词')
  const apiKeyStr = getApiKey()
  if (!apiKeyStr) return ElMessage.warning('该密钥无法查看完整值，请重新创建一个 API 密钥')

  running.value = true
  responseText.value = ''
  resultUrl.value = ''
  taskStatus.value = ''
  stopPolling()

  const body = {
    model: selectedModel.value,
    prompt: form.value.prompt,
    size: form.value.size,
    aspect_ratio: form.value.aspect_ratio,
    duration: form.value.duration,
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
    } else if (data.url || data.data?.[0]?.url) {
      resultUrl.value = data.url || data.data[0].url
      taskStatus.value = 'done'
      running.value = false
    } else {
      responseText.value = JSON.stringify(data, null, 2)
      taskStatus.value = resp.ok ? '' : 'failed'
      running.value = false
    }
  } catch (e) {
    ElMessage.error('请求失败：' + (e?.message || '未知错误'))
    running.value = false
  }
}

function startPolling() {
  pollTimer = setInterval(async () => {
    if (!currentTaskId) return
    const elapsed = (Date.now() - pollStart) / 1000
    pollProgress.value = Math.min(90, 5 + elapsed * 0.3)
    try {
      const resp = await fetch(`/v1/tasks/${currentTaskId}`, {
        headers: { Authorization: `Bearer ${currentApiKey}` }
      })
      if (!resp.ok) return
      const data = await resp.json()
      if (data.code === 200 || data.status === 2) {
        stopPolling()
        pollProgress.value = 100
        taskStatus.value = 'done'
        running.value = false
        resultUrl.value = data.url || data.result?.url || data.result?.video_url || ''
      } else if (data.code >= 400 || data.status === 3) {
        stopPolling()
        taskStatus.value = 'failed'
        running.value = false
        responseText.value = data.msg || '生成失败'
      }
    } catch {}
  }, 5000)
}

function stopPolling() { if (pollTimer) { clearInterval(pollTimer); pollTimer = null } }

onMounted(async () => {
  try {
    const [chRes, keyRes] = await Promise.all([userApi.listChannels(), userApi.listAPIKeys()])
    channels.value = chRes.channels ?? []
    apiKeys.value = keyRes.api_keys ?? []
    if (apiKeys.value.length > 0) selectedKeyId.value = apiKeys.value[0].id
  } catch {}
})
onUnmounted(stopPolling)
</script>

<style scoped>
.gen-page { display: flex; flex-direction: column; }

.page-title-header {
  padding: 15px 24px;
  border-radius: 12px;
  background: #ffffff;
  border: 1px solid #ffffff;
  box-shadow: rgba(0,0,0,0.02) 0px 10px 20px 0px;
  margin-bottom: 15px;
}
.page-title-header h2 { margin: 0; font-size: 20px; font-weight: 600; color: rgb(26, 27, 28); }

.content-card { background: #ffffff; border-radius: 12px; padding: 20px; margin-bottom: 15px; }

.gen-layout { display: flex; gap: 15px; align-items: flex-start; }
.gen-panel { width: 320px; flex-shrink: 0; display: flex; flex-direction: column; gap: 14px; margin-bottom: 0; }
.gen-result { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 16px; margin-bottom: 0; }

.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-group label { font-size: 13px; color: #4e5969; font-weight: 500; }
.hint { color: #86909c; font-weight: 400; }

.result-header { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.result-title { font-size: 15px; font-weight: 600; color: #1d2129; }

.video-result { display: flex; flex-direction: column; gap: 12px; }
.result-video { width: 100%; max-height: 480px; border-radius: 8px; background: #000; }
.result-link { display: flex; align-items: center; gap: 6px; font-size: 12px; color: #165dff; }
.result-link a { color: #165dff; text-decoration: none; word-break: break-all; }
.result-link a:hover { text-decoration: underline; }

.empty-state { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 8px; color: #c0c4cc; min-height: 300px; }
.empty-icon { font-size: 48px; }
.empty-sub { font-size: 12px; }

.response-pre { background: #f7f8fa; border: 1px solid #e5e6eb; border-radius: 6px; padding: 12px; font-size: 12px; overflow: auto; white-space: pre-wrap; word-break: break-all; color: #f53f3f; }

@media (max-width: 768px) {
  .gen-layout { flex-direction: column; }
  .gen-panel { width: 100%; }
}
</style>
