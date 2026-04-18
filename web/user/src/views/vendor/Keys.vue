<template>
  <div>
    <div class="page-title">我的 API Key</div>

    <div v-if="loading" class="loading-wrap">
      <el-skeleton :rows="4" animated />
    </div>

    <el-table v-else :data="keys" border style="width:100%">
      <el-table-column label="渠道名称" prop="channel_name" min-width="140" />
      <el-table-column label="Key（隐藏）" min-width="200">
        <template #default="{ row }">
          <span class="masked-key">{{ row.masked_value }}</span>
        </template>
      </el-table-column>
      <el-table-column label="累计消耗 (元)" min-width="130">
        <template #default="{ row }">
          {{ formatCredits(row.total_cost) }}
        </template>
      </el-table-column>
      <el-table-column label="我的收益 (元)" min-width="130">
        <template #default="{ row }">
          <span class="earn-value">{{ formatCredits(row.my_earn) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.is_active ? 'success' : 'danger'" size="small">
            {{ row.is_active ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="添加时间" min-width="160">
        <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
      </el-table-column>
    </el-table>

    <el-empty v-if="!loading && keys.length === 0" description="暂无 Key，请联系管理员添加" />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { vendorApi } from '@/api/vendor'

const loading = ref(true)
const keys = ref([])

function formatCredits(v) {
  if (!v) return '0.0000'
  return (v / 1e6).toFixed(4)
}

function formatTime(v) {
  if (!v) return '-'
  return new Date(v).toLocaleString('zh-CN', { hour12: false })
}

onMounted(async () => {
  try {
    const res = await vendorApi.getKeys()
    keys.value = res.keys || []
  } catch {
    // 错误已由拦截器展示
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.page-title {
  font-size: 20px;
  font-weight: 700;
  color: #1d2129;
  margin-bottom: 20px;
}
.loading-wrap { padding: 20px 0; }
.masked-key { font-family: monospace; color: #4e5969; }
.earn-value { color: #00b42a; font-weight: 600; }
</style>
