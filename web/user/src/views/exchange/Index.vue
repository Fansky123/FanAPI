<template>
  <div class="exchange-page">
    <div class="page-title-header">
      <h2>兑换中心</h2>
    </div>

    <div class="content-card">
      <div class="section-label">兑换卡密</div>
      <div class="redeem-row">
        <el-input
          v-model="code"
          placeholder="请输入兑换码（例如：XXXX-XXXX-XXXX-XXXX）"
          clearable
          style="flex:1"
          @keyup.enter="redeem"
          :disabled="loading"
        />
        <el-button type="primary" :loading="loading" @click="redeem" style="width:120px">
          立即兑换
        </el-button>
      </div>
      <el-alert v-if="result.msg" :type="result.type" :title="result.msg" show-icon closable @close="result.msg = ''" style="margin-top:16px" />
    </div>

    <div class="content-card">
      <div class="section-label" style="margin-bottom:16px">兑换记录</div>
      <el-table :data="history" stripe border size="default" v-loading="historyLoading">
        <el-table-column prop="code" label="兑换码" min-width="200" show-overflow-tooltip />
        <el-table-column label="积分数量" width="150" align="right">
          <template #default="{ row }">
            <span style="color:#165dff;font-weight:600">{{ fmtCredits(row.credits ?? row.amount) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="兑换时间" width="180">
          <template #default="{ row }">
            {{ fmtTime(row.created_at ?? row.redeemed_at) }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default>
            <el-tag type="success" size="small">已兑换</el-tag>
          </template>
        </el-table-column>
      </el-table>
      <div v-if="!historyLoading && history.length === 0" class="empty-hint">暂无兑换记录</div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import { userApi } from '@/api'
import { ElMessage } from 'element-plus'

const store = useUserStore()
const code = ref('')
const loading = ref(false)
const result = ref({ msg: '', type: 'success' })

const history = ref([])
const historyLoading = ref(false)

function fmtCredits(v) {
  if (!v) return '0.0000'
  return (v / 1e6).toFixed(4)
}

function fmtTime(v) {
  if (!v) return '-'
  return new Date(v).toLocaleString('zh-CN')
}

async function redeem() {
  const c = code.value.trim()
  if (!c) return ElMessage.warning('请输入兑换码')
  loading.value = true
  result.value = { msg: '', type: 'success' }
  try {
    const res = await userApi.redeemCard(c)
    const credits = res.credits ?? res.credits_added ?? res.amount ?? 0
    result.value = {
      type: 'success',
      msg: `兑换成功！获得 ${(credits / 1e6).toFixed(4)} 积分`
    }
    code.value = ''
    store.fetchBalance()
    fetchHistory()
  } catch (e) {
    result.value = {
      type: 'error',
      msg: e?.response?.data?.error || e?.message || '兑换失败，兑换码无效或已使用'
    }
  } finally {
    loading.value = false
  }
}

async function fetchHistory() {
  historyLoading.value = true
  try {
    const res = await userApi.getRedeemHistory?.()
    if (res) history.value = res.records ?? res.list ?? res ?? []
  } catch {
    history.value = []
  } finally {
    historyLoading.value = false
  }
}

onMounted(() => {
  if (store.token) {
    store.fetchBalance()
    fetchHistory()
  }
})
</script>

<style scoped>
.exchange-page { display: flex; flex-direction: column; }

.page-title-header {
  padding: 15px 24px;
  border-radius: 12px;
  background: #ffffff;
  border: 1px solid #ffffff;
  box-shadow: rgba(0,0,0,0.02) 0px 10px 20px 0px;
  margin-bottom: 15px;
}
.page-title-header h2 { margin: 0; font-size: 20px; font-weight: 600; color: rgb(26, 27, 28); }

.content-card { background: #ffffff; border-radius: 12px; padding: 20px; margin-bottom: 15px; }

.section-label { font-size: 15px; font-weight: 600; color: #1a1b1c; margin-bottom: 14px; }
.redeem-row { display: flex; gap: 12px; align-items: center; }
.empty-hint { text-align: center; color: #86909c; padding: 30px 0; font-size: 14px; }
</style>
