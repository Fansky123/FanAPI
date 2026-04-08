<template>
  <div class="llm-logs-page">
    <el-card class="hero-card">
      <div class="hero-row">
        <div>
          <div class="eyebrow">LLM Logs</div>
          <h3>LLM 请求日志</h3>
          <p>查看每次 LLM 请求的模型、Token 用量、扣费积分及状态。</p>
        </div>
      </div>
    </el-card>

    <el-card class="toolbar-card">
      <div class="toolbar-row">
        <el-input v-model="filters.corr_id" placeholder="Corr ID" clearable style="width: 220px" />
        <el-input v-model="filters.model" placeholder="Model" clearable style="width: 160px" />
        <el-select v-model="filters.status" placeholder="状态" clearable style="width: 120px">
          <el-option label="pending" value="pending" />
          <el-option label="ok" value="ok" />
          <el-option label="error" value="error" />
          <el-option label="refunded" value="refunded" />
        </el-select>
        <el-date-picker
          v-model="filters.dateRange"
          type="datetimerange"
          range-separator="至"
          start-placeholder="开始时间"
          end-placeholder="结束时间"
          value-format="YYYY-MM-DD HH:mm:ss"
          style="width: 380px"
        />
        <el-button type="primary" @click="doSearch">查询</el-button>
        <el-button @click="resetFilters">重置</el-button>
      </div>
    </el-card>

    <el-card>
      <el-table :data="logs" stripe border>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="model" label="Model" min-width="160" show-overflow-tooltip />
        <el-table-column prop="is_stream" label="流式" width="70">
          <template #default="{ row }">
            <el-tag :type="row.is_stream ? 'primary' : 'info'" size="small">{{ row.is_stream ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="扣费积分" width="140">
          <template #default="{ row }">
            <span v-if="row.credits_charged" style="color:#f56c6c">-{{ row.credits_charged.toLocaleString() }} cr</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="Token 用量" min-width="160">
          <template #default="{ row }">
            <span v-if="row.usage">
              ↑{{ row.usage.prompt_tokens ?? '-' }} / ↓{{ row.usage.completion_tokens ?? '-' }}
              <el-tag v-if="row.usage.estimated" type="warning" size="small">估算</el-tag>
            </span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="corr_id" label="Corr ID" min-width="240" show-overflow-tooltip />
        <el-table-column prop="created_at" label="时间" min-width="180" :formatter="fmtTime" />
      </el-table>

      <el-pagination
        v-model:current-page="page"
        :page-size="pageSize"
        :total="total"
        style="margin-top:16px"
        @current-change="fetchLogs"
      />
    </el-card>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue'
import { llmLogApi } from '@/api/index'

const logs = ref([])
const page = ref(1)
const pageSize = 20
const total = ref(0)
const filters = reactive({
  corr_id: '',
  model: '',
  status: '',
  dateRange: null,
})

onMounted(fetchLogs)

function doSearch() {
  page.value = 1
  fetchLogs()
}

function resetFilters() {
  Object.assign(filters, { corr_id: '', model: '', status: '', dateRange: null })
  page.value = 1
  fetchLogs()
}

async function fetchLogs() {
  const params = { page: page.value, page_size: pageSize }
  if (filters.corr_id) params.corr_id = filters.corr_id
  if (filters.model) params.model = filters.model
  if (filters.status) params.status = filters.status
  if (filters.dateRange?.[0]) params.start_at = filters.dateRange[0]
  if (filters.dateRange?.[1]) params.end_at = filters.dateRange[1]
  const res = await llmLogApi.list(params)
  logs.value = res.logs ?? []
  total.value = res.total ?? 0
}

function fmtTime(row, col, val) {
  return val ? new Date(val).toLocaleString('zh-CN') : '-'
}

function statusType(status) {
  return ({ ok: 'success', error: 'danger', refunded: 'warning', pending: 'info' }[status] ?? 'info')
}
</script>

<style scoped>
.llm-logs-page { max-width: 1400px; }
.hero-card, .toolbar-card { margin-bottom: 16px; }
.hero-row { display:flex;align-items:center;justify-content:space-between;gap:16px; }
.eyebrow { color:#1e66ff;font-size:.82rem;font-weight:700;text-transform:uppercase;letter-spacing:.08em; }
.hero-row h3 { margin:8px 0 10px;font-size:1.55rem; }
.hero-row p { margin:0;color:#617086; }
.toolbar-row { display:flex;align-items:center;gap:12px;flex-wrap:wrap; }
.json-block { margin-bottom: 16px; }
.json-title { font-weight: 700; margin-bottom: 8px; color: #1a2b45; }
pre {
  margin: 0;
  padding: 12px;
  background: #f7fafd;
  border-radius: 6px;
  font-size: 0.82rem;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
