<template>
  <div class="billing-page">
    <el-card class="hero-card">
      <div class="hero-row">
        <div>
          <div class="eyebrow">Wallet & Billing</div>
          <h3>余额与消费流水</h3>
          <p>余额不过期，所有充值、预扣、结算与退款记录都可以在这里统一查看。</p>
        </div>
        <div class="balance-box">
          <span>当前余额</span>
          <strong>¥{{ (store.balance / 1e6).toFixed(4) }}</strong>
          <small>{{ store.balance.toLocaleString() }} credits</small>
          <el-button type="primary" @click="showRecharge = true">充值</el-button>
        </div>
      </div>
    </el-card>

    <!-- 账单明细 -->
    <el-card>
      <template #header>消费记录</template>
      <el-table :data="txList" stripe>
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.type === 'charge' ? 'danger' : 'success'" size="small">
              {{ txTypeLabel(row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="金额" width="140">
          <template #default="{ row }">
            <span :style="{ color: row.type === 'charge' ? '#f56c6c' : '#67c23a' }">
              {{ row.type === 'charge' ? '-' : '+' }}{{ (Math.abs(row.credits) / 1e6).toFixed(6) }} ¥
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="时间" :formatter="fmtTime" />
        <el-table-column prop="note" label="备注" show-overflow-tooltip />
      </el-table>
      <el-pagination
        v-model:current-page="page"
        :page-size="20"
        :total="total"
        style="margin-top:16px"
        @current-change="fetchTx"
      />
    </el-card>

    <!-- 充值弹窗（占位，实际需对接支付）-->
    <el-dialog v-model="showRecharge" title="充值" width="400px">
      <el-alert type="info" :closable="false">
        请联系管理员手动充值或对接支付渠道
      </el-alert>
      <template #footer>
        <el-button @click="showRecharge = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import { userApi } from '@/api'

const store = useUserStore()
const txList = ref([])
const page = ref(1)
const total = ref(0)
const showRecharge = ref(false)

onMounted(fetchTx)

async function fetchTx() {
  const res = await userApi.getTransactions(page.value, 20)
  txList.value = res.transactions ?? []
  total.value = res.total ?? 0
}

const txTypeLabel = (t) => ({
  charge: '扣费', refund: '退款', recharge: '充值', hold: '预扣', settle: '结算'
}[t] ?? t)

function fmtTime(row, col, val) {
  return val ? new Date(val).toLocaleString('zh-CN') : '-'
}
</script>

<style scoped>
.billing-page {
  max-width: 1320px;
}
.hero-card {
  margin-bottom: 20px;
}
.hero-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 20px;
}
.eyebrow {
  color: #1e66ff;
  font-size: .82rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: .08em;
}
.hero-row h3 {
  margin: 8px 0 10px;
  font-size: 1.55rem;
}
.hero-row p {
  margin: 0;
  color: #617086;
}
.balance-box {
  min-width: 240px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 18px;
  border-radius: 18px;
  border: 1px solid #dce7fb;
  background: linear-gradient(180deg, #f7fbff, #eef6ff);
}
.balance-box span,
.balance-box small {
  color: #69809e;
}
.balance-box strong {
  font-size: 1.9rem;
}

@media (max-width: 900px) {
  .hero-row {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
