<template>
  <div>
    <div class="page-header">
      <div class="page-title">我的 API Key</div>
      <el-button type="primary" @click="showUpload = true">+ 上传新 Key</el-button>
    </div>

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
          <span class="earn-value">{{ formatEarn(row.my_earn) }}</span>
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

    <div v-if="!loading && keys.length === 0" class="empty-wrap">
      <el-empty description=" ">
        <template #description>
          <p class="empty-title">暂无已提交的 Key</p>
          <p class="empty-tip">点击右上角「上传新 Key」，提交您的 API Key。系统会自动验证有效性，通过后即进入渠道号池。</p>
        </template>
      </el-empty>
    </div>

    <!-- 上传 Key 对话框 -->
    <el-dialog v-model="showUpload" title="上传 API Key" width="500px" :close-on-click-modal="false" @open="onDialogOpen">
      <el-form :model="form" label-width="80px" @submit.prevent>
        <el-form-item label="目标渠道">
          <el-select
            v-model="form.pool_id"
            placeholder="请选择要加入的渠道号池"
            style="width:100%"
            :loading="poolsLoading"
          >
            <el-option
              v-for="pool in pools"
              :key="pool.id"
              :label="`${pool.channel_name}（${pool.name}）`"
              :value="pool.id"
            />
          </el-select>
          <div v-if="!poolsLoading && pools.length === 0" class="no-pools-tip">
            当前暂无开放自助上传的渠道，请联系平台运营方开通
          </div>
        </el-form-item>
        <el-form-item label="API Key">
          <el-input
            v-model="form.value"
            type="textarea"
            :rows="3"
            placeholder="请粘贴您的 API Key（如 sk-xxxxxxxx）"
            clearable
          />
        </el-form-item>
      </el-form>

      <div v-if="submitResult" :class="['result-box', submitResult.ok ? 'result-ok' : 'result-fail']">
        <el-icon v-if="submitResult.ok"><CircleCheckFilled /></el-icon>
        <el-icon v-else><CircleCloseFilled /></el-icon>
        <span>{{ submitResult.message }}</span>
      </div>

      <template #footer>
        <el-button @click="showUpload = false">取消</el-button>
        <el-button
          type="primary"
          :loading="submitting"
          :disabled="!form.pool_id || !form.value.trim() || pools.length === 0"
          @click="handleSubmit"
        >
          验证并提交
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { CircleCheckFilled, CircleCloseFilled } from '@element-plus/icons-vue'
import { vendorApi } from '@/api/vendor'

const loading = ref(true)
const keys = ref([])

const showUpload = ref(false)
const poolsLoading = ref(false)
const pools = ref([])
const submitting = ref(false)
const submitResult = ref(null)
const form = ref({ pool_id: null, value: '' })

function formatCredits(v) {
  if (!v) return '0.0000'
  return (v / 1e6).toFixed(4)
}

function formatEarn(v) {
  if (!v) return '0.0000'
  return (v / 1e6).toFixed(4)
}

function formatTime(v) {
  if (!v) return '-'
  return new Date(v).toLocaleString('zh-CN', { hour12: false })
}

async function loadKeys() {
  try {
    const res = await vendorApi.getKeys()
    keys.value = res.keys || []
  } catch {
    // 错误已由拦截器展示
  } finally {
    loading.value = false
  }
}

async function onDialogOpen() {
  submitResult.value = null
  form.value = { pool_id: null, value: '' }
  poolsLoading.value = true
  try {
    const res = await vendorApi.getPools()
    pools.value = res.pools || []
  } catch {
    pools.value = []
  } finally {
    poolsLoading.value = false
  }
}

async function handleSubmit() {
  submitResult.value = null
  submitting.value = true
  try {
    const res = await vendorApi.submitKey({
      pool_id: form.value.pool_id,
      value: form.value.value.trim(),
    })
    submitResult.value = { ok: true, message: res.message || 'Key 已成功加入号池' }
    // 刷新列表
    await loadKeys()
    // 延迟关闭对话框
    setTimeout(() => { showUpload.value = false }, 1500)
  } catch (e) {
    const msg = e.response?.data?.error || '提交失败，请稍后重试'
    submitResult.value = { ok: false, message: msg }
  } finally {
    submitting.value = false
  }
}

onMounted(loadKeys)
</script>

<style scoped>
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}
.page-title {
  font-size: 20px;
  font-weight: 700;
  color: #1d2129;
}
.loading-wrap { padding: 20px 0; }
.empty-wrap { padding: 40px 0; text-align: center; }
.empty-title { font-size: 15px; font-weight: 600; color: #1d2129; margin: 0 0 8px; }
.empty-tip { font-size: 13px; color: #86909c; max-width: 380px; margin: 0 auto; line-height: 1.6; }
.masked-key { font-family: monospace; color: #4e5969; }
.earn-value { color: #00b42a; font-weight: 600; }
.no-pools-tip { font-size: 12px; color: #f56c6c; margin-top: 6px; }
.result-box {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 16px;
  padding: 10px 14px;
  border-radius: 6px;
  font-size: 14px;
}
.result-ok { background: #f0f9eb; color: #67c23a; border: 1px solid #c2e7b0; }
.result-fail { background: #fef0f0; color: #f56c6c; border: 1px solid #fbc4c4; }
</style>

