<template>
  <div>
    <div class="page-header">
      <h2 class="page-title">号商管理</h2>
    </div>

    <el-table :data="vendors" v-loading="loading" border style="width:100%">
      <el-table-column label="ID" prop="id" width="80" />
      <el-table-column label="用户名" prop="username" min-width="140" />
      <el-table-column label="余额 (元)" min-width="120">
        <template #default="{ row }">{{ formatCredits(row.balance) }}</template>
      </el-table-column>
      <el-table-column label="手续费比例" min-width="130">
        <template #default="{ row }">
          {{ formatPercent(row.commission_ratio) }}
          <span v-if="row.commission_ratio === null || row.commission_ratio === undefined" style="color:#86909c;font-size:12px">（全局）</span>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.is_active ? 'success' : 'danger'" size="small">
            {{ row.is_active ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="注册时间" min-width="160">
        <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="openEdit(row)">编辑</el-button>
          <el-button
            size="small"
            :type="row.is_active ? 'danger' : 'success'"
            @click="toggleActive(row)"
          >{{ row.is_active ? '禁用' : '启用' }}</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 编辑弹窗 -->
    <el-dialog v-model="editVisible" title="编辑号商" width="400px" align-center>
      <el-form :model="editForm" label-width="120px">
        <el-form-item label="手续费比例">
          <el-input-number
            v-model="commissionPercent"
            :min="0" :max="100" :precision="2" :step="1"
            @change="v => editForm.commission_ratio = v === null ? null : v / 100"
          />
          <span style="margin-left:8px">%</span>
          <div style="font-size:12px;color:#86909c;margin-top:4px">留空则使用全局设置；平台从收益中抽取该比例</div>
        </el-form-item>
        <el-form-item label="重置密码">
          <el-input v-model="editForm.new_password" type="password" show-password placeholder="留空则不修改" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">取消</el-button>
        <el-button type="primary" :loading="editLoading" @click="saveEdit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { vendorAdminApi } from '@/api/admin'

const loading = ref(true)
const vendors = ref([])
const editVisible = ref(false)
const editLoading = ref(false)
const editForm = ref({ id: null, commission_ratio: null, new_password: '' })
const commissionPercent = ref(null)

function formatCredits(v) {
  if (!v) return '0.0000'
  return (v / 1e6).toFixed(4)
}
function formatPercent(v) {
  if (v === null || v === undefined) return '-'
  return (v * 100).toFixed(2) + '%'
}
function formatTime(v) {
  if (!v) return '-'
  return new Date(v).toLocaleString('zh-CN', { hour12: false })
}

async function fetchVendors() {
  loading.value = true
  try {
    const res = await vendorAdminApi.list()
    vendors.value = res.vendors || []
  } catch {
    // 错误已展示
  } finally {
    loading.value = false
  }
}

function openEdit(row) {
  editForm.value = { id: row.id, commission_ratio: row.commission_ratio, new_password: '' }
  commissionPercent.value = row.commission_ratio !== null && row.commission_ratio !== undefined
    ? parseFloat((row.commission_ratio * 100).toFixed(2))
    : null
  editVisible.value = true
}

async function saveEdit() {
  editLoading.value = true
  try {
    const payload = {}
    if (editForm.value.commission_ratio !== null) payload.commission_ratio = editForm.value.commission_ratio
    if (editForm.value.new_password) payload.password = editForm.value.new_password
    await vendorAdminApi.update(editForm.value.id, payload)
    ElMessage.success('已保存')
    editVisible.value = false
    await fetchVendors()
  } finally {
    editLoading.value = false
  }
}

async function toggleActive(row) {
  try {
    await vendorAdminApi.update(row.id, { is_active: !row.is_active })
    row.is_active = !row.is_active
  } catch {
    // 错误已展示
  }
}

onMounted(fetchVendors)
</script>

<style scoped>
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.page-title {
  font-size: 20px;
  font-weight: 700;
  margin: 0;
}
</style>
