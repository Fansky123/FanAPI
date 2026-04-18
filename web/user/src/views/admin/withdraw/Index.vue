<template>
  <div>
    <!-- 过滤栏 -->
    <el-card shadow="never" style="margin-bottom:16px">
      <div class="filter-row">
        <el-select v-model="filterStatus" placeholder="全部状态" clearable style="width:140px" @change="doSearch">
          <el-option label="待审核" value="pending" />
          <el-option label="已通过" value="approved" />
          <el-option label="已拒绝" value="rejected" />
        </el-select>
        <el-button type="primary" @click="doSearch">刷新</el-button>
        <el-badge :value="pendingCount || ''" :hidden="!pendingCount" type="danger" style="margin-left:auto">
          <span class="pending-tip">待处理：{{ pendingCount }} 条</span>
        </el-badge>
      </div>
    </el-card>

    <!-- 列表 -->
    <el-card shadow="never">
      <el-table :data="records" v-loading="loading" empty-text="暂无记录">
        <el-table-column label="ID" prop="id" width="70" />
        <el-table-column label="用户" prop="username" width="140" />
        <el-table-column label="申请时间" prop="created_at" width="170" :formatter="fmtTime" />
        <el-table-column label="积分数量" prop="amount" width="150" :formatter="(r)=>fmtCredits(r.amount)" />
        <el-table-column label="收款方式" prop="payment_type" width="100">
          <template #default="{ row }">
            <el-tag :type="row.payment_type === 'wechat' ? 'success' : 'primary'" size="small">
              {{ row.payment_type === 'wechat' ? '微信' : '支付宝' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" prop="status" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="备注" prop="admin_remark" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <template v-if="row.status === 'pending'">
              <el-button size="small" type="primary" @click="openQR(row)">查看收款码</el-button>
              <el-button size="small" type="success" @click="approve(row)">通过</el-button>
              <el-button size="small" type="danger" @click="openReject(row)">拒绝</el-button>
            </template>
            <template v-else>
              <el-button size="small" @click="openQR(row)">收款码</el-button>
            </template>
          </template>
        </el-table-column>
      </el-table>

      <div class="pager-wrap">
        <el-pagination v-model:current-page="page" :page-size="pageSize" :total="total"
          layout="total, prev, pager, next" @current-change="fetchList" />
      </div>
    </el-card>

    <!-- 收款码预览 -->
    <el-dialog v-model="qrVisible" title="用户收款码" width="320px" align-center>
      <div class="qr-dialog-body">
        <el-tag :type="curRow?.payment_type === 'wechat' ? 'success' : 'primary'" style="margin-bottom:10px">
          {{ curRow?.payment_type === 'wechat' ? '微信收款码' : '支付宝收款码' }}
        </el-tag>
        <img v-if="curRow?.payment_qr" :src="curRow.payment_qr" class="qr-img" />
        <el-empty v-else description="暂无收款码图片" />
        <div class="qr-amount">提现金额：<b>{{ fmtCredits(curRow?.amount) }}</b></div>
      </div>
      <template #footer>
        <el-button @click="qrVisible = false">关闭</el-button>
        <el-button type="success" v-if="curRow?.status === 'pending'" @click="approve(curRow); qrVisible = false">
          确认已转账 → 通过申请
        </el-button>
      </template>
    </el-dialog>

    <!-- 拒绝弹窗 -->
    <el-dialog v-model="rejectVisible" title="拒绝提现申请" width="400px" align-center>
      <el-form label-width="80px">
        <el-form-item label="拒绝原因">
          <el-input v-model="rejectRemark" type="textarea" :rows="3" placeholder="请填写拒绝原因（将展示给用户）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rejectVisible = false">取消</el-button>
        <el-button type="danger" :loading="acting" @click="doReject">确认拒绝</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { withdrawApi } from '@/api/admin'

const loading = ref(false)
const records = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const filterStatus = ref('')
const pendingCount = ref(0)
const acting = ref(false)

// QR 预览
const qrVisible = ref(false)
const curRow = ref(null)
function openQR(row) { curRow.value = row; qrVisible.value = true }

// 拒绝弹窗
const rejectVisible = ref(false)
const rejectRemark = ref('')
let rejectTarget = null
function openReject(row) { rejectTarget = row; rejectRemark.value = ''; rejectVisible.value = true }

function fmtCredits(v) {
  if (!v) return '0'
  return (v / 1e6).toFixed(4) + ' 积分'
}
function fmtTime(row, col, val) {
  return val ? new Date(val).toLocaleString('zh-CN') : '-'
}
function statusLabel(s) {
  return { pending: '待审核', approved: '已通过', rejected: '已拒绝' }[s] || s
}
function statusType(s) {
  return { pending: 'warning', approved: 'success', rejected: 'danger' }[s] || 'info'
}

async function fetchList() {
  loading.value = true
  try {
    const params = { page: page.value, size: pageSize }
    if (filterStatus.value) params.status = filterStatus.value
    const res = await withdrawApi.list(params)
    records.value = res.records || []
    total.value = res.total || 0
  } finally {
    loading.value = false
  }
}

async function fetchPendingCount() {
  try {
    const res = await withdrawApi.pendingCount()
    pendingCount.value = res.count || 0
  } catch {}
}

function doSearch() {
  page.value = 1
  fetchList()
}

async function approve(row) {
  await ElMessageBox.confirm(
    `确认通过 ${row.username} 的提现申请（${fmtCredits(row.amount)}）？\n请确保已完成线下转账后再操作。`,
    '确认审批通过', { type: 'warning', confirmButtonText: '确认通过', confirmButtonClass: 'el-button--success' }
  )
  acting.value = true
  try {
    await withdrawApi.approve(row.id)
    ElMessage.success('已通过，冻结积分已划扣')
    await Promise.all([fetchList(), fetchPendingCount()])
  } finally {
    acting.value = false
  }
}

async function doReject() {
  if (!rejectRemark.value.trim()) { ElMessage.warning('请填写拒绝原因'); return }
  acting.value = true
  try {
    await withdrawApi.reject(rejectTarget.id, rejectRemark.value)
    ElMessage.success('已拒绝该提现申请')
    rejectVisible.value = false
    await Promise.all([fetchList(), fetchPendingCount()])
  } finally {
    acting.value = false
  }
}

// 轮询待处理数
let pollTimer = null
onMounted(() => {
  fetchList()
  fetchPendingCount()
  pollTimer = setInterval(fetchPendingCount, 30000)
})
onUnmounted(() => clearInterval(pollTimer))
</script>

<style scoped>
.filter-row { display: flex; gap: 10px; align-items: center; }
.pending-tip { font-size: 14px; color: #f56c6c; font-weight: 600; }
.pager-wrap { display: flex; justify-content: flex-end; margin-top: 16px; }
.qr-dialog-body { display: flex; flex-direction: column; align-items: center; gap: 8px; }
.qr-img { width: 240px; height: 240px; object-fit: contain; border: 1px solid #eee; border-radius: 8px; }
.qr-amount { font-size: 14px; color: #606266; margin-top: 8px; }
</style>
