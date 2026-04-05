<template>
  <div class="billing-page">
    <!-- 余额卡片 -->
    <el-row :gutter="18" style="margin-bottom:18px">
      <el-col :span="10">
        <div class="balance-card">
          <div class="balance-label">当前余额</div>
          <div class="balance-val">¥{{ (store.balance / 1e6).toFixed(4) }}</div>
          <div class="balance-sub">{{ store.balance.toLocaleString() }} credits</div>
          <el-button type="primary" style="margin-top:16px" @click="showRedeem = true">
            兑换卡密充值
          </el-button>
        </div>
      </el-col>
    </el-row>

    <!-- 账户安全 -->
    <el-card style="margin-bottom:18px">
      <template #header>账户安全</template>
      <el-descriptions :column="1" border>
        <el-descriptions-item label="用户名">{{ store.username || '—' }}</el-descriptions-item>
        <el-descriptions-item label="绑定邮箱">
          <span v-if="store.email" style="color:#67c23a">{{ store.email }}</span>
          <template v-else>
            <span style="color:#909399;margin-right:12px">未绑定（绑定后可找回密码）</span>
            <el-button type="primary" size="small" @click="showBindEmail = true">绑定邮箱</el-button>
          </template>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <!-- 余额流水 -->
    <el-card>
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
        v-model:current-page="page"
        :page-size="20"
        :total="total"
        style="margin-top:16px"
        @current-change="fetchTx"
      />
    </el-card>

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
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { authApi, userApi } from '@/api'

const store = useUserStore()
const txList = ref([])
const page = ref(1)
const total = ref(0)
const showRedeem = ref(false)
const redeemCode = ref('')
const redeeming = ref(false)

// 绑定邮箱
const showBindEmail = ref(false)
const bindForm = ref({ email: '', code: '' })
const codeCooldown = ref(0)
const binding = ref(false)
let cooldownTimer = null

onMounted(fetchTx)

async function fetchTx(p = page.value) {
  page.value = p
  const res = await userApi.getTransactions(p, 20)
  txList.value = res.transactions ?? []
  total.value = res.total ?? 0
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
  } finally {
    redeeming.value = false
  }
}

const txTypeLabel = (t) => ({ charge: '扣费', refund: '退款', recharge: '充值', hold: '预扣', settle: '结算' }[t] ?? t)
const txTagType = (t) => ({ charge: 'danger', hold: 'warning', settle: 'info', refund: 'success', recharge: 'success' }[t] ?? '')
const txSign = (t) => (['charge', 'hold', 'settle'].includes(t) ? '-' : '+')
const txAmtClass = (t) => (['charge', 'hold', 'settle'].includes(t) ? 'amt-neg' : 'amt-pos')

function fmtTime(row, col, val) {
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
}
.balance-label { font-size: .84rem; opacity: .72 }
.balance-val { font-size: 2.4rem; font-weight: 700; margin: 6px 0 2px }
.balance-sub { font-size: .82rem; opacity: .6 }
.amt-neg { color: #f56c6c }
.amt-pos { color: #67c23a }
</style>

