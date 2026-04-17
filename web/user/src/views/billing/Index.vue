<template>
  <div class="billing-page">
    <!-- ───────── 充值积分 ───────── -->
    <template v-if="route.path === '/recharge'">
      <!-- 页面标题 -->
      <div class="recharge-header">
        <div class="page-title">积分充值</div>
      </div>

      <!-- 余额 + 套餐 + 支付方式 -->
      <div class="form-contain">
        <!-- 当前积分 -->
        <div class="section-title">当前积分</div>
        <div class="balance-display">
          <span class="balance-big">{{ (store.balance / 1e6).toFixed(2) }}</span>
          <span class="balance-unit">积分</span>
        </div>
        <div class="balance-tip">
          <svg width="16" height="16" viewBox="0 0 16 16" fill="none" style="flex-shrink:0"><circle cx="8" cy="8" r="7" stroke="currentColor" stroke-width="1.5"/><path d="M8 5v4M8 11v.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
          积分永不过期，随时可用
        </div>

        <!-- 套餐选择 -->
        <div class="plans-grid" v-if="site.plans && site.plans.length">
          <div
            class="plan-card"
            v-for="(plan, i) in site.plans"
            :key="i"
            :class="{ selected: selectedPlan === i }"
            @click="selectedPlan = i; selectedAmount = plan.amount"
          >
            <div v-if="selectedPlan === i" class="plan-check">✓</div>
            <div class="plan-credits">{{ plan.credits }}积分<span v-if="plan.bonus">（赠送{{ plan.bonus }}积分）</span></div>
            <div class="plan-price">
              <span class="price-symbol">￥</span>
              <span class="price-num">{{ plan.amount }}</span>
              <span class="price-origin" v-if="plan.origin_amount">原价￥{{ plan.origin_amount }}</span>
            </div>
            <div class="plan-desc" v-if="plan.desc">{{ plan.desc }}</div>
          </div>
        </div>

        <!-- 自定义金额 -->
        <div class="custom-amount-row" v-if="!site.plans || !site.plans.length || site.allowCustom">
          <div class="section-title" style="margin-top:16px">充值金额（元）</div>
          <el-input-number
            v-model="selectedAmount"
            :min="1"
            :max="10000"
            :precision="2"
            style="width:240px;margin-top:8px"
            @change="selectedPlan = -1"
          />
        </div>

        <!-- 支付方式 -->
        <div class="section-title" style="margin-top:20px">支付方式</div>
        <div class="pay-methods">
          <div
            class="pay-method-card"
            :class="{ 'pay-selected': payMethod === 'wechat' }"
            @click="payMethod = 'wechat'"
            v-if="site.epayEnabled || site.payApplyEnabled"
          >
            <div class="pay-logo-wrap" style="background:#07C160;border-radius:6px;width:30px;height:30px;display:flex;align-items:center;justify-content:center;flex-shrink:0">
              <svg role="img" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" width="20" height="20" fill="white"><title>WeChat</title><path d="M8.691 2.188C3.891 2.188 0 5.476 0 9.53c0 2.212 1.17 4.203 3.002 5.55a.59.59 0 0 1 .213.665l-.39 1.48c-.019.07-.048.141-.048.213 0 .163.13.295.29.295a.326.326 0 0 0 .167-.054l1.903-1.114a.864.864 0 0 1 .717-.098 10.16 10.16 0 0 0 2.837.403c.276 0 .543-.027.811-.05-.857-2.578.157-4.972 1.932-6.446 1.703-1.415 3.882-1.98 5.853-1.838-.576-3.583-4.196-6.348-8.596-6.348zM5.785 5.991c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 0 1-1.162 1.178A1.17 1.17 0 0 1 4.623 7.17c0-.651.52-1.18 1.162-1.18zm5.813 0c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 0 1-1.162 1.178 1.17 1.17 0 0 1-1.162-1.178c0-.651.52-1.18 1.162-1.18zm5.34 2.867c-1.797-.052-3.746.512-5.28 1.786-1.72 1.428-2.687 3.72-1.78 6.22.942 2.453 3.666 4.229 6.884 4.229.826 0 1.622-.12 2.361-.336a.722.722 0 0 1 .598.082l1.584.926a.272.272 0 0 0 .14.047c.134 0 .24-.111.24-.247 0-.06-.023-.12-.038-.177l-.327-1.233a.582.582 0 0 1-.023-.156.49.49 0 0 1 .201-.398C23.024 18.48 24 16.82 24 14.98c0-3.21-2.931-5.837-6.656-6.088V8.89c-.135-.01-.27-.027-.407-.03zm-2.53 3.274c.535 0 .969.44.969.982a.976.976 0 0 1-.969.983.976.976 0 0 1-.969-.983c0-.542.434-.982.97-.982zm4.844 0c.535 0 .969.44.969.982a.976.976 0 0 1-.969.983.976.976 0 0 1-.969-.983c0-.542.434-.982.969-.982z"/></svg>
            </div>
            <span>微信支付</span>
          </div>
          <div
            class="pay-method-card"
            :class="{ 'pay-selected': payMethod === 'alipay' }"
            @click="payMethod = 'alipay'"
            v-if="site.epayEnabled || site.payApplyEnabled"
          >
            <div class="pay-logo-wrap" style="background:#1677FF;border-radius:6px;width:30px;height:30px;display:flex;align-items:center;justify-content:center;flex-shrink:0">
              <svg role="img" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" width="20" height="20" fill="white"><title>Alipay</title><path d="M19.695 15.07c3.426 1.158 4.203 1.22 4.203 1.22V3.846c0-2.124-1.705-3.845-3.81-3.845H3.914C1.808.001.102 1.722.102 3.846v16.31c0 2.123 1.706 3.845 3.813 3.845h16.173c2.105 0 3.81-1.722 3.81-3.845v-.157s-6.19-2.602-9.315-4.119c-2.096 2.602-4.8 4.181-7.607 4.181-4.75 0-6.361-4.19-4.112-6.949.49-.602 1.324-1.175 2.617-1.497 2.025-.502 5.247.313 8.266 1.317a16.796 16.796 0 0 0 1.341-3.302H5.781v-.952h4.799V6.975H4.77v-.953h5.81V3.591s0-.409.411-.409h2.347v2.84h5.744v.951h-5.744v1.704h4.69a19.453 19.453 0 0 1-1.986 5.06c1.424.52 2.702 1.011 3.654 1.333m-13.81-2.032c-.596.06-1.71.325-2.321.869-1.83 1.608-.735 4.55 2.968 4.55 2.151 0 4.301-1.388 5.99-3.61-2.403-1.182-4.438-2.028-6.637-1.809"/></svg>
            </div>
            <span>支付宝支付</span>
          </div>
        </div>

        <!-- 支付按钮 -->
        <div class="pay-btn-row">
          <el-button
            type="primary"
            size="large"
            style="border-radius:44px;font-size:16px;width:300px;height:44px"
            :loading="paying || payApplying"
            @click="doEpay"
            v-if="site.epayEnabled || site.payApplyEnabled"
          >
            立即支付 ￥{{ (selectedAmount || 0).toFixed(2) }}
          </el-button>
        </div>
      </div>

      <el-card style="margin-top:16px">
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
        <template #header>
          <div style="display:flex;align-items:center;justify-content:space-between">
            <span>余额流水</span>
            <div style="display:flex;gap:8px">
              <el-input
                v-model="txTaskIdFilter"
                placeholder="按任务 ID 查询"
                clearable
                style="width:160px"
                size="small"
                @keyup.enter="fetchTx(1)"
                @clear="fetchTx(1)"
              />
              <el-button size="small" type="primary" @click="fetchTx(1)">查询</el-button>
            </div>
          </div>
        </template>
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
          <el-table-column label="关联任务" width="120">
            <template #default="{ row }">
              <span v-if="row.metrics?.task_id" style="color:#409eff;font-size:12px;font-family:monospace">
                #{{ row.metrics.task_id }}
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
          <el-table-column label="到账金额" width="140">
            <template #default="{ row }">
              <span style="color:#67c23a">+¥{{ (row.credits / 1e6).toFixed(2) }}</span>
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



    <!-- 中台支付内嵌弹窗 -->
    <el-dialog v-model="showPayFrame" title="扫码支付" width="360px" :close-on-click-modal="false" @close="onPayFrameClose">
      <div style="text-align:center">
        <div style="color:#606266;font-size:13px;margin-bottom:16px">
          请使用手机扫码完成支付，支付成功后余额将自动到账
        </div>
        <canvas ref="qrcodeCanvas" style="border-radius:8px;max-width:100%" />
        <div style="margin-top:12px;color:#909399;font-size:12px">如扫码无效，可点击下方按鈕在新窗口打开</div>
      </div>
      <template #footer>
        <el-button @click="showPayFrame = false">我已完成支付</el-button>
        <el-button type="primary" @click="openPayInTab">在新窗口打开</el-button>
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
import { ref, onMounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'

import QRCode from 'qrcode'
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
const txTaskIdFilter = ref('')

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

const paying = ref(false)
const payApplying = ref(false)

// New recharge UI
const selectedPlan = ref(0)
const selectedAmount = ref(10)
const payMethod = ref('wechat')

async function doEpay() {
  if (!selectedAmount.value || selectedAmount.value < 0.01) {
    return ElMessage.warning('请输入有效的充值金额')
  }
  const type = payMethod.value === 'wechat' ? 'wxpay' : 'alipay'
  const payFlat = payMethod.value === 'wechat' ? 1 : 2

  if (site.payApplyEnabled) {
    payApplying.value = true
    try {
      const isMobile = /Mobi|Android/i.test(navigator.userAgent)
      const payFrom = isMobile ? (payFlat === 1 ? 'wapwx' : 'wap') : 'pc'
      const res = await payApi.createPayApplyOrder({
        amount: selectedAmount.value,
        pay_flat: payFlat,
        pay_from: payFrom,
      })
      if (res.pay_url) {
        payFrameURL.value = res.pay_url
        currentOutTradeNo.value = res.out_trade_no
        showPayFrame.value = true
        await nextTick()
        if (qrcodeCanvas.value) {
          QRCode.toCanvas(qrcodeCanvas.value, res.pay_url, { width: 280, margin: 2 })
        }
        startPayPolling(res.out_trade_no)
      }
    } finally {
      payApplying.value = false
    }
  } else if (site.epayEnabled) {
    paying.value = true
    try {
      const res = await payApi.createEpayOrder({ amount: selectedAmount.value, type })
      if (res.pay_url) {
        window.open(res.pay_url, '_blank')
        ElMessage.info('请在新窗口中完成支付，支付成功后余额将自动到账')
      }
    } finally {
      paying.value = false
    }
  }
}

const showPayFrame = ref(false)
const payFrameURL = ref('')
const qrcodeCanvas = ref(null)
const currentOutTradeNo = ref('')
let payPollTimer = null

onMounted(() => {
  fetchTx()
  fetchOrders()
  site.fetchSettings()
})

async function fetchTx(p = txPage.value) {
  txPage.value = p
  const res = await userApi.getTransactions(p, 20, txTaskIdFilter.value.trim())
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

function stopPayPolling() {
  if (payPollTimer) {
    clearInterval(payPollTimer)
    payPollTimer = null
  }
}

function startPayPolling(outTradeNo) {
  stopPayPolling()
  payPollTimer = setInterval(async () => {
    try {
      const res = await payApi.getOrderStatus(outTradeNo)
      if (res.status === 'paid') {
        stopPayPolling()
        showPayFrame.value = false
        ElMessage.success('支付成功，余额已到账！')
        store.fetchBalance()
        fetchTx(1)
        fetchOrders(1)
      }
    } catch {
      // 忽略轮询期间的网络错误
    }
  }, 3000)
}

function onPayFrameClose() {
  stopPayPolling()
  payFrameURL.value = ''
  currentOutTradeNo.value = ''
  // 弹窗关闭后刷新余额和流水，确保到账显示最新
  store.fetchBalance()
  fetchTx(1)
  fetchOrders(1)
}

function openPayInTab() {
  window.open(payFrameURL.value, '_blank')
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

/* 充值页新样式 */
.recharge-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
}
.page-title {
  font-size: 24px;
  font-weight: 600;
  color: rgb(26, 27, 28);
  line-height: 32px;
}
.exchange-rate {
  font-size: 24px;
  font-weight: bold;
  color: rgb(22, 93, 255);
}
.form-contain {
  background: white;
  padding: 20px;
  border-radius: 8px;
  border: 1px solid #e5e6eb;
  display: flex;
  flex-direction: column;
}
.section-title {
  font-size: 18px;
  font-weight: bold;
  margin-bottom: 15px;
  color: rgb(15, 23, 42);
}
.balance-display {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-bottom: 8px;
}
.balance-big {
  font-size: 40px;
  font-weight: bold;
  color: rgb(15, 23, 42);
  font-family: 'Source Han Sans CN', sans-serif;
}
.balance-unit {
  font-size: 16px;
  color: rgb(100, 116, 139);
}
.balance-tip {
  display: flex;
  align-items: center;
  gap: 4px;
  color: rgb(100, 116, 139);
  font-size: 16px;
  margin: 15px 0;
}

/* 套餐卡片 */
.plans-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 15px;
  margin-bottom: 8px;
}
.plan-card {
  border: 1px solid #e5e6eb;
  border-radius: 8px;
  padding: 16px;
  cursor: pointer;
  position: relative;
  background: white;
  transition: border-color .15s, box-shadow .15s;
}
.plan-card:hover {
  border-color: #165dff;
  box-shadow: 0 2px 8px rgba(22,93,255,.12);
}
.plan-card.selected {
  border-color: #165dff;
  background: rgb(232, 243, 255);
}
.plan-check {
  position: absolute;
  right: -1px;
  top: -1px;
  width: 22px;
  height: 22px;
  background: #165dff;
  color: white;
  font-size: 11px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 0 8px 0 6px;
}
.plan-credits {
  font-size: 16px;
  font-weight: bold;
  margin-bottom: 6px;
  color: #1d2129;
}
.plan-price {
  display: flex;
  align-items: baseline;
  gap: 2px;
  font-size: 28px;
  font-weight: 600;
  line-height: 40px;
  color: #1d2129;
}
.price-symbol { font-size: 18px; }
.price-num { font-size: 28px; }
.price-origin {
  font-size: 13px;
  font-weight: 400;
  color: rgb(178, 178, 178);
  text-decoration: line-through;
  margin-left: 4px;
}
.plan-desc {
  font-size: 15px;
  color: rgb(153, 153, 153);
  margin-top: 4px;
}

/* 支付方式 */
.pay-methods {
  display: flex;
  gap: 16px;
  margin: 10px 0 16px;
}
.pay-method-card {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 14px 24px;
  border: 1px solid #e5e6eb;
  border-radius: 8px;
  cursor: pointer;
  background: rgb(248, 248, 248);
  font-size: 16px;
  position: relative;
  transition: border-color .15s, background .15s;
}
.pay-method-card:hover { border-color: #165dff; }
.pay-method-card.pay-selected {
  border-color: #165dff;
  background: rgb(232, 243, 255);
}

/* 支付按钮 */
.pay-btn-row {
  display: flex;
  align-items: center;
  gap: 16px;
  padding-top: 12px;
  border-top: 1px solid #f0f1f5;
  margin-top: 16px;
}

.amt-neg { color: #f56c6c }
.amt-pos { color: #67c23a }
.credits-preview { color: #67c23a; font-weight: 600; font-size: 1rem }
</style>
