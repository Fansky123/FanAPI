<template>
  <div class="invite-page">
    <!-- 邀请统计 -->
    <el-card class="panel" shadow="never">
      <template #header><span class="card-title">邀请中心</span></template>

      <div v-if="loading" class="loading-wrap"><el-skeleton :rows="4" animated /></div>
      <template v-else>
        <div class="stat-row">
          <div class="stat-item">
            <div class="stat-label">我的邀请码</div>
            <div class="stat-value code-value">
              {{ info.invite_code }}
              <el-button size="small" link @click="copyCode"><el-icon><CopyDocument /></el-icon></el-button>
            </div>
          </div>
          <div class="stat-item">
            <div class="stat-label">已邀请人数</div>
            <div class="stat-value">{{ info.invite_count }}</div>
          </div>
          <div class="stat-item">
            <div class="stat-label">冻结返佣积分</div>
            <div class="stat-value highlight">{{ formatCredits(info.frozen_balance) }}</div>
          </div>
        </div>

        <el-divider />
        <div class="link-section">
          <div class="section-label">邀请链接</div>
          <div class="link-row">
            <el-input :value="inviteLink" readonly class="link-input" />
            <el-button type="primary" @click="copyLink">复制链接</el-button>
          </div>
          <div class="form-tip">将此链接分享给好友，对方通过该链接注册后即成为您的邀请用户</div>
        </div>

        <!-- 解冻为可用积分 -->
        <el-divider />
        <div class="convert-section">
          <div class="section-label">解冻为可用积分</div>
          <p class="form-tip">将冻结返佣积分兑换为可用积分，当前可解冻：<b>{{ formatCredits(info.frozen_balance) }}</b></p>
          <div class="convert-row">
            <el-input-number v-model="convertAmount" :min="0" :max="info.frozen_balance" :step="100"
              placeholder="0 表示全部" style="width:200px" />
            <el-button type="primary" :loading="converting" @click="doConvert"
              :disabled="info.frozen_balance <= 0">解冻</el-button>
          </div>
        </div>

        <!-- 返佣规则 -->
        <el-divider />
        <el-alert type="info" :closable="false" show-icon>
          <template #title>返佣规则</template>
          <template #default>
            <ul class="rule-list">
              <li>您邀请的用户每次消费，系统将按比例将返佣积分冻结至您的账户</li>
              <li>冻结积分可随时解冻为可用积分，也可申请提现（平台手动转账）</li>
              <li>具体返佣比例以平台设置为准，请咨询客服</li>
            </ul>
          </template>
        </el-alert>
      </template>
    </el-card>

    <!-- 收款码设置 -->
    <el-card class="panel" shadow="never" style="margin-top:16px">
      <template #header>
        <span class="card-title">收款码设置</span>
        <span class="header-tip">提现时平台将通过此收款码向您转账</span>
      </template>
      <div class="qr-upload-row">
        <div class="qr-upload-item">
          <div class="qr-label"><el-icon style="color:#07c160"><ChatDotRound /></el-icon> 微信收款码</div>
          <div class="qr-preview-wrap" @click="triggerUpload('wechat')">
            <img v-if="qrForm.wechat_qr" :src="qrForm.wechat_qr" class="qr-img" />
            <div v-else class="qr-placeholder"><el-icon :size="32"><Upload /></el-icon><span>点击上传</span></div>
          </div>
          <el-button size="small" text @click="triggerUpload('wechat')">{{ qrForm.wechat_qr ? '重新上传' : '上传图片' }}</el-button>
        </div>
        <div class="qr-upload-item">
          <div class="qr-label"><el-icon style="color:#1677ff"><Money /></el-icon> 支付宝收款码</div>
          <div class="qr-preview-wrap" @click="triggerUpload('alipay')">
            <img v-if="qrForm.alipay_qr" :src="qrForm.alipay_qr" class="qr-img" />
            <div v-else class="qr-placeholder"><el-icon :size="32"><Upload /></el-icon><span>点击上传</span></div>
          </div>
          <el-button size="small" text @click="triggerUpload('alipay')">{{ qrForm.alipay_qr ? '重新上传' : '上传图片' }}</el-button>
        </div>
      </div>
      <input ref="fileInput" type="file" accept="image/*" style="display:none" @change="onFileChange" />
      <div style="margin-top:16px">
        <el-button type="primary" :loading="savingQR" @click="saveQR">保存收款码</el-button>
      </div>
    </el-card>

    <!-- 申请提现 -->
    <el-card class="panel" shadow="never" style="margin-top:16px">
      <template #header><span class="card-title">申请提现</span></template>
      <el-form :model="withdrawForm" label-width="100px" style="max-width:480px">
        <el-form-item label="提现积分">
          <el-input-number v-model="withdrawForm.amount" :min="1" :max="info.frozen_balance"
            :step="100" style="width:200px" />
          <span class="form-tip" style="margin-left:8px">冻结积分：{{ formatCredits(info.frozen_balance) }}</span>
        </el-form-item>
        <el-form-item label="收款方式">
          <el-radio-group v-model="withdrawForm.payment_type">
            <el-radio value="wechat">微信</el-radio>
            <el-radio value="alipay">支付宝</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="submittingWithdraw" @click="doWithdraw"
            :disabled="info.frozen_balance <= 0">提交提现申请</el-button>
        </el-form-item>
      </el-form>
      <el-alert v-if="info.frozen_balance <= 0" type="warning" :closable="false" show-icon
        title="暂无冻结积分可提现" style="max-width:480px" />
    </el-card>

    <!-- 提现记录 -->
    <el-card class="panel" shadow="never" style="margin-top:16px">
      <template #header><span class="card-title">提现记录</span></template>
      <el-table :data="history" v-loading="historyLoading" empty-text="暂无提现记录">
        <el-table-column label="申请时间" prop="created_at" :formatter="fmtTime" width="170" />
        <el-table-column label="积分数量" prop="amount" :formatter="(r)=>formatCredits(r.amount)" width="140" />
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
      </el-table>
      <div class="pager-wrap">
        <el-pagination v-model:current-page="historyPage" :page-size="20" :total="historyTotal"
          layout="total, prev, pager, next" @current-change="fetchHistory" />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { CopyDocument, Upload, ChatDotRound, Money } from '@element-plus/icons-vue'
