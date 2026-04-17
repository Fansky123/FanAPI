<template>
  <div class="gen-page">
    <div class="page-title-header">
      <h2>图片生成</h2>
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
          <el-select v-model="selectedModel" placeholder="选择图片模型" style="width:100%" filterable @change="onModelChange">
            <el-option
              v-for="ch in imageChannels"
              :key="ch.id"
              :label="`${ch.name}${ch.price_display ? '  —  ' + ch.price_display : ''}`"
              :value="ch.routing_model || ch.name"
            />
          </el-select>
        </div>

        <div class="form-group">
          <label><span class="req">*</span> 提示词 <span class="hint">（Prompt）</span></label>
          <el-input v-model="form.prompt" type="textarea" :rows="5" placeholder="描述你想生成的图片，例如：赛博朋克风格的猫咪" />
        </div>

        <div class="form-group">
          <label>分辨率档位</label>
          <el-select v-model="form.size" style="width:100%">
            <el-option label="1k（约 1024px）" value="1k" />
            <el-option label="2k（约 2048px）" value="2k" />
            <el-option label="3k（约 3072px）" value="3k" />
            <el-option label="4k（约 4096px）" value="4k" />
          </el-select>
        </div>

        <div class="form-group">
          <label>宽高比</label>
          <el-select v-model="form.aspect_ratio" style="width:100%">
            <el-option label="1:1 方图" value="1:1" />
            <el-option label="16:9 横版" value="16:9" />
            <el-option label="9:16 竖版" value="9:16" />
            <el-option label="4:3" value="4:3" />
            <el-option label="3:4" value="3:4" />
          </el-select>
        </div>

        <div class="form-group">
          <label>参考图片 <span class="hint">（选填，每行一个 URL）</span></label>
          <el-input
            v-model="referImagesText"
            type="textarea"
            :rows="3"
            placeholder="https://example.com/ref1.jpg&#10;https://example.com/ref2.jpg"
          />
        </div>

        <el-button type="primary" style="width:100%;margin-top:8px" :loading="running" @click="generate">
          <el-icon><MagicStick /></el-icon>
          生成图片
        </el-button>
      </div>

      <div class="gen-result content-card">
        <div class="result-header">
          <span class="result-title">生成结果</span>
          <el-tag v-if="taskStatus === 'polling'" type="warning" size="small">生成中…</el-tag>
          <el-tag v-if="taskStatus === 'done'" type="success" size="small">完成</el-tag>
          <el-tag v-if="taskStatus === 'failed'" type="danger" size="small">失败</el-tag>
        </div>

        <el-progress v-if="taskStatus === 'polling'" :percentage="pollProgress" :stroke-width="4" status="striped" striped-flow :duration="10" style="margin-bottom:16px" />

        <div class="image-grid" v-if="resultImages.length">
          <a v-for="(url, i) in resultImages" :key="i" :href="url" target="_blank" rel="noopener">
            <img :src="url" class="result-img" :alt="`result-${i}`" />
          </a>
        </div>

        <div class="empty-state" v-else-if="!responseText">
          <el-icon class="empty-icon"><Picture /></el-icon>
          <p>填写参数后点击「生成图片」</p>
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
import { MagicStick, Picture } from '@element-plus/icons-vue'

const channels = ref([])
const apiKeys = ref([])
const selectedKeyId = ref(null)
const selectedModel = ref(null)
const running = ref(false)
const responseText = ref('')
const resultImages = ref([])
const taskStatus = ref('')
const pollProgress = ref(0)
const referImagesText = ref('')
let pollTimer = null
let pollStart = 0
let currentTaskId = null
let currentApiKey = null

const form = ref({ prompt: '', size: '1k', aspect_ratio: '1:1' })

const imageChannels = computed(() => channels.value.filter(c => c.type === 'image'))

const getApiKey = () => {
  const k = apiKeys.value.find(k => k.id === selectedKeyId.value)
  return k?.raw_key || null
}

function onModelChange() {
  responseText.value = ''
  resultImages.value = []
  taskStatus.value = ''
}

async function generate() {
  if (!selectedModel.value) return ElMessage.warning('请选择图片模型')
  if (!form.value.prompt.trim()) return ElMessage.warning('请输入提示词')
  const apiKeyStr = getApiKey()
  if (!apiKeyStr) return ElMessage.warning('该密钥无法查看完整值，请重新创建一个 API 密钥')

  running.value = true
  responseText.value = ''
  resultImages.value = []
  taskStatus.value = ''
  stopPolling()

  const body = {
    model: selectedModel.value,
    prompt: form.value.prompt,
    size: form.value.size,
    aspect_ratio: form.value.aspect_ratio,
  }

  // 解析参考图片（每行一个 URL，过滤空行）
  const refs = referImagesText.value.split('\n').map(s => s.trim()).filter(Boolean)
  if (refs.length > 0) body.refer_images = refs

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
    pollProgress.value = Math.min(90, 5 + elapsed * 1.2)
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
        if (data.result?.data) {
          resultImages.value = data.result.data.map(d => d.url).filter(Boolean)
        } else if (data.result?.url) {
          resultImages.value = [data.result.url]
        } else if (data.url) {
          resultImages.value = [data.url]
        }
      } else if (data.code >= 400 || data.status === 3) {
        stopPolling()
        taskStatus.value = 'failed'
        running.value = false
        responseText.value = data.msg || '生成失败'
      }
    } catch {}
  }, 3000)
}

function stopPolling() {
  if (pollTimer) { clearInterval(pollTimer); pollTimer = null }
}

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

.result-header { display: flex; align-items: center; gap: 10px; }
.result-title { font-size: 15px; font-weight: 600; color: #1d2129; }

.image-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: 12px; }
.result-img { width: 100%; border-radius: 8px; display: block; transition: transform .2s; cursor: zoom-in; }
.result-img:hover { transform: scale(1.02); }

.empty-state { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 10px; color: #c0c4cc; min-height: 300px; }
.empty-icon { font-size: 48px; }

.response-pre { background: #f7f8fa; border: 1px solid #e5e6eb; border-radius: 6px; padding: 12px; font-size: 12px; overflow: auto; white-space: pre-wrap; word-break: break-all; color: #f53f3f; }

@media (max-width: 768px) {
  .gen-layout { flex-direction: column; }
  .gen-panel { width: 100%; }
}
</style>
