<template>
  <div>
    <el-card style="margin-bottom:20px">
      <div style="display:flex;align-items:center;gap:24px">
        <div>
          <div style="color:#909399;font-size:.85rem">当前余额</div>
          <div style="font-size:2rem;font-weight:700">¥{{ (store.balance / 1e6).toFixed(4) }}</div>
          <div style="color:#c0c4cc;font-size:.8rem">{{ store.balance.toLocaleString() }} credits</div>
        </div>
        <el-divider direction="vertical" style="height:60px" />
        <el-button type="primary" size="large" @click="showRecharge = true">充值</el-button>
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
              {{ row.type === 'charge' ? '-' : '+' }}{{ (Math.abs(row.amount) / 1e6).toFixed(6) }} ¥
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
