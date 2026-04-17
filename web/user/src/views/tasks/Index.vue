<template>
  <div class="logs-page">
    <div class="page-title-header">
      <h2>调用日志</h2>
    </div>

    <div class="content-card filter-card">
      <div class="toolbar-row">
        <el-select v-model="task.filters.type" placeholder="全部类型" clearable style="width:130px">
          <el-option label="图片生成" value="image" />
          <el-option label="视频生成" value="video" />
          <el-option label="音频生成" value="audio" />
          <el-option label="音乐生成" value="music" />
        </el-select>
        <el-select v-model="task.filters.status" placeholder="全部状态" clearable style="width:130px">
          <el-option label="排队中" value="pending" />
          <el-option label="处理中" value="processing" />
          <el-option label="已完成" value="done" />
          <el-option label="失败" value="failed" />
        </el-select>
        <el-input v-model="task.filters.task_id" placeholder="任务 ID" clearable style="width:130px" @keyup.enter="taskSearch" />
        <el-date-picker
          v-model="task.filters.dateRange"
          type="datetimerange"
          range-separator="至"
          start-placeholder="开始时间"
          end-placeholder="结束时间"
          value-format="YYYY-MM-DD HH:mm:ss"
          style="width:360px"
        />
        <el-button type="primary" @click="taskSearch">查询</el-button>
        <el-button @click="taskReset">重置</el-button>
      </div>
    </div>

    <div class="content-card">
      <el-table :data="task.list" stripe border size="default">
        <el-table-column prop="task_id" label="任务 ID" width="90" align="center" />
        <el-table-column label="类型" width="110" align="center">
          <template #default="{ row }">
            <el-tag :type="typeTagType(row.task_type)" size="small">{{ typeLabel(row.task_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="请求时间" width="180" :formatter="fmtTime" />
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
            <el-tag :type="taskStatusType(statusCode(row.status))" size="small">{{ taskStatusLabel(statusCode(row.status)) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="错误信息" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="row.msg" style="color:#f56c6c;font-size:12px">{{ row.msg }}</span>
            <span v-else style="color:#c0c4cc">—</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="80" align="center" fixed="right">
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
    </div>

    <!-- 任务详情抽屉 -->
    <el-drawer v-model="drawerVisible" title="任务详情" direction="rtl" size="560px" destroy-on-close>
      <div v-if="detailLoading" style="padding:40px;text-align:center">
        <el-icon class="is-loading" style="font-size:32px"><Loading /></el-icon>
      </div>
      <div v-else-if="currentTask" class="detail-body">
        <el-descriptions :column="2" border size="small" style="margin-bottom:16px">
          <el-descriptions-item label="任务 ID">{{ currentTask.task_id }}</el-descriptions-item>
          <el-descriptions-item label="类型">
            <el-tag :type="typeTagType(currentTask.task_type)" size="small">{{ typeLabel(currentTask.task_type) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="状态" :span="2">
            <el-tag :type="taskStatusType(statusCode(currentTask.status))" size="small">{{ taskStatusLabel(statusCode(currentTask.status)) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="消耗积分" :span="2">
            <span v-if="currentTask.credits_charged" style="color:#f56c6c;font-weight:600">
              -{{ (currentTask.credits_charged / 1e6).toFixed(6) }}
            </span>
            <span v-else>—</span>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间" :span="2">{{ fmtTime(null, null, currentTask.created_at) }}</el-descriptions-item>
          <el-descriptions-item v-if="currentTask.finished_at" label="完成时间" :span="2">{{ fmtTime(null, null, currentTask.finished_at) }}</el-descriptions-item>
          <el-descriptions-item v-if="currentTask.upstream_task_id" label="上游任务 ID" :span="2">
            <span style="font-family:monospace;font-size:12px">{{ currentTask.upstream_task_id }}</span>
          </el-descriptions-item>
          <el-descriptions-item v-if="currentTask.msg" label="备注" :span="2">
            <span style="color:#f56c6c">{{ currentTask.msg }}</span>
          </el-descriptions-item>
        </el-descriptions>

        <!-- 请求参数 -->
        <template v-if="currentTask.request && Object.keys(currentTask.request).length">
          <div class="detail-section">
            <span class="detail-section-title">请求参数</span>
            <el-button size="small" plain type="primary" style="margin-left:auto" @click="copyJson(currentTask.request)">
              <el-icon><CopyDocument /></el-icon> 复制
            </el-button>
          </div>
          <pre class="detail-pre">{{ JSON.stringify(currentTask.request, null, 2) }}</pre>
        </template>

        <!-- 结果 / 响应 -->
        <template v-if="currentTask.result && Object.keys(currentTask.result).length">
          <div class="detail-section">
            <span class="detail-section-title">响应结果</span>
            <el-button size="small" plain type="primary" style="margin-left:auto" @click="copyJson(currentTask.result)">
              <el-icon><CopyDocument /></el-icon> 复制
            </el-button>
          </div>
          <pre class="detail-pre">{{ JSON.stringify(currentTask.result, null, 2) }}</pre>
        </template>

        <!-- 生成结果：直接展示媒体 -->
        <template v-if="currentTask.url || currentTask.items?.length">
          <div class="detail-section">
            <span class="detail-section-title">生成结果</span>
            <span v-if="currentTask.items?.length" style="font-size:12px;color:#86909c;margin-left:6px">（{{ currentTask.items.length }} 项）</span>
          </div>

          <!-- 单个结果 -->
          <template v-if="currentTask.url && !currentTask.items?.length">
            <div class="media-result-box">
              <img v-if="currentTask.task_type === 'image'" :src="currentTask.url" class="result-media result-img" @click="openMediaLink(currentTask.url)" alt="生成图片" />
              <video v-else-if="currentTask.task_type === 'video'" :src="currentTask.url" class="result-media result-video" controls />
              <audio v-else-if="currentTask.task_type === 'audio' || currentTask.task_type === 'music'" :src="currentTask.url" class="result-audio" controls />
              <a v-else :href="currentTask.url" target="_blank" class="result-link">{{ currentTask.url }}</a>
              <a v-if="currentTask.task_type === 'image'" :href="currentTask.url" target="_blank" class="result-dl-btn">↗ 在新标签查看原图</a>
            </div>
          </template>

          <!-- 多个结果（items） -->
          <div v-for="(item, i) in currentTask.items" :key="i" class="media-result-box" style="margin-bottom:12px">
            <template v-if="currentTask.task_type === 'image'">
              <img :src="itemUrl(item)" class="result-media result-img" @click="openMediaLink(itemUrl(item))" alt="生成图片" />
              <a :href="itemUrl(item)" target="_blank" class="result-dl-btn">↗ 查看原图</a>
            </template>
            <template v-else-if="currentTask.task_type === 'video'">
              <video :src="itemUrl(item)" class="result-media result-video" controls />
            </template>
            <template v-else-if="currentTask.task_type === 'audio' || currentTask.task_type === 'music'">
              <div v-if="itemTitle(item)" style="font-size:13px;font-weight:600;color:#1d2129;margin-bottom:4px">{{ itemTitle(item) }}</div>
              <div v-if="itemTags(item)" style="font-size:12px;color:#86909c;margin-bottom:6px">{{ itemTags(item) }}</div>
              <audio :src="itemUrl(item)" class="result-audio" controls />
              <a v-if="itemCover(item)" :href="itemCover(item)" target="_blank">
                <img :src="itemCover(item)" class="music-cover" alt="封面" />
              </a>
            </template>
            <template v-else>
              <a :href="itemUrl(item)" target="_blank" class="result-link">{{ itemUrl(item) }}</a>
            </template>
          </div>
        </template>
      </div>
    </el-drawer>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue'
import { taskApi } from '@/api'
import { Loading, CopyDocument } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const task = reactive({
  list: [], page: 1, total: 0,
  filters: { task_id: '', type: '', status: '', dateRange: null }
})

const drawerVisible = ref(false)
const detailLoading = ref(false)
const currentTask = ref(null)

onMounted(fetchTask)

function taskSearch() { task.page = 1; fetchTask() }
function taskReset() {
  Object.assign(task.filters, { task_id: '', type: '', status: '', dateRange: null })
  task.page = 1; fetchTask()
}

async function fetchTask() {
  const params = { page: task.page, size: 20 }
  if (task.filters.task_id) params.task_id = task.filters.task_id
  if (task.filters.type) params.type = task.filters.type
  if (task.filters.status) params.status = task.filters.status
  if (task.filters.dateRange?.[0]) params.start_at = task.filters.dateRange[0]
  if (task.filters.dateRange?.[1]) params.end_at = task.filters.dateRange[1]
  try {
    const res = await taskApi.list(params)
    task.list = res.tasks ?? []
    task.total = res.total ?? 0
  } catch {}
}

async function openDetail(id) {
  drawerVisible.value = true
  detailLoading.value = true
  currentTask.value = null
  try {
    currentTask.value = await taskApi.get(id)
  } finally {
    detailLoading.value = false
  }
}

// TaskResult.status: 0=pending,1=processing,2=done,3=failed
function statusCode(s) {
  return (['pending', 'processing', 'done', 'failed'][s] ?? String(s))
}

function taskStatusType(s) {
  return ({ pending: 'info', processing: 'warning', done: 'success', failed: 'danger' }[s] ?? 'info')
}
function taskStatusLabel(s) {
  return ({ pending: '排队中', processing: '处理中', done: '已完成', failed: '失败' }[s] ?? s)
}

function typeTagType(t) {
  return ({ image: '', video: 'warning', audio: 'success', music: 'danger' }[t] ?? 'info')
}
function typeLabel(t) {
  return ({ image: '图片生成', video: '视频生成', audio: '音频生成', music: '音乐生成' }[t] ?? t ?? '—')
}

function fmtTime(row, col, val) {
  return val ? new Date(val).toLocaleString('zh-CN') : '-'
}

function copyJson(obj) {
  navigator.clipboard.writeText(JSON.stringify(obj, null, 2)).then(() => {
    ElMessage({ message: '已复制', type: 'success', duration: 1200 })
  })
}

// 从 item 对象中提取 URL（兼容 { url }, { audio_url }, 字符串本身）
function itemUrl(item) {
  if (typeof item === 'string') return item
  return item.url || item.audio_url || item.video_url || item.image_url || ''
}
function itemTitle(item) { return item?.title || item?.name || '' }
function itemTags(item) { return item?.tags || item?.style || '' }
function itemCover(item) { return item?.image_url || item?.cover_url || item?.cover || '' }

function openMediaLink(url) {
  if (url && url.startsWith('data:')) return // base64 图片点击不跳转
  window.open(url, '_blank')
}
</script>

<style scoped>
.logs-page { display: flex; flex-direction: column; padding-bottom: 60px; }

.page-title-header {
  padding: 15px 24px;
  border-radius: 12px;
  background: #ffffff;
  border: 1px solid #f0f1f5;
  box-shadow: rgba(0,0,0,0.02) 0px 10px 20px 0px;
  margin-bottom: 15px;
}
.page-title-header h2 { margin: 0; font-size: 20px; font-weight: 600; color: rgb(26, 27, 28); }

.content-card { background: #ffffff; border-radius: 12px; padding: 20px; margin-bottom: 15px; border: 1px solid #f0f1f5; }
.filter-card { padding: 16px 20px; }
.toolbar-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }

.detail-body { padding: 4px 0; }
.detail-section { display: flex; align-items: center; margin: 16px 0 8px; }
.detail-section-title { font-size: 13px; font-weight: 600; color: #1d2129; }
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
  max-height: 280px;
  overflow-y: auto;
}
.result-url-row { margin-bottom: 8px; }
.result-link {
  color: #165dff;
  font-size: 12px;
  word-break: break-all;
  text-decoration: none;
}
.result-link:hover { text-decoration: underline; }

/* 媒体结果展示 */
.media-result-box {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 8px;
}
.result-media { border-radius: 8px; }
.result-img {
  max-width: 100%;
  max-height: 400px;
  object-fit: contain;
  cursor: zoom-in;
  border: 1px solid #e4e7ed;
}
.result-video {
  width: 100%;
  max-height: 360px;
  background: #000;
}
.result-audio {
  width: 100%;
}
.music-cover {
  width: 80px;
  height: 80px;
  border-radius: 6px;
  object-fit: cover;
  border: 1px solid #e4e7ed;
}
.result-dl-btn {
  font-size: 12px;
  color: #165dff;
  text-decoration: none;
}
.result-dl-btn:hover { text-decoration: underline; }
</style>
