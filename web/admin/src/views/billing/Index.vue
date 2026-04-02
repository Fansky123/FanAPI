<template>
  <div>
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
            {{ row.type === 'charge' ? '-' : '+' }}{{ Math.abs(row.amount).toLocaleString() }} cr
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
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { txApi } from '@/api'

const txList = ref([])
const page = ref(1)
const total = ref(0)

onMounted(fetchTx)

async function fetchTx() {
  const res = await txApi.list(page.value, 20)
  txList.value = res.transactions ?? []
  total.value = res.total ?? 0
}

const txTypeLabel = (t) => ({
  charge: '扣费', refund: '退款', recharge: '充值', hold: '预扣', settle: '结算'
}[t] ?? t)

const typeColor = (t) => ({ charge: 'danger', refund: 'success', recharge: 'success' }[t] ?? 'info')

function fmtTime(row, col, val) {
  return val ? new Date(val).toLocaleString('zh-CN') : '-'
}
</script>
