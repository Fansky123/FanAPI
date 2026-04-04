<template>
  <div class="tasks-page">
    <el-card class="hero-card">
      <div class="hero-row">
        <div>
          <div class="eyebrow">Task Center</div>
          <h3>异步任务与上游对接排障</h3>
          <p>查看用户提交的任务、平台入参、发送给第三方的请求体、第三方响应体，以及最终扣费信息。</p>
        </div>
      </div>
    </el-card>

    <el-card class="toolbar-card">
      <div class="toolbar-row">
        <el-input v-model="filters.task_id" placeholder="Task ID" clearable style="width: 120px" />
        <el-input v-model="filters.user_id" placeholder="用户 ID" clearable style="width: 120px" />
        <el-select v-model="filters.type" placeholder="任务类型" clearable style="width: 130px">
          <el-option label="图片" value="image" />
          <el-option label="视频" value="video" />
          <el-option label="音频" value="audio" />
        </el-select>
        <el-select v-model="filters.status" placeholder="状态" clearable style="width: 120px">
          <el-option label="排队中" value="pending" />
          <el-option label="处理中" value="processing" />
          <el-option label="成功" value="done" />
          <el-option label="失败" value="failed" />
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
      <el-table :data="tasks" stripe border>
        <el-table-column prop="id" label="Task ID" width="90" />
        <el-table-column prop="user_id" label="用户 ID" width="90" />
        <el-table-column prop="type" label="类型" width="90" />
        <el-table-column prop="status" label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="channel_id" label="渠道 ID" width="90" />
        <el-table-column label="扣费" width="120">
          <template #default="{ row }">{{ row.credits_charged?.toLocaleString?.() ?? row.credits_charged }} cr</template>
        </el-table-column>
        <el-table-column prop="upstream_task_id" label="第三方任务 ID" min-width="180" show-overflow-tooltip />
        <el-table-column prop="created_at" label="创建时间" min-width="180" :formatter="fmtTime" />
        <el-table-column label="操作" width="100" align="center">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row.id)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="page"
        :page-size="20"
        :total="total"
        style="margin-top:16px"
        @current-change="fetchTasks"
      />
    </el-card>

    <el-drawer v-model="drawerVisible" title="任务详情" size="56%">
      <template v-if="currentTask">
        <el-descriptions :column="2" border style="margin-bottom: 16px">
          <el-descriptions-item label="Task ID">{{ currentTask.id }}</el-descriptions-item>
          <el-descriptions-item label="用户 ID">{{ currentTask.user_id }}</el-descriptions-item>
          <el-descriptions-item label="任务类型">{{ currentTask.type }}</el-descriptions-item>
          <el-descriptions-item label="状态">{{ currentTask.status }}</el-descriptions-item>
          <el-descriptions-item label="渠道 ID">{{ currentTask.channel_id }}</el-descriptions-item>
          <el-descriptions-item label="扣费">{{ currentTask.credits_charged }} cr</el-descriptions-item>
          <el-descriptions-item label="第三方任务 ID" :span="2">{{ currentTask.upstream_task_id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="错误信息" :span="2">{{ currentTask.error_msg || '-' }}</el-descriptions-item>
        </el-descriptions>

        <div class="json-block">
          <div class="json-title">用户提交请求体</div>
          <pre>{{ pretty(currentTask.request) }}</pre>
        </div>
        <div class="json-block">
          <div class="json-title">发送给第三方的请求体</div>
          <pre>{{ pretty(currentTask.upstream_request) }}</pre>
        </div>
        <div class="json-block">
          <div class="json-title">第三方原始响应体</div>
          <pre>{{ pretty(currentTask.upstream_response) }}</pre>
        </div>
        <div class="json-block">
          <div class="json-title">平台标准结果</div>
          <pre>{{ pretty(currentTask.result) }}</pre>
        </div>
      </template>
    </el-drawer>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue'
import { taskApi } from '@/api/admin'

const tasks = ref([])
const page = ref(1)
const total = ref(0)
const drawerVisible = ref(false)
const currentTask = ref(null)
const filters = reactive({ task_id: '', user_id: '', type: '', status: '', dateRange: null })

onMounted(fetchTasks)

function doSearch() {
  page.value = 1
  fetchTasks()
}

function resetFilters() {
  filters.task_id = ''
  filters.user_id = ''
  filters.type = ''
  filters.status = ''
  filters.dateRange = null
  page.value = 1
  fetchTasks()
}

async function fetchTasks() {
  const params = { page: page.value, size: 20 }
  if (filters.task_id) params.task_id = filters.task_id
  if (filters.user_id) params.user_id = filters.user_id
  if (filters.type) params.type = filters.type
  if (filters.status) params.status = filters.status
  if (filters.dateRange?.[0]) params.start_at = filters.dateRange[0]
  if (filters.dateRange?.[1]) params.end_at = filters.dateRange[1]
  const res = await taskApi.list(params)
  tasks.value = res.tasks ?? []
  total.value = res.total ?? 0
}

async function openDetail(id) {
  const res = await taskApi.get(id)
  currentTask.value = res.task
  drawerVisible.value = true
}

function pretty(value) {
  return JSON.stringify(value ?? {}, null, 2)
}

function fmtTime(row, col, val) {
  return val ? new Date(val).toLocaleString('zh-CN') : '-'
}

function statusType(status) {
  return ({ pending: 'info', processing: 'warning', done: 'success', failed: 'danger' }[status] ?? 'info')
}
</script>

<style scoped>
.tasks-page { max-width: 1320px; }
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
}
@media (max-width: 900px) { .hero-row { flex-direction:column;align-items:flex-start; } }
</style>