import { userApi } from '@/api'

// ── 邀请信息 ──────────────────────────────────────────────────────────────
const loading = ref(true)
const converting = ref(false)
const convertAmount = ref(0)
const info = ref({ invite_code: '', invite_count: 0, frozen_balance: 0 })

const inviteLink = computed(() => `${window.location.origin}/register?ref=${info.value.invite_code}`)

function formatCredits(v) {
  if (!v) return '0'
  return (v / 1e6).toFixed(4) + ' 积分'
}

async function copyCode() {
  try { await navigator.clipboard.writeText(info.value.invite_code); ElMessage.success('邀请码已复制') }
  catch { ElMessage.error('复制失败') }
}
async function copyLink() {
  try { await navigator.clipboard.writeText(inviteLink.value); ElMessage.success('邀请链接已复制') }
  catch { ElMessage.error('复制失败') }
}

async function doConvert() {
  converting.value = true
  try {
    const res = await userApi.convertFrozen(convertAmount.value)
    ElMessage.success('成功解冻 ' + formatCredits(res.converted))
    convertAmount.value = 0
    await fetchInfo()
  } finally {
    converting.value = false
  }
}

async function fetchInfo() {
  try { info.value = await userApi.getInviteInfo() } catch {}
}

// ── 收款码 ──────────────────────────────────────────────────────────────────
const qrForm = ref({ wechat_qr: '', alipay_qr: '' })
const savingQR = ref(false)
const fileInput = ref(null)
let uploadTarget = ''

function triggerUpload(type) {
  uploadTarget = type
  fileInput.value.value = ''
  fileInput.value.click()
}

function onFileChange(e) {
  const file = e.target.files[0]
  if (!file) return
  if (file.size > 300 * 1024) { ElMessage.warning('图片请勿超过 300KB'); return }
  const reader = new FileReader()
  reader.onload = (ev) => {
    if (uploadTarget === 'wechat') qrForm.value.wechat_qr = ev.target.result
    else qrForm.value.alipay_qr = ev.target.result
  }
  reader.readAsDataURL(file)
}

async function saveQR() {
  savingQR.value = true
  try {
    await userApi.savePaymentQR({ wechat_qr: qrForm.value.wechat_qr, alipay_qr: qrForm.value.alipay_qr })
    ElMessage.success('收款码保存成功')
  } finally {
    savingQR.value = false
  }
}

// ── 提现申请 ────────────────────────────────────────────────────────────────
const withdrawForm = ref({ amount: 0, payment_type: 'wechat' })
const submittingWithdraw = ref(false)

