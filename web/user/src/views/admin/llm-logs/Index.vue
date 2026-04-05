<template>
  <div class="llm-logs-page">
    <el-card class="hero-card">
      <div class="hero-row">
        <div>
          <div class="eyebrow">LLM Logs</div>
          <h3>LLM 请求日志</h3>
          <p>查看每次 LLM 请求的上游入参、响应状态、Token 用量及计费信息，方便排查问题。</p>
        </div>
      </div>
    </el-card>

    <el-card class="toolbar-card">
      <div class="toolbar-row">
        <el-input v-model="filters.corr_id" placeholder="Corr ID" clearable style="width: 220px" />
        <el-input v-model="filters.user_id" placeholder="用户 ID" clearable style="width: 110px" />
        <el-input v-model="filters.channel_id" placeholder="渠道 ID" clearable style="width: 110px" />
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
        <el-table-column prop="user_id" label="用户" width="80" />
        <el-table-column prop="channel_id" label="渠道" width="80" />
        <el-table-column prop="model" label="Model" min-width="160" show-overflow-tooltip />
        <el-table-column prop="is_stream" label="流式" width="70">
          <template #default="{ row }">
            <el-tag :type="row.is_stream ? 'primary' : 'info'" size="small">{{ row.is_stream ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="upstream_status" label="上游状态" width="90" />
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
        <el-table-column label="操作" width="80" align="center">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row.id)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="page"
        :page-size="pageSize"
        :total="total"
        style="margin-top:16px"
        @current-change="fetchLogs"
      />
    </el-card>

    <el-drawer v-model="drawerVisible" title="LLM 请求详情" size="60%">
      <template v-if="currentLog">
        <el-descriptions :column="2" border style="margin-bottom: 16px">
          <el-descriptions-item label="ID">{{ currentLog.id }}</el-descriptions-item>
          <el-descriptions-item label="用户 ID">{{ currentLog.user_id }}</el-descriptions-item>
          <el-descriptions-item label="渠道 ID">{{ currentLog.channel_id }}</el-descriptions-item>
          <el-descriptions-item label="API Key ID">{{ currentLog.api_key_id }}</el-descriptions-item>
          <el-descriptions-item label="Model">{{ currentLog.model || '-' }}</el-descriptions-item>
          <el-descriptions-item label="流式">{{ currentLog.is_stream ? '是' : '否' }}</el-descriptions-item>
          <el-descriptions-item label="上游 HTTP 状态">{{ currentLog.upstream_status }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusType(currentLog.status)" size="small">{{ currentLog.status }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="Corr ID" :span="2">{{ currentLog.corr_id }}</el-descriptions-item>
          <el-descriptions-item v-if="currentLog.error_msg" label="错误信息" :span="2">{{ currentLog.error_msg }}</el-descriptions-item>
          <el-descriptions-item label="创建时间" :span="2">{{ fmtTimeStr(currentLog.created_at) }}</el-descriptions-item>
        </el-descriptions>

        <div v-if="currentLog.usage" class="json-block">
          <div class="json-title">Token 用量</div>
          <pre>{{ pretty(currentLog.usage) }}</pre>
        </div>
        <div class="json-block">
          <div class="json-title">发往上游的请求体</div>
          <pre>{{ pretty(currentLog.upstream_request) }}</pre>
        </div>
        <div v-if="currentLog.upstream_response" class="json-block">
          <div class="json-title">上游响应体</div>
          <pre>{{ pretty(currentLog.upstream_response) }}</pre>
        </div>
      </template>
    </el-drawer>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue'
import { llmLogApi } from '@/api/admin'

const logs = ref([])
const page = ref(1)
const pageSize = 20
const total = ref(0)
const drawerVisible = ref(false)
const currentLog = ref(null)
const filters = reactive({
  corr_id: '',
  user_id: '',
  channel_id: '',
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
  Object.assign(filters, { corr_id: '', user_id: '', channel_id: '', model: '', status: '', dateRange: null })
  page.value = 1
  fetchLogs()
}

async function fetchLogs() {
  const params = { page: page.value, page_size: pageSize }
  if (filters.corr_id) params.corr_id = filters.corr_id
  if (filters.user_id) params.user_id = filters.user_id
  if (filters.channel_id) params.channel_id = filters.channel_id
  if (filters.model) params.model = filters.model
  if (filters.status) params.status = filters.status
  if (filters.dateRange?.[0]) params.start_at = filters.dateRange[0]
  if (filters.dateRange?.[1]) params.end_at = filters.dateRange[1]
  const res = await llmLogApi.list(params)
  logs.value = res.logs ?? []
  total.value = res.total ?? 0
}

async function openDetail(id) {
  const res = await llmLogApi.get(id)
  currentLog.value = res
  drawerVisible.value = true
}

function pretty(value) {
  return JSON.stringify(value ?? {}, null, 2)
}

function fmtTime(row, col, val) {
  return val ? new Date(val).toLocaleString('zh-CN') : '-'
}

function fmtTimeStr(val) {
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
  border: 1px solid #e4ecf7;
  border-radius: 10px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
  font-family: monospace;
  font-size: .82rem;
  max-height: 500px;
}
</style>
