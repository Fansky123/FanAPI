<template>
  <div class="tasks-page">
    <el-card class="hero-card">
      <div class="hero-row">
        <div>
          <div class="eyebrow">My Tasks</div>
          <h3>我的任务记录</h3>
          <p>查看通过 API 提交的所有异步任务：状态、扣除积分与任务结果。</p>
        </div>
      </div>
    </el-card>

    <el-card class="toolbar-card">
      <div class="toolbar-row">
        <el-input v-model="filters.task_id" placeholder="Task ID" clearable style="width: 150px" />
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
        <el-button type="primary" @click="doSearch">查询</el-button>
        <el-button @click="resetFilters">重置</el-button>
      </div>
    </el-card>

    <el-card>
      <el-table :data="tasks" stripe border>
        <el-table-column prop="task_id" label="Task ID" width="110" />
        <el-table-column prop="task_type" label="类型" width="90" />
        <el-table-column prop="status" label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
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
        v-model:current-page="page"
        :page-size="20"
        :total="total"
        layout="total, prev, pager, next"
        style="margin-top:16px"
        @current-change="fetchTasks"
      />
    </el-card>

    <el-drawer v-model="drawerVisible" title="任务详情" size="52%">
      <template v-if="currentTask">
        <el-descriptions :column="2" border style="margin-bottom: 16px">
          <el-descriptions-item label="Task ID">{{ currentTask.task_id }}</el-descriptions-item>
          <el-descriptions-item label="任务类型">{{ currentTask.task_type }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusType(currentTask.status)" size="small">{{ statusLabel(currentTask.status) }}</el-tag>
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
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue'
import { taskApi } from '@/api'

const tasks = ref([])
const page = ref(1)
const total = ref(0)
const drawerVisible = ref(false)
const currentTask = ref(null)
const filters = reactive({ task_id: '', type: '', status: '' })

onMounted(fetchTasks)

function doSearch() {
  page.value = 1
  fetchTasks()
}

function resetFilters() {
  filters.task_id = ''
  filters.type = ''
  filters.status = ''
  page.value = 1
  fetchTasks()
}

async function fetchTasks() {
  const params = { page: page.value, size: 20 }
  if (filters.task_id) params.task_id = filters.task_id
  if (filters.type) params.type = filters.type
  if (filters.status) params.status = filters.status
  const res = await taskApi.list(params)
  tasks.value = res.tasks ?? []
  total.value = res.total ?? 0
}

async function openDetail(id) {
  const res = await taskApi.get(id)
  currentTask.value = res
  drawerVisible.value = true
}

function pretty(value) {
  if (!value) return '-'
  return JSON.stringify(value, null, 2)
}

function statusType(status) {
  return ({ pending: 'info', processing: 'warning', done: 'success', failed: 'danger' }[status] ?? 'info')
}

function statusLabel(status) {
  return ({ pending: '排队中', processing: '处理中', done: '已完成', failed: '失败' }[status] ?? status)
}
</script>

<style scoped>
.tasks-page { max-width: 1200px; }
.hero-card, .toolbar-card { margin-bottom: 16px; }
.hero-row { display: flex; align-items: center; justify-content: space-between; gap: 16px; }
.eyebrow { color: var(--cf-primary, #1e66ff); font-size: .82rem; font-weight: 700; text-transform: uppercase; letter-spacing: .08em; }
.hero-row h3 { margin: 8px 0 10px; font-size: 1.55rem; }
.hero-row p { margin: 0; color: #617086; }
.toolbar-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
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
</style>
