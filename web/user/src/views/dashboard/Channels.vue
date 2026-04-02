<template>
  <div>
    <!-- 筛选栏 -->
    <div style="margin-bottom:16px;display:flex;gap:12px;align-items:center">
      <el-select v-model="filterType" placeholder="按类型筛选" clearable style="width:140px">
        <el-option label="LLM 对话" value="llm" />
        <el-option label="图片生成" value="image" />
        <el-option label="视频生成" value="video" />
        <el-option label="音频生成" value="audio" />
      </el-select>
      <el-input v-model="filterName" placeholder="搜索渠道名称..." clearable style="width:220px">
        <template #prefix><el-icon><Search /></el-icon></template>
      </el-input>
    </div>

    <!-- 渠道卡片列表 -->
    <el-row :gutter="16">
      <el-col
        v-for="ch in filteredChannels"
        :key="ch.id"
        :span="8"
        style="margin-bottom:16px"
      >
        <el-card shadow="hover" class="channel-card">
          <div class="channel-header">
            <el-tag :type="typeColor(ch.type)" size="small">{{ ch.type.toUpperCase() }}</el-tag>
            <span class="channel-name">{{ ch.name }}</span>
          </div>
          <div class="channel-model">模型：{{ ch.model }}</div>
          <div class="channel-price">{{ ch.price_display || '价格面议' }}</div>
          <div class="channel-id">
            channel_id = <code>{{ ch.id }}</code>
            <el-button size="small" link @click="copy(ch.id)">复制</el-button>
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
    if (filterName.value && !ch.name.toLowerCase().includes(filterName.value.toLowerCase())) return false
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
.channel-card { cursor: default; }
.channel-header { display: flex; align-items: center; gap: 8px; margin-bottom: 12px; }
.channel-name { font-weight: 600; font-size: .95rem; }
.channel-model { color: #909399; font-size: .82rem; margin-bottom: 6px; }
.channel-price { color: #f56c6c; font-weight: 600; margin: 8px 0; }
.channel-id { font-size: .82rem; color: #606266; }
code { background: #f5f5f5; padding: 2px 6px; border-radius: 4px; }
</style>
