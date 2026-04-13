<template>
  <div class="logs-page">
    <el-tabs v-model="activeTab" @tab-change="onTabChange">
      <!-- ── LLM 日志 ── -->
      <el-tab-pane label="LLM 日志" name="llm">
        <el-card class="toolbar-card">
          <div class="toolbar-row">
            <el-input v-model="llm.filters.corr_id" placeholder="Corr ID" clearable style="width:220px" />
            <el-input v-model="llm.filters.model" placeholder="Model" clearable style="width:160px" />
            <el-select v-model="llm.filters.status" placeholder="状态" clearable style="width:120px">
              <el-option label="pending" value="pending" />
              <el-option label="ok" value="ok" />
              <el-option label="error" value="error" />
              <el-option label="refunded" value="refunded" />
            </el-select>
            <el-date-picker
              v-model="llm.filters.dateRange"
              type="datetimerange"
              range-separator="至"
              start-placeholder="开始时间"
              end-placeholder="结束时间"
              value-format="YYYY-MM-DD HH:mm:ss"
              style="width:380px"
            />
            <el-button type="primary" @click="llmSearch">查询</el-button>
            <el-button @click="llmReset">重置</el-button>
          </div>
        </el-card>

        <el-card>
          <el-table :data="llm.list" stripe border>
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
                <el-tag :type="llmStatusType(row.status)" size="small">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="corr_id" label="Corr ID" min-width="240" show-overflow-tooltip />
            <el-table-column prop="created_at" label="时间" min-width="180" :formatter="fmtTime" />
          </el-table>
          <el-pagination
            v-model:current-page="llm.page"
            :page-size="20"
            :total="llm.total"
            layout="total, prev, pager, next"
            style="margin-top:16px"
            @current-change="fetchLlm"
          />
        </el-card>
      </el-tab-pane>

      <!-- ── 任务记录 ── -->
      <el-tab-pane label="异步任务" name="task">
        <el-card class="toolbar-card">
          <div class="toolbar-row">
            <el-input v-model="task.filters.task_id" placeholder="Task ID" clearable style="width:150px" />
            <el-select v-model="task.filters.type" placeholder="任务类型" clearable style="width:130px">
              <el-option label="图片" value="image" />
              <el-option label="视频" value="video" />
              <el-option label="音频" value="audio" />
            </el-select>
            <el-select v-model="task.filters.status" placeholder="状态" clearable style="width:120px">
              <el-option label="排队中" value="pending" />
              <el-option label="处理中" value="processing" />
              <el-option label="成功" value="done" />
              <el-option label="失败" value="failed" />
            </el-select>
            <el-date-picker
              v-model="task.filters.dateRange"
              type="datetimerange"
              range-separator="至"
              start-placeholder="开始时间"
              end-placeholder="结束时间"
              value-format="YYYY-MM-DD HH:mm:ss"
              style="width:380px"
            />
            <el-button type="primary" @click="taskSearch">查询</el-button>
            <el-button @click="taskReset">重置</el-button>
          </div>
        </el-card>

        <el-card>
          <el-table :data="task.list" stripe border>
            <el-table-column prop="task_id" label="Task ID" width="110" />
            <el-table-column prop="task_type" label="类型" width="90" />
            <el-table-column prop="status" label="状态" width="110">
              <template #default="{ row }">
                <el-tag :type="taskStatusType(row.status)" size="small">{{ taskStatusLabel(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="扣费积分" width="130">
              <template #default="{ row }">
                {{ row.credits_charged?.toLocaleString?.() ?? row.credits_charged ?? '-' }} cr
              </template>
            </el-table-column>
            <el-table-column label="消息" min-width="180" show-overflow-tooltip>
              <template #default="{ row }">{{ row.msg || '-' }}</template>
            </el-table-column>
            <el-table-column label="操作" width="100" align="center">
              <template #default="{ row }">
                <el-button link type="primary" @click="openDetail(row.task_id)">详情</el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-pagination
            v-model:current-page="task.page"
            :page-size="20"
            :total="task.total"
            layout="total, prev, pager, next"
            style="margin-top:16px"
            @current-change="fetchTask"
          />
        </el-card>

        <el-drawer v-model="drawerVisible" title="任务详情" size="52%">
          <template v-if="currentTask">
            <el-descriptions :column="2" border style="margin-bottom:16px">
              <el-descriptions-item label="Task ID">{{ currentTask.task_id }}</el-descriptions-item>
              <el-descriptions-item label="任务类型">{{ currentTask.task_type }}</el-descriptions-item>
              <el-descriptions-item label="状态">
                <el-tag :type="taskStatusType(currentTask.status)" size="small">{{ taskStatusLabel(currentTask.status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="扣费积分">{{ currentTask.credits_charged ?? '-' }} cr</el-descriptions-item>
              <el-descriptions-item label="第三方任务 ID" :span="2">{{ currentTask.upstream_task_id || '-' }}</el-descriptions-item>
              <el-descriptions-item label="消息" :span="2">{{ currentTask.msg || '-' }}</el-descriptions-item>
            </el-descriptions>
            <div class="json-block">
              <div class="json-title">提交的请求体</div>
              <pre>{{ pretty(currentTask.request) }}</pre>
            </div>
            <div class="json-block">
              <div class="json-title">任务结果</div>
              <pre>{{ pretty(currentTask.result) }}</pre>
            </div>
          </template>
        </el-drawer>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { taskApi, llmLogApi } from '@/api'

const route = useRoute()
const activeTab = ref(route.query.tab === 'task' ? 'task' : 'llm')

// ── LLM 日志状态 ──
const llm = reactive({
  list: [], page: 1, total: 0,
  filters: { corr_id: '', model: '', status: '', dateRange: null }
})

// ── 任务状态 ──
const task = reactive({
  list: [], page: 1, total: 0,
  filters: { task_id: '', type: '', status: '', dateRange: null }
})
const drawerVisible = ref(false)
const currentTask = ref(null)

onMounted(() => {
  fetchLlm()
  fetchTask()
})

function onTabChange() {}

// LLM
function llmSearch() { llm.page = 1; fetchLlm() }
function llmReset() { Object.assign(llm.filters, { corr_id: '', model: '', status: '', dateRange: null }); llm.page = 1; fetchLlm() }
async function fetchLlm() {
  const params = { page: llm.page, page_size: 20 }
  if (llm.filters.corr_id) params.corr_id = llm.filters.corr_id
  if (llm.filters.model) params.model = llm.filters.model
  if (llm.filters.status) params.status = llm.filters.status
  if (llm.filters.dateRange?.[0]) params.start_at = llm.filters.dateRange[0]
  if (llm.filters.dateRange?.[1]) params.end_at = llm.filters.dateRange[1]
  const res = await llmLogApi.list(params)
  llm.list = res.logs ?? []
  llm.total = res.total ?? 0
}
function llmStatusType(s) { return ({ ok: 'success', error: 'danger', refunded: 'warning', pending: 'info' }[s] ?? 'info') }

// Task
function taskSearch() { task.page = 1; fetchTask() }
function taskReset() { Object.assign(task.filters, { task_id: '', type: '', status: '', dateRange: null }); task.page = 1; fetchTask() }
async function fetchTask() {
  const params = { page: task.page, size: 20 }
  if (task.filters.task_id) params.task_id = task.filters.task_id
  if (task.filters.type) params.type = task.filters.type
  if (task.filters.status) params.status = task.filters.status
  if (task.filters.dateRange?.[0]) params.start_at = task.filters.dateRange[0]
  if (task.filters.dateRange?.[1]) params.end_at = task.filters.dateRange[1]
  const res = await taskApi.list(params)
  task.list = res.tasks ?? []
  task.total = res.total ?? 0
}
async function openDetail(id) {
  const res = await taskApi.get(id)
  currentTask.value = res
  drawerVisible.value = true
}
function pretty(value) { return value ? JSON.stringify(value, null, 2) : '-' }
function taskStatusType(s) { return ({ pending: 'info', processing: 'warning', done: 'success', failed: 'danger' }[s] ?? 'info') }
function taskStatusLabel(s) { return ({ pending: '排队中', processing: '处理中', done: '已完成', failed: '失败' }[s] ?? s) }
function fmtTime(row, col, val) { return val ? new Date(val).toLocaleString('zh-CN') : '-' }
</script>

<style scoped>
.logs-page { max-width: 1400px; }
.toolbar-card { margin-bottom: 16px; }
.toolbar-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.json-block { margin-bottom: 16px; }
.json-title { font-weight: 700; margin-bottom: 8px; color: #1a2b45; }
pre {
  margin: 0; padding: 12px;
  background: #f7fafd;
  border: 1px solid #e4ecf7;
  border-radius: 6px;
  overflow: auto; white-space: pre-wrap; word-break: break-all;
  font-family: monospace; font-size: .82rem;
}
</style>
