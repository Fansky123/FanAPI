<template>
  <div class="users-page">
    <el-card class="hero-card">
      <div class="hero-row">
        <div>
          <div class="eyebrow">Users</div>
          <h3>用户管理</h3>
          <p>按余额从少到多排列，红色标记余额不足用户，可直接为用户充值积分。</p>
        </div>
        <div class="hero-metric">
          <strong>{{ total }}</strong>
          <span>总用户数</span>
        </div>
      </div>
    </el-card>

    <el-card>
      <el-table :data="users" stripe border :row-class-name="rowClass">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="username" label="用户名" width="130" />
        <el-table-column prop="email" label="邮箱" min-width="160" />
        <el-table-column label="余额（¥）" width="140" sortable :sort-method="() => 0">
          <template #default="{ row }">
            <span :class="{ 'low-balance': row.balance < 1000000 }">
              ¥{{ (row.balance / 1e6).toFixed(4) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="注册时间" :formatter="fmtTime" width="180" />
        <el-table-column label="操作" width="100" align="center">
          <template #default="{ row }">
            <el-button size="small" type="success" @click="openRecharge(row)">充值</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="page"
        :page-size="50"
        :total="total"
        style="margin-top:16px"
        @current-change="fetchUsers"
      />
    </el-card>

    <!-- 充值弹窗 -->
    <el-dialog v-model="showRecharge" title="手动充值" width="360px">
      <p style="margin-bottom:12px">为用户 <b>{{ rechargeUser?.username }}</b> 充值</p>
      <el-input-number v-model="rechargeAmount" :min="1" :precision="0" style="width:100%" />
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
import { agentUserApi } from '@/api/agent'
import { ElMessage } from 'element-plus'

const users = ref([])
const page = ref(1)
const total = ref(0)

const showRecharge = ref(false)
const rechargeUser = ref(null)
const rechargeAmount = ref(1000000)

onMounted(fetchUsers)

async function fetchUsers() {
  const res = await agentUserApi.list(page.value, 50)
  users.value = res.users ?? []
  total.value = res.total ?? 0
}

function openRecharge(user) {
  rechargeUser.value = user
  rechargeAmount.value = 1000000
  showRecharge.value = true
}

async function doRecharge() {
  await agentUserApi.recharge(rechargeUser.value.id, rechargeAmount.value)
  ElMessage.success(`已为 ${rechargeUser.value.username} 充值 ${rechargeAmount.value.toLocaleString()} credits`)
  showRecharge.value = false
  fetchUsers()
}

// 余额低于 1元 的行标红
function rowClass({ row }) {
  return row.balance < 1000000 ? 'row-low' : ''
}

function fmtTime(row, col, val) {
  return val ? new Date(val).toLocaleString('zh-CN') : '-'
}
</script>

<style scoped>
.users-page { max-width: 1200px; }
.hero-card { margin-bottom: 16px; }
.hero-row { display:flex;align-items:center;justify-content:space-between;gap:16px; }
.eyebrow { color:#10b981;font-size:.82rem;font-weight:700;text-transform:uppercase;letter-spacing:.08em; }
.hero-row h3 { margin:8px 0 10px;font-size:1.45rem; }
.hero-row p { margin:0;color:#617086;font-size:.9rem; }
.hero-metric { min-width:110px;padding:14px;border-radius:14px;background:linear-gradient(180deg,#f0fff8,#e6ffef);border:1px solid #a7f3d0; }
.hero-metric strong { display:block;font-size:1.4rem;color:#065f46; }
.hero-metric span { color:#6b9e88;font-size:.82rem; }
.low-balance { color: #ef4444; font-weight: 700; }
</style>

<!-- 全局样式：标红行 -->
<style>
.el-table .row-low td { background: #fff5f5 !important; }
</style>