async function doWithdraw() {
  if (!withdrawForm.value.amount || withdrawForm.value.amount <= 0) {
    ElMessage.warning('请输入提现积分数量'); return
  }
  const qr = withdrawForm.value.payment_type === 'wechat' ? qrForm.value.wechat_qr : qrForm.value.alipay_qr
  if (!qr) {
    ElMessage.warning('请先在上方保存' + (withdrawForm.value.payment_type === 'wechat' ? '微信' : '支付宝') + '收款码')
    return
  }
  await ElMessageBox.confirm(
    '确认申请提现 ' + formatCredits(withdrawForm.value.amount) + '？平台将通过' + (withdrawForm.value.payment_type === 'wechat' ? '微信' : '支付宝') + '向您转账。',
    '确认提现', { type: 'warning' }
  )
  submittingWithdraw.value = true
  try {
    await userApi.submitWithdraw(withdrawForm.value.amount, withdrawForm.value.payment_type)
    ElMessage.success('提现申请已提交，请等待平台审核')
    withdrawForm.value.amount = 0
    await fetchHistory()
  } finally {
    submittingWithdraw.value = false
  }
}

// ── 提现记录 ────────────────────────────────────────────────────────────────
const history = ref([])
const historyTotal = ref(0)
const historyPage = ref(1)
const historyLoading = ref(false)

function statusLabel(s) {
  return { pending: '待审核', approved: '已通过', rejected: '已拒绝' }[s] || s
}
function statusType(s) {
  return { pending: 'warning', approved: 'success', rejected: 'danger' }[s] || 'info'
}
function fmtTime(row, col, val) {
  return val ? new Date(val).toLocaleString('zh-CN') : '-'
}

async function fetchHistory() {
  historyLoading.value = true
  try {
    const res = await userApi.listWithdrawHistory(historyPage.value, 20)
    history.value = res.records || []
    historyTotal.value = res.total || 0
  } finally {
    historyLoading.value = false
  }
}

onMounted(async () => {
  await Promise.all([fetchInfo(), fetchHistory()])
  try {
    const qr = await userApi.getPaymentQR()
    qrForm.value.wechat_qr = qr.wechat_qr || ''
    qrForm.value.alipay_qr = qr.alipay_qr || ''
  } catch {}
  loading.value = false
})
</script>

<style scoped>
.invite-page { max-width: 760px; }
.panel { border-radius: 12px; }
.card-title { font-size: 16px; font-weight: 600; }
.header-tip { font-size: 12px; color: #86909c; margin-left: 8px; }
.loading-wrap { padding: 20px 0; }

.stat-row { display: flex; gap: 16px; flex-wrap: wrap; }
.stat-item {
  flex: 1; min-width: 140px; background: #f5f7fa;
  border-radius: 10px; padding: 16px 20px;
}
.stat-label { font-size: 13px; color: #86909c; margin-bottom: 8px; }
.stat-value { font-size: 22px; font-weight: 700; color: #1d2129; display: flex; align-items: center; gap: 6px; }
.stat-value.highlight { color: #2563eb; }
.code-value { font-size: 18px; letter-spacing: .05em; }

.section-label { font-size: 14px; font-weight: 600; margin-bottom: 10px; }
.link-section { }
.link-row { display: flex; gap: 8px; }
.link-input { flex: 1; }
.convert-section { }
.convert-row { display: flex; gap: 12px; align-items: center; }
.form-tip { font-size: 12px; color: #86909c; margin-top: 6px; }
.rule-list { margin: 4px 0 0; padding-left: 18px; font-size: 13px; line-height: 1.8; }

.qr-upload-row { display: flex; gap: 40px; flex-wrap: wrap; }
.qr-upload-item { display: flex; flex-direction: column; align-items: center; gap: 10px; }
.qr-label { font-size: 14px; font-weight: 500; display: flex; align-items: center; gap: 4px; }
.qr-preview-wrap {
  width: 140px; height: 140px; border: 2px dashed #d0d5dd; border-radius: 10px;
  cursor: pointer; overflow: hidden; display: flex; align-items: center; justify-content: center;
  transition: border-color .2s;
}
.qr-preview-wrap:hover { border-color: #2563eb; }
.qr-img { width: 100%; height: 100%; object-fit: contain; }
.qr-placeholder { display: flex; flex-direction: column; align-items: center; gap: 6px; color: #86909c; font-size: 12px; }

.pager-wrap { display: flex; justify-content: flex-end; margin-top: 16px; }
</style>
