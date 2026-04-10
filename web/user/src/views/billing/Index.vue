<template>
  <div class="billing-page">
    <!-- ───────── 充值积分 ───────── -->
    <template v-if="route.path === '/recharge'">
      <div class="balance-card">
        <div class="balance-label">当前余额</div>
        <div class="balance-val">¥{{ (store.balance / 1e6).toFixed(4) }}</div>
        <div class="balance-sub">{{ store.balance.toLocaleString() }} credits</div>
        <div class="balance-actions">
          <el-button type="primary" @click="showRedeem = true">兑换卡密</el-button>
          <el-button v-if="site.epayEnabled" type="success" @click="showEpay = true">
            <el-icon><CreditCard /></el-icon>
            在线充值
          </el-button>
        </div>
      </div>

      <el-card style="margin-top:18px">
        <template #header>账户信息</template>
        <el-descriptions :column="1" border>
          <el-descriptions-item label="用户名">{{ store.username || '—' }}</el-descriptions-item>
          <el-descriptions-item label="定价分组">
            <el-tag v-if="store.group" type="warning" effect="light">{{ store.group }}</el-tag>
            <span v-else style="color:#909399">默认</span>
            <span style="color:#c0c4cc;font-size:12px;margin-left:8px">影响模型调用的实际计费价格</span>
          </el-descriptions-item>
          <el-descriptions-item label="绑定邮箱">
            <span v-if="store.email" style="color:#67c23a">{{ store.email }}</span>
            <template v-else>
              <span style="color:#909399;margin-right:12px">未绑定（绑定后可找回密码）</span>
              <el-button type="primary" size="small" @click="showBindEmail = true">绑定邮箱</el-button>
            </template>
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <el-card style="margin-top:18px">
        <template #header>余额流水</template>
        <el-table :data="txList" stripe>
          <el-table-column prop="type" label="类型" width="100">
            <template #default="{ row }">
              <el-tag :type="txTagType(row.type)" size="small">{{ txTypeLabel(row.type) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="金额" width="160">
            <template #default="{ row }">
              <span :class="txAmtClass(row.type)">
                {{ txSign(row.type) }}¥{{ (Math.abs(row.credits) / 1e6).toFixed(6) }}
              </span>
            </template>
          </el-table-column>
          <el-table-column label="操作后余额" width="160">
            <template #default="{ row }">
              <span v-if="row.balance_after" style="color:#617086;font-size:12px">
                ¥{{ (row.balance_after / 1e6).toFixed(4) }}
              </span>
              <span v-else style="color:#ccc;font-size:12px">—</span>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="时间" :formatter="fmtTime" />
        </el-table>
        <el-pagination
          v-model:current-page="txPage"
          :page-size="20"
          :total="txTotal"
          style="margin-top:16px"
          @current-change="fetchTx"
        />
      </el-card>
    </template>

    <!-- ───────── 我的订单 ───────── -->
    <template v-else>
      <el-card>
        <template #header>充值订单记录</template>
        <el-table :data="orderList" stripe>
          <el-table-column prop="out_trade_no" label="订单号" min-width="200" />
          <el-table-column label="充值金额" width="110">
            <template #default="{ row }">
              <span style="color:#165dff;font-weight:600">¥{{ row.amount.toFixed(2) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="到账积分" width="140">
            <template #default="{ row }">
              <span style="color:#67c23a">+{{ row.credits.toLocaleString() }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="90">
            <template #default="{ row }">
              <el-tag :type="orderStatusType(row.status)" size="small">
                {{ orderStatusLabel(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="支付时间" min-width="160">
            <template #default="{ row }">
              {{ row.paid_at ? fmtDate(row.paid_at) : '—' }}
            </template>
          </el-table-column>
          <el-table-column label="创建时间" min-width="160">
            <template #default="{ row }">{{ fmtDate(row.created_at) }}</template>
          </el-table-column>
        </el-table>
        <el-pagination
          v-model:current-page="orderPage"
          :page-size="20"
          :total="orderTotal"
          style="margin-top:16px"
          @current-change="fetchOrders"
        />
      </el-card>
    </template>

    <!-- 兑换卡密弹窗 -->
    <el-dialog v-model="showRedeem" title="兑换卡密" width="400px" @close="redeemCode = ''">
      <el-form @submit.prevent="doRedeem">
        <el-form-item label="卡密">
          <el-input
            v-model="redeemCode"
            placeholder="FANAPI-XXXXXXXXXXXXXXXX"
            clearable
            @keyup.enter="doRedeem"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRedeem = false">取消</el-button>
        <el-button type="primary" :loading="redeeming" @click="doRedeem">立即兑换</el-button>
      </template>
    </el-dialog>

    <!-- 在线充值弹窗（易支付） -->
    <el-dialog v-model="showEpay" title="在线充值" width="420px" @close="resetEpayForm">
      <el-alert type="info" :closable="false" show-icon style="margin-bottom:16px">
        <template #title>充值后余额将自动到账，1元 = 1,000,000 积分</template>
      </el-alert>
      <el-form :model="epayForm" label-width="90px">
        <el-form-item label="充值金额">
          <el-input-number
            v-model="epayForm.amount"
            :min="1"
            :precision="2"
            :step="10"
            style="width:100%"
            placeholder="请输入充值金额（元）"
          />
        </el-form-item>
        <el-form-item label="支付方式">
          <el-radio-group v-model="epayForm.type">
            <el-radio value="alipay">
              <el-icon style="color:#165dff;vertical-align:middle"><Wallet /></el-icon>
              支付宝
            </el-radio>
            <el-radio value="wxpay">
              <el-icon style="color:#07c160;vertical-align:middle"><ChatDotRound /></el-icon>
              微信支付
            </el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="实际积分">
          <span class="credits-preview">+{{ (epayForm.amount * 1e6).toLocaleString() }} credits</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEpay = false">取消</el-button>
        <el-button type="success" :loading="paying" @click="doEpayPay">前往支付</el-button>
      </template>
    </el-dialog>

    <!-- 绑定邮箱弹窗 -->
    <el-dialog v-model="showBindEmail" title="绑定邮箱" width="420px" @close="resetBindForm">
      <el-form label-width="80px">
        <el-form-item label="邮箱">
          <el-input v-model="bindForm.email" placeholder="请输入邮箱地址" />
        </el-form-item>
        <el-form-item label="验证码">
          <div style="display:flex;gap:8px">
            <el-input v-model="bindForm.code" placeholder="6位验证码" />
            <el-button :disabled="codeCooldown > 0" @click="sendBindCode" style="flex-shrink:0">
              {{ codeCooldown > 0 ? `${codeCooldown}s` : '获取验证码' }}
            </el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showBindEmail = false">取消</el-button>
        <el-button type="primary" :loading="binding" @click="doBindEmail">确认绑定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { CreditCard, Wallet, ChatDotRound } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { useSiteStore } from '@/stores/site'
import { authApi, userApi, payApi } from '@/api'

const route = useRoute()
const store = useUserStore()
const site = useSiteStore()

// 余额流水
const txList = ref([])
const txPage = ref(1)
const txTotal = ref(0)

// 充值订单
const orderList = ref([])
const orderPage = ref(1)
const orderTotal = ref(0)

// 弹窗
const showRedeem = ref(false)
const redeemCode = ref('')
const redeeming = ref(false)

const showBindEmail = ref(false)
const bindForm = ref({ email: '', code: '' })
const codeCooldown = ref(0)
const binding = ref(false)
let cooldownTimer = null

const showEpay = ref(false)
const paying = ref(false)
const epayForm = reactive({ amount: 10, type: 'alipay' })

onMounted(() => {
  fetchTx()
  fetchOrders()
  site.fetchSettings()
})

async function fetchTx(p = txPage.value) {
  txPage.value = p
  const res = await userApi.getTransactions(p, 20)
  txList.value = res.transactions ?? []
  txTotal.value = res.total ?? 0
}

async function fetchOrders(p = orderPage.value) {
  orderPage.value = p
  const res = await payApi.listOrders(p, 20)
  orderList.value = res.orders ?? []
  orderTotal.value = res.total ?? 0
}

async function doRedeem() {
  const code = redeemCode.value.trim()
  if (!code) return
  redeeming.value = true
  try {
    const res = await userApi.redeemCard(code)
    ElMessage.success(res.message ?? '兑换成功')
    showRedeem.value = false
    redeemCode.value = ''
    store.fetchBalance()
    fetchTx(1)
    fetchOrders(1)
  } finally {
    redeeming.value = false
  }
}

function resetEpayForm() {
  epayForm.amount = 10
  epayForm.type = 'alipay'
}

async function doEpayPay() {
  if (!epayForm.amount || epayForm.amount < 0.01) {
    return ElMessage.warning('请输入有效的充值金额')
  }
  paying.value = true
  try {
    const res = await payApi.createEpayOrder({ amount: epayForm.amount, type: epayForm.type })
    if (res.pay_url) {
      window.open(res.pay_url, '_blank')
      showEpay.value = false
      ElMessage.info('请在新窗口中完成支付，支付成功后余额将自动到账')
    }
  } finally {
    paying.value = false
  }
}

const txTypeLabel = (t) => ({ charge: '扣费', refund: '退款', recharge: '充值', hold: '预扣', settle: '结算' }[t] ?? t)
const txTagType = (t) => ({ charge: 'danger', hold: 'warning', settle: 'info', refund: 'success', recharge: 'success' }[t] ?? '')
const txSign = (t) => (['charge', 'hold', 'settle'].includes(t) ? '-' : '+')
const txAmtClass = (t) => (['charge', 'hold', 'settle'].includes(t) ? 'amt-neg' : 'amt-pos')

const orderStatusLabel = (s) => ({ pending: '待支付', paid: '已支付', failed: '已失败' }[s] ?? s)
const orderStatusType = (s) => ({ pending: 'warning', paid: 'success', failed: 'danger' }[s] ?? 'info')

function fmtTime(row, col, val) {
  return val ? new Date(val).toLocaleString('zh-CN', { hour12: false }) : '—'
}
function fmtDate(val) {
  return val ? new Date(val).toLocaleString('zh-CN', { hour12: false }) : '—'
}

function resetBindForm() {
  bindForm.value = { email: '', code: '' }
  codeCooldown.value = 0
  clearInterval(cooldownTimer)
}

async function sendBindCode() {
  const email = bindForm.value.email.trim()
  if (!email) return ElMessage.warning('请先输入邮箱')
  try {
    await authApi.sendCode(email)
    ElMessage.success('验证码已发送，请查收邮件')
    codeCooldown.value = 60
    cooldownTimer = setInterval(() => {
      if (--codeCooldown.value <= 0) clearInterval(cooldownTimer)
    }, 1000)
  } catch {
    // error handled by http interceptor
  }
}

async function doBindEmail() {
  const { email, code } = bindForm.value
  if (!email.trim() || !code.trim()) return ElMessage.warning('请填写邮箱和验证码')
  binding.value = true
  try {
    await userApi.bindEmail({ email: email.trim(), code: code.trim() })
    store.setEmail(email.trim())
    ElMessage.success('邮箱绑定成功')
    showBindEmail.value = false
  } finally {
    binding.value = false
  }
}
</script>

<style scoped>
.billing-page { max-width: 960px }
.balance-card {
  background: linear-gradient(135deg, #0b1a3e, #163879);
  border-radius: 18px;
  padding: 28px 32px;
  color: #fff;
  display: inline-block;
  min-width: 320px;
}
.balance-label { font-size: .84rem; opacity: .72 }
.balance-val { font-size: 2.4rem; font-weight: 700; margin: 6px 0 2px }
.balance-sub { font-size: .82rem; opacity: .6 }
.balance-actions { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 16px }
.amt-neg { color: #f56c6c }
.amt-pos { color: #67c23a }
.credits-preview { color: #67c23a; font-weight: 600; font-size: 1rem }
</style>
