<template>
  <div class="billing-page">
    <el-card class="hero-card">
      <div class="hero-row">
        <div>
          <div class="eyebrow">Profit Analytics</div>
          <h3>账单流水与利润统计</h3>
          <p>按时间范围查看平台收入、成本和利润，支持运营复盘与对账。</p>
        </div>
        <div class="filter-panel">
          <el-date-picker
            v-model="range"
            type="datetimerange"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="width: 360px"
          />
          <el-button type="primary" @click="fetchTx">查询</el-button>
        </div>
      </div>
    </el-card>

    <el-row :gutter="16" class="summary-row">
      <el-col :xs="24" :sm="8"><el-card class="summary-card"><div class="label">收入</div><div class="value">¥{{ toYuan(summary.revenue) }}</div></el-card></el-col>
      <el-col :xs="24" :sm="8"><el-card class="summary-card"><div class="label">成本</div><div class="value">¥{{ toYuan(summary.cost) }}</div></el-card></el-col>
      <el-col :xs="24" :sm="8"><el-card class="summary-card profit"><div class="label">利润</div><div class="value">¥{{ toYuan(summary.profit) }}</div></el-card></el-col>
    </el-row>

    <el-card>
    <el-table :data="txList" stripe border>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="user_id" label="用户 ID" width="80" />
      <el-table-column prop="type" label="类型" width="90">
        <template #default="{ row }">
          <el-tag :type="typeColor(row.type)" size="small">{{ txTypeLabel(row.type) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="金额" width="160">
        <template #default="{ row }">
          <span :style="{ color: row.type === 'charge' ? '#f56c6c' : '#67c23a' }">
            {{ row.type === 'charge' ? '-' : '+' }}{{ Math.abs(row.credits).toLocaleString() }} cr
          </span>
        </template>
      </el-table-column>
      <el-table-column label="成本" width="140">
        <template #default="{ row }">
          <span>{{ (row.cost ?? 0).toLocaleString() }} cr</span>
        </template>
      </el-table-column>
      <el-table-column label="利润" width="140">
        <template #default="{ row }">
          <span :style="{ color: profitOf(row) >= 0 ? '#155eef' : '#f56c6c' }">
            {{ profitOf(row).toLocaleString() }} cr
          </span>
        </template>
      </el-table-column>
      <el-table-column prop="channel_id" label="渠道 ID" width="80" />
      <el-table-column prop="corr_id" label="关联 ID" show-overflow-tooltip />
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
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { txApi } from '@/api'

const txList = ref([])
const page = ref(1)
const total = ref(0)
const range = ref([])
const summary = ref({ revenue: 0, cost: 0, profit: 0, transaction_count: 0 })

onMounted(fetchTx)

async function fetchTx() {
  const params = { page: page.value, size: 20 }
  if (range.value?.length === 2) {
    params.start_at = range.value[0]
    params.end_at = range.value[1]
  }
  const res = await txApi.list(params)
  txList.value = res.transactions ?? []
  total.value = res.total ?? 0
  summary.value = res.summary ?? { revenue: 0, cost: 0, profit: 0, transaction_count: 0 }
}

const txTypeLabel = (t) => ({
  charge: '扣费', refund: '退款', recharge: '充值', hold: '预扣', settle: '结算'
}[t] ?? t)

const typeColor = (t) => ({ charge: 'danger', refund: 'success', recharge: 'success' }[t] ?? 'info')

function fmtTime(row, col, val) {
  return val ? new Date(val).toLocaleString('zh-CN') : '-'
}

function toYuan(credits) {
  return ((credits ?? 0) / 1e6).toFixed(4)
}

function profitOf(row) {
  const credits = row.credits ?? 0
  const cost = row.cost ?? 0
  if (row.type === 'refund') return -credits
  if (row.type === 'charge' || row.type === 'settle') return credits - cost
  return 0
}
</script>

<style scoped>
.billing-page { max-width: 1320px; }
.hero-card { margin-bottom: 16px; }
.hero-row { display:flex;align-items:center;justify-content:space-between;gap:16px; }
.eyebrow { color:#1e66ff;font-size:.82rem;font-weight:700;text-transform:uppercase;letter-spacing:.08em; }
.hero-row h3 { margin:8px 0 10px;font-size:1.55rem; }
.hero-row p { margin:0;color:#617086; }
.filter-panel { display:flex;align-items:center;gap:12px; }
.summary-row { margin-bottom: 16px; }
.summary-card .label { color:#72829a;font-size:.84rem;margin-bottom:8px; }
.summary-card .value { font-size:1.8rem;font-weight:800;color:#101828; }
.summary-card.profit .value { color:#155eef; }
@media (max-width: 900px) {
  .hero-row, .filter-panel { flex-direction:column;align-items:flex-start; }
}
</style>
