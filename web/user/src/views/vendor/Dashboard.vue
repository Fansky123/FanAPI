<template>
  <div>
    <div class="page-title">仪表板</div>

    <div v-if="loading" class="loading-wrap">
      <el-skeleton :rows="3" animated />
    </div>

    <template v-else>
      <div class="stat-row">
        <div class="stat-card">
          <div class="stat-label">账户余额</div>
          <div class="stat-value blue">{{ formatCredits(profile.balance) }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">提供 Key 数量</div>
          <div class="stat-value">{{ profile.key_count ?? 0 }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">手续费比例</div>
          <div class="stat-value">{{ formatPercent(profile.commission_ratio) }}</div>
          <div class="stat-hint">平台收取，实际到账 {{ formatPercent(1 - (profile.commission_ratio || 0)) }}</div>
        </div>
      </div>

      <el-alert v-if="!profile.is_active" type="warning" show-icon :closable="false" style="margin-top:20px">
        <template #title>账号已被禁用，请联系管理员处理。</template>
      </el-alert>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { vendorApi } from '@/api/vendor'

const loading = ref(true)
const profile = ref({})

function formatCredits(v) {
  if (!v) return '0.0000 元'
  return (v / 1e6).toFixed(4) + ' 元'
}

function formatPercent(v) {
  if (!v) return '0%'
  return (v * 100).toFixed(2) + '%'
}

onMounted(async () => {
  try {
    profile.value = await vendorApi.getProfile()
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
.stat-row {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
}
.stat-card {
  flex: 1;
  min-width: 160px;
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 12px;
  padding: 20px 24px;
}
.stat-label {
  font-size: 13px;
  color: #86909c;
  margin-bottom: 8px;
}
.stat-value {
  font-size: 26px;
  font-weight: 700;
  color: #1d2129;
}
.stat-value.blue { color: #165dff; }
.stat-hint {
  font-size: 12px;
  color: #86909c;
  margin-top: 4px;
}
</style>
