<template>
  <div>
    <el-table :data="users" stripe border>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="email" label="邮箱" />
      <el-table-column prop="role" label="角色" width="80">
        <template #default="{ row }">
          <el-tag :type="row.role === 'admin' ? 'danger' : 'info'" size="small">{{ row.role }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="余额（¥）" width="140">
        <template #default="{ row }">
          ¥{{ (row.balance / 1e6).toFixed(4) }}
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="注册时间" :formatter="fmtTime" />
      <el-table-column label="操作" width="120" align="center">
        <template #default="{ row }">
          <el-button size="small" type="success" @click="openRecharge(row)">充值</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-model:current-page="page"
      :page-size="20"
      :total="total"
      style="margin-top:16px"
      @current-change="fetchUsers"
    />

    <!-- 充值弹窗 -->
    <el-dialog v-model="showRecharge" title="手动充值" width="360px">
      <p style="margin-bottom:12px">为用户 <b>{{ rechargeUser?.email }}</b> 充值</p>
      <el-input-number v-model="rechargeAmount" :min="1" :precision="0" placeholder="credits 数量" style="width:100%" />
      <p style="color:#909399;font-size:.82rem;margin-top:8px">
        {{ rechargeAmount.toLocaleString() }} credits = ¥{{ (rechargeAmount / 1e6).toFixed(6) }}
      </p>
      <template #footer>
        <el-button @click="showRecharge = false">取消</el-button>
        <el-button type="primary" @click="doRecharge">确认充值</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { userApi } from '@/api'
import { ElMessage } from 'element-plus'

const users = ref([])
const page = ref(1)
const total = ref(0)
const showRecharge = ref(false)
const rechargeUser = ref(null)
const rechargeAmount = ref(1000000)

onMounted(fetchUsers)

async function fetchUsers() {
  const res = await userApi.list(page.value, 20)
  users.value = res.users ?? []
  total.value = res.total ?? 0
}

function openRecharge(user) {
  rechargeUser.value = user
  rechargeAmount.value = 1000000
  showRecharge.value = true
}

async function doRecharge() {
  await userApi.recharge(rechargeUser.value.id, rechargeAmount.value)
  ElMessage.success(`已为 ${rechargeUser.value.email} 充值 ${rechargeAmount.value.toLocaleString()} credits`)
  showRecharge.value = false
  fetchUsers()
}

function fmtTime(row, col, val) {
  return val ? new Date(val).toLocaleString('zh-CN') : '-'
}
</script>
