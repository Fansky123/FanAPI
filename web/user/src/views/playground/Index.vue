<template>
  <div class="playground">
    <!-- 配置栏 -->
    <el-card style="margin-bottom:16px">
      <el-row :gutter="12">
        <el-col :span="12">
          <el-select
            v-model="selectedChannel"
            placeholder="选择渠道（channel_id）"
            style="width:100%"
            filterable
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
        <el-card>
          <template #header>请求 Body（JSON）</template>
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
        <el-card>
          <template #header>响应</template>
          <pre class="response-pre">{{ responseText || '（等待发送...）' }}</pre>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { userApi } from '@/api'
import { ElMessage } from 'element-plus'

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

onMounted(async () => {
  const res = await userApi.listChannels()
  channels.value = res.channels ?? []
})

// 按 type 分组，方便渠道下拉分组显示
const groupedChannels = computed(() => {
  const groups = {}
  for (const ch of channels.value) {
    if (!groups[ch.type]) groups[ch.type] = []
    groups[ch.type].push(ch)
  }
  return groups
})

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

  try {
    // 根据选中渠道类型决定请求路径
    const ch = channels.value.find(c => c.id === selectedChannel.value)
    const path = ch?.type === 'llm' ? '/v1/llm' : `/v1/${ch?.type ?? 'llm'}`
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
    } else {
      const json = await res.json()
      responseText.value = JSON.stringify(json, null, 2)
    }
  } finally {
    running.value = false
  }
}
</script>

<style scoped>
.response-pre {
  white-space: pre-wrap; word-break: break-all; font-size: .82rem;
  font-family: monospace; min-height: 400px; color: #303133;
  background: #fafafa; padding: 12px; border-radius: 6px; margin: 0;
}
</style>
