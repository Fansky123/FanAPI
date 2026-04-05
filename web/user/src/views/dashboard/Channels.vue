<template>
  <div>
    <!-- 工具栏 -->
    <div class="toolbar">
      <el-input v-model="filterName" placeholder="搜索名称 / 模型..." clearable class="search-input">
        <template #prefix><el-icon><Search /></el-icon></template>
      </el-input>
      <el-select v-model="filterType" placeholder="全部类型" clearable style="width:130px">
        <el-option label="LLM 对话" value="llm" />
        <el-option label="图片生成" value="image" />
        <el-option label="视频生成" value="video" />
        <el-option label="音频生成" value="audio" />
      </el-select>
      <el-select v-model="filterProtocol" placeholder="全部协议" clearable style="width:130px">
        <el-option label="OpenAI" value="openai" />
        <el-option label="Claude" value="claude" />
        <el-option label="Gemini" value="gemini" />
      </el-select>
      <span class="count-tip">共 <b>{{ filteredChannels.length }}</b> 个可用模型</span>
    </div>

    <!-- 表格 -->
    <el-card class="table-card">
      <el-table :data="filteredChannels" stripe :loading="loading" row-class-name="model-row">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="渠道名称" min-width="180">
          <template #default="{ row }">
            <span class="name-text">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="model" label="模型" min-width="180">
          <template #default="{ row }">
            <el-tag size="small" effect="plain" type="info">{{ row.model }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="typeColor(row.type)" size="small">{{ row.type?.toUpperCase() }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="协议" width="100">
          <template #default="{ row }">
            <span class="proto-badge" :class="row.protocol">{{ row.protocol || 'openai' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="价格" min-width="220">
          <template #default="{ row }">
            <span class="price-text">{{ row.price_display || '免费' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="调用" width="140" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="copyId(row.id)">复制 ID={{ row.id }}</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && !filteredChannels.length" description="暂无可用模型" style="padding:40px 0" />
    </el-card>

    <!-- 调用提示 -->
    <el-card class="hint-card">
      <div class="hint-title">调用示例</div>
      <div class="hint-row">
        <el-tag type="info" size="small">OpenAI</el-tag>
        <code>POST /v1/chat/completions?channel_id=<b>&lt;ID&gt;</b></code>
        <el-tag size="small">X-API-Key: sk-xxx</el-tag>
      </div>
      <div class="hint-row">
        <el-tag type="warning" size="small">Claude</el-tag>
        <code>POST /v1/messages?channel_id=<b>&lt;ID&gt;</b></code>
      </div>
      <div class="hint-row">
        <el-tag type="success" size="small">Gemini</el-tag>
        <code>POST /v1/gemini?channel_id=<b>&lt;ID&gt;</b></code>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { publicApi } from '@/api'
import { ElMessage } from 'element-plus'

const channels = ref([])
const loading = ref(true)
const filterType = ref('')
const filterName = ref('')
const filterProtocol = ref('')

onMounted(async () => {
  try {
    const res = await publicApi.listChannels()
    channels.value = res.channels ?? []
  } finally {
    loading.value = false
  }
})

const filteredChannels = computed(() =>
  channels.value.filter(ch => {
    if (filterType.value && ch.type !== filterType.value) return false
    if (filterProtocol.value && (ch.protocol || 'openai') !== filterProtocol.value) return false
    if (filterName.value) {
      const kw = filterName.value.toLowerCase()
      if (!ch.name?.toLowerCase().includes(kw) && !ch.model?.toLowerCase().includes(kw)) return false
    }
    return true
  })
)

const typeColor = (t) => ({ llm: 'primary', image: 'success', video: 'warning', audio: 'info' }[t] ?? '')

function copyId(id) {
  navigator.clipboard.writeText(String(id))
  ElMessage.success(`channel_id=${id} 已复制`)
}
</script>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 14px;
  flex-wrap: wrap;
}
.search-input { width: 260px; }
.count-tip { color: #8a94a8; font-size: .82rem; margin-left: auto; }
.table-card { margin-bottom: 14px; }
.name-text { font-weight: 500; color: #0d1526; }
.price-text { font-size: .82rem; color: #1677ff; }
.proto-badge {
  display: inline-block;
  padding: 1px 8px;
  border-radius: 999px;
  font-size: .75rem;
  font-weight: 600;
  background: #f0f2f7;
  color: #454f63;
}
.proto-badge.claude { background: #fff4e5; color: #d46b08; }
.proto-badge.gemini { background: #e8f8f0; color: #237804; }
.proto-badge.openai { background: #e6f0ff; color: #1677ff; }
.hint-card { }
.hint-title { font-weight: 600; color: #0d1526; margin-bottom: 10px; font-size: .9rem; }
.hint-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 7px 0;
  border-bottom: 1px dashed #edf0f7;
  font-size: .82rem;
}
.hint-row:last-child { border: none; }
code { background: #f6f8fc; padding: 2px 8px; border-radius: 6px; font-family: monospace; }
</style>

