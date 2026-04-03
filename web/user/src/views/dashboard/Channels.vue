<template>
  <div class="channel-page">
    <el-card class="hero-card">
      <div class="hero-row">
        <div>
          <div class="eyebrow">模型渠道</div>
          <h3>选择最适合的接入线路</h3>
          <p>按模型能力、价格和类型筛选，复制 channel_id 后可直接用于 SDK 或在线体验。</p>
        </div>
        <div class="hero-metrics">
          <div class="metric-box">
            <strong>{{ channels.length }}</strong>
            <span>全部渠道</span>
          </div>
          <div class="metric-box">
            <strong>{{ filteredChannels.length }}</strong>
            <span>筛选结果</span>
          </div>
        </div>
      </div>
    </el-card>

    <el-card class="filter-card">
      <div class="filter-row">
        <el-select v-model="filterType" placeholder="按类型筛选" clearable style="width:160px">
          <el-option label="LLM 对话" value="llm" />
          <el-option label="图片生成" value="image" />
          <el-option label="视频生成" value="video" />
          <el-option label="音频生成" value="audio" />
        </el-select>
        <el-input v-model="filterName" placeholder="搜索渠道名称或模型..." clearable class="filter-input">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
      </div>
    </el-card>

    <el-row :gutter="16">
      <el-col
        v-for="ch in filteredChannels"
        :key="ch.id"
        :xs="24"
        :sm="12"
        :lg="8"
        style="margin-bottom:16px"
      >
        <el-card shadow="hover" class="channel-card">
          <div class="channel-header">
            <el-tag :type="typeColor(ch.type)" size="small" effect="plain">{{ ch.type.toUpperCase() }}</el-tag>
            <span class="channel-name">{{ ch.name }}</span>
          </div>
          <div class="channel-model">{{ ch.model }}</div>
          <div class="channel-price">{{ ch.price_display || '价格面议' }}</div>
          <div class="channel-footer">
            <div class="channel-id">channel_id = <code>{{ ch.id }}</code></div>
            <el-button size="small" type="primary" plain @click="copy(ch.id)">复制</el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-empty v-if="!filteredChannels.length" description="暂无可用渠道" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { userApi } from '@/api'
import { ElMessage } from 'element-plus'

const channels = ref([])
const filterType = ref('')
const filterName = ref('')

onMounted(async () => {
  const res = await userApi.listChannels()
  channels.value = res.channels ?? []
})

const filteredChannels = computed(() =>
  channels.value.filter(ch => {
    if (filterType.value && ch.type !== filterType.value) return false
    if (filterName.value) {
      const keyword = filterName.value.toLowerCase()
      if (!ch.name.toLowerCase().includes(keyword) && !ch.model.toLowerCase().includes(keyword)) return false
    }
    return true
  })
)

const typeColor = (t) => ({ llm: 'primary', image: 'success', video: 'warning', audio: 'info' }[t] ?? '')

function copy(id) {
  navigator.clipboard.writeText(String(id))
  ElMessage.success(`channel_id=${id} 已复制`)
}
</script>

<style scoped>
.channel-page {
  max-width: 1320px;
}
.hero-card,
.filter-card {
  margin-bottom: 16px;
}
.hero-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
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
  font-size: 1.65rem;
}
.hero-row p {
  margin: 0;
  color: #617086;
}
.hero-metrics {
  display: flex;
  gap: 12px;
}
.metric-box {
  min-width: 120px;
  padding: 16px;
  border-radius: 16px;
  background: linear-gradient(180deg, #f7fbff, #eef5ff);
  border: 1px solid #d8e6ff;
}
.metric-box strong {
  display: block;
  font-size: 1.4rem;
  color: #0f172a;
}
.metric-box span {
  color: #72829a;
  font-size: .82rem;
}
.filter-row {
  display: flex;
  gap: 12px;
  align-items: center;
}
.filter-input {
  max-width: 320px;
}
.channel-card {
  cursor: default;
  min-height: 190px;
}
.channel-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 14px;
}
.channel-name {
  font-weight: 700;
  font-size: 1rem;
}
.channel-model {
  color: #68778f;
  font-size: .86rem;
  margin-bottom: 12px;
}
.channel-price {
  color: #155eef;
  font-weight: 700;
  font-size: .98rem;
  margin: 12px 0 20px;
}
.channel-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.channel-id {
  font-size: .82rem;
  color: #606266;
}
code {
  background: #f1f5fb;
  border: 1px solid #dde7f6;
  padding: 2px 6px;
  border-radius: 6px;
}

@media (max-width: 900px) {
  .hero-row,
  .filter-row,
  .channel-footer {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
