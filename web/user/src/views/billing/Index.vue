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
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { userApi } from '@/api'

const store = useUserStore()
const txList = ref([])
const page = ref(1)
const total = ref(0)
const showRedeem = ref(false)
const redeemCode = ref('')
const redeeming = ref(false)

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

