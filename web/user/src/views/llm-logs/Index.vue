<template>
  <div class="llm-logs-page">
    <div class="page-title-header"><h2>调用日志</h2></div>

    <div class="content-card toolbar-card">
      <div class="toolbar-row">
        <el-input v-model="filters.model" placeholder="模型名称" clearable style="width:160px" @keyup.enter="doSearch" />
        <el-select v-model="filters.status" placeholder="全部状态" clearable style="width:130px">
          <el-option label="ok" value="ok" />
          <el-option label="error" value="error" />
          <el-option label="refunded" value="refunded" />
          <el-option label="pending" value="pending" />
        </el-select>
        <el-date-picker
          v-model="filters.dateRange"
          type="datetimerange"
          range-separator="至"
          start-placeholder="开始时间"
          end-placeholder="结束时间"
          value-format="YYYY-MM-DD HH:mm:ss"
          style="width:380px"
        />
        <el-button type="primary" @click="doSearch">查询</el-button>
        <el-button @click="resetFilters">重置</el-button>
      </div>
    </div>

    <div class="content-card">
      <el-table :data="logs" stripe border>
        <el-table-column prop="model" label="模型" min-width="180" show-overflow-tooltip />
        <el-table-column prop="created_at" label="请求时间" width="180" :formatter="fmtTime" />
        <el-table-column label="输入 Tokens" width="120" align="right">
          <template #default="{ row }">
            <span v-if="row.usage?.prompt_tokens != null">{{ row.usage.prompt_tokens.toLocaleString() }}</span>
            <span v-else style="color:#c0c4cc">—</span>
          </template>
        </el-table-column>
        <el-table-column label="输出 Tokens" width="120" align="right">
          <template #default="{ row }">
            <span v-if="row.usage?.completion_tokens != null">
              {{ row.usage.completion_tokens.toLocaleString() }}
              <el-tag v-if="row.usage.estimated" type="warning" size="small" style="font-size:10px">估算</el-tag>
            </span>
            <span v-else style="color:#c0c4cc">—</span>
          </template>
        </el-table-column>
        <el-table-column label="消耗积分" width="140" align="right">
          <template #default="{ row }">
            <span v-if="row.credits_charged" style="color:#f56c6c;font-weight:600">
              -{{ (row.credits_charged / 1e6).toFixed(4) }}
            </span>
            <span v-else style="color:#c0c4cc">—</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="80" align="center" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" link @click="openDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="page"
        :page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next"
        style="margin-top:16px"
        @current-change="fetchLogs"
      />
    </div>

    <!-- 详情抽屉 -->
    <el-drawer v-model="detailVisible" title="日志详情" direction="rtl" size="560px" destroy-on-close>
      <div v-if="detailLoading" style="padding:40px;text-align:center">
        <el-icon class="is-loading" style="font-size:32px"><Loading /></el-icon>
      </div>
      <div v-else-if="detail" class="detail-content">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
          <el-descriptions-item label="模型">{{ detail.model }}</el-descriptions-item>
          <el-descriptions-item label="Corr ID">
            <span style="font-family:monospace;font-size:12px;word-break:break-all">{{ detail.corr_id }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusType(detail.status)" size="small">{{ statusLabel(detail.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="流式">{{ detail.is_stream ? '是' : '否' }}</el-descriptions-item>
          <el-descriptions-item label="输入 Tokens">{{ detail.usage?.prompt_tokens ?? '—' }}</el-descriptions-item>
          <el-descriptions-item label="输出 Tokens">
            {{ detail.usage?.completion_tokens ?? '—' }}
            <el-tag v-if="detail.usage?.estimated" type="warning" size="small" style="margin-left:6px">估算</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="消耗积分">
            <span v-if="detail.credits_charged" style="color:#f56c6c;font-weight:600">
              -{{ (detail.credits_charged / 1e6).toFixed(6) }}
            </span>
            <span v-else>—</span>
          </el-descriptions-item>
          <el-descriptions-item label="请求时间">{{ fmtTime(null, null, detail.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="完成时间">{{ detail.status !== 'pending' ? fmtTime(null, null, detail.updated_at) : '—' }}</el-descriptions-item>
          <el-descriptions-item v-if="detail.error_msg" label="错误信息">
            <span style="color:#f56c6c">{{ detail.error_msg }}</span>
          </el-descriptions-item>
        </el-descriptions>

        <template v-if="detail.client_request">
          <div class="detail-section-title">您发送的请求</div>
          <pre class="detail-pre">{{ JSON.stringify(detail.client_request, null, 2) }}</pre>
        </template>

        <template v-if="detail.client_response">
          <div class="detail-section-title">平台返回内容</div>
          <template v-if="detail.client_response.stream">
            <pre class="detail-pre">{{ detail.client_response.content }}</pre>
          </template>
          <template v-else>
            <pre class="detail-pre">{{ JSON.stringify(detail.client_response, null, 2) }}</pre>
          </template>
        </template>
      </div>
    </el-drawer>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue'
import { llmLogApi } from '@/api/index'
import { Loading } from '@element-plus/icons-vue'

const logs = ref([])
const page = ref(1)
const pageSize = 20
const total = ref(0)
const filters = reactive({ model: '', status: '', dateRange: null })

const detailVisible = ref(false)
const detailLoading = ref(false)
const detail = ref(null)

onMounted(fetchLogs)

function doSearch() { page.value = 1; fetchLogs() }
function resetFilters() {
  Object.assign(filters, { model: '', status: '', dateRange: null })
  page.value = 1; fetchLogs()
}

async function fetchLogs() {
  const params = { page: page.value, page_size: pageSize }
  if (filters.model) params.model = filters.model
  if (filters.status) params.status = filters.status
  if (filters.dateRange?.[0]) params.start_at = filters.dateRange[0]
  if (filters.dateRange?.[1]) params.end_at = filters.dateRange[1]
  const res = await llmLogApi.list(params)
  logs.value = res.logs ?? []
  total.value = res.total ?? 0
}

async function openDetail(row) {
  detailVisible.value = true
  detailLoading.value = true
  detail.value = null
  try {
    const res = await llmLogApi.get(row.id)
    detail.value = { ...res, credits_charged: row.credits_charged }
  } finally {
    detailLoading.value = false
  }
}

function fmtTime(row, col, val) {
  return val ? new Date(val).toLocaleString('zh-CN') : '-'
}

function statusType(s) {
  return ({ ok: 'success', error: 'danger', refunded: 'warning', pending: 'info' }[s] ?? 'info')
}

function statusLabel(s) {
  return ({ ok: '成功', error: '失败', refunded: '已退款', pending: '进行中' }[s] ?? s)
}
</script>

<style scoped>
.llm-logs-page { padding-bottom: 60px; }
.page-title-header h2 { font-size: 24px; font-weight: 600; color: #1a1b1c; margin: 0 0 16px; }
.content-card {
  background: white;
  border-radius: 8px;
  padding: 16px;
  border: 1px solid #f0f1f5;
  margin-bottom: 16px;
}
.toolbar-card { }
.toolbar-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.detail-content { padding: 4px 0; }
.detail-section-title {
  font-size: 13px; font-weight: 600; color: #1d2129;
  margin: 16px 0 8px;
}
.detail-pre {
  background: #1e1e2e;
  color: #cdd6f4;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  line-height: 1.6;
  border-radius: 6px;
  padding: 12px 14px;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 320px;
  overflow-y: auto;
}
</style>
