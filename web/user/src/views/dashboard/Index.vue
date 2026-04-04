<template>
  <div class="dash-page">
    <el-row :gutter="16" style="margin-bottom:18px">
      <el-col :span="8">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-label">账户余额</div>
          <div class="stat-value">¥{{ (store.balance / 1e6).toFixed(4) }}</div>
          <div class="stat-sub">{{ store.balance.toLocaleString() }} credits</div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-label">可用渠道</div>
          <div class="stat-value">{{ channelCount }}</div>
          <div class="stat-sub">个渠道</div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-label">API 密钥</div>
          <div class="stat-value">{{ keyCount }}</div>
          <div class="stat-sub">个密钥</div>
        </el-card>
      </el-col>
    </el-row>

    <el-card class="guide-card">
      <template #header>
        <div class="guide-title">快速入门</div>
      </template>
      <div class="quick-start">
        <p>1. 在 <router-link to="/keys">API 密钥</router-link> 页面创建你的密钥</p>
        <p>2. 在 <router-link to="/channels">渠道列表</router-link> 页面选择合适的渠道（可按价格筛选）</p>
        <p>3. 调用接口时传入 <code>?channel_id=xxx</code> 和 <code>X-API-Key: sk-xxx</code></p>
        <el-divider />
        <p style="color:#909399">Base URL：<code>https://your-domain.com</code></p>
        <el-table :data="apiEndpoints" style="margin-top:12px">
          <el-table-column prop="method" label="方法" width="80" />
          <el-table-column prop="path" label="路径" />
          <el-table-column prop="desc" label="说明" />
        </el-table>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import { userApi } from '@/api'

const store = useUserStore()
const channelCount = ref(0)
const keyCount = ref(0)

const apiEndpoints = [
  { method: 'POST', path: '/v1/chat/completions?channel_id=1', desc: 'LLM 对话（OpenAI 标准格式）' },
  { method: 'POST', path: '/v1/messages?channel_id=1',       desc: 'LLM 对话（Claude 原生格式）' },
  { method: 'POST', path: '/v1/gemini?channel_id=1',          desc: 'LLM 对话（Gemini 原生格式）' },
  { method: 'POST', path: '/v1/image?channel_id=2',           desc: '图片生成（异步）' },
  { method: 'POST', path: '/v1/video?channel_id=3',           desc: '视频生成（异步）' },
  { method: 'POST', path: '/v1/audio?channel_id=4',           desc: '音频生成（异步）' },
  { method: 'GET',  path: '/v1/tasks/:id',                    desc: '查询任务结果' },
]

onMounted(async () => {
  const [ch, keys] = await Promise.all([
    userApi.listChannels().catch(() => ({ channels: [] })),
    userApi.listAPIKeys().catch(() => ({ api_keys: [] })),
  ])
  channelCount.value = ch.channels?.length ?? 0
  keyCount.value = keys.api_keys?.length ?? 0
})
</script>

<style scoped>
.dash-page {
  max-width: 1320px;
}
.stat-card {
  background:
    linear-gradient(180deg, rgba(255, 255, 255, .92), rgba(247, 251, 255, .94));
}
.stat-label {
  color: #72829a;
  font-size: .84rem;
  margin-bottom: 8px;
}
.stat-value {
  font-size: 2rem;
  font-weight: 800;
  color: #101828;
  letter-spacing: .01em;
}
.stat-sub {
  color: #97a3b6;
  font-size: .8rem;
  margin-top: 4px;
}
.guide-card {
  overflow: hidden;
}
.guide-title {
  font-weight: 700;
  color: #1a2b45;
}
.quick-start p {
  margin: 8px 0;
  color: #52627a;
}
code {
  background: #f1f5fb;
  border: 1px solid #dce6f7;
  padding: 2px 6px;
  border-radius: 6px;
  font-size: .85rem;
}
</style>
