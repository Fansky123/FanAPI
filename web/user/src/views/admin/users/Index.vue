<template>
  <div class="users-page">
    <el-card class="hero-card">
      <div class="hero-row">
        <div>
          <div class="eyebrow">Users</div>
          <h3>用户与余额管理</h3>
          <p>查看用户注册状态、余额和手动充值情况，用于日常运营支持。</p>
        </div>
        <div class="hero-metric">
          <strong>{{ total }}</strong>
          <span>总用户数</span>
        </div>
      </div>
    </el-card>

    <el-card>
    <el-table :data="users" stripe border>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="username" label="用户名" width="120" />
      <el-table-column prop="email" label="邮箱" />
      <el-table-column prop="role" label="角色" width="80">
        <template #default="{ row }">
          <el-tag :type="row.role === 'admin' ? 'danger' : 'info'" size="small">{{ row.role }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="定价分组" width="130" align="center">
        <template #default="{ row }">
          <el-tag v-if="row.group" type="warning" size="small" style="cursor:pointer" @click="openSetGroup(row)">{{ row.group }}</el-tag>
          <el-button v-else size="small" text @click="openSetGroup(row)">默认（点击设置）</el-button>
        </template>
      </el-table-column>
      <el-table-column label="余额（¥）" width="140">
        <template #default="{ row }">
          ¥{{ (row.balance / 1e6).toFixed(4) }}
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="注册时间" :formatter="fmtTime" />
      <el-table-column label="操作" width="240" align="center">
        <template #default="{ row }">
          <el-button size="small" type="success" @click="openRecharge(row)">充值</el-button>
          <el-button size="small" type="warning" @click="openResetPwd(row)">改密</el-button>
          <el-button size="small" @click="openSetGroup(row)">分组</el-button>
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
    </el-card>

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

    <!-- 重置密码弹窗 -->
    <el-dialog v-model="showResetPwd" title="重置密码" width="380px">
      <p style="margin-bottom:12px">为用户 <b>{{ resetPwdUser?.email }}</b> 设置新密码</p>
      <el-form :model="pwdForm" :rules="pwdRules" ref="pwdFormRef" label-width="90px">
        <el-form-item label="新密码" prop="password">
          <el-input v-model="pwdForm.password" type="password" show-password placeholder="至少 8 位" />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirm">
          <el-input v-model="pwdForm.confirm" type="password" show-password placeholder="再次输入" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showResetPwd = false">取消</el-button>
        <el-button type="primary" @click="doResetPwd">确认修改</el-button>
      </template>
    </el-dialog>

    <!-- 设置定价分组弹窗 -->
    <el-dialog v-model="showSetGroup" title="设置定价分组" width="400px">
      <p style="margin-bottom:12px">用户 <b>{{ groupUser?.username || groupUser?.email }}</b></p>
      <el-input v-model="groupInput" placeholder="留空=默认定价，如 vip / premium" clearable />
      <p style="color:#909399;font-size:.82rem;margin-top:8px">分组名须与渠道 billing_config.pricing_groups 中的键对应</p>
      <template #footer>
        <el-button @click="showSetGroup = false">取消</el-button>
        <el-button type="primary" @click="doSetGroup">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { userApi } from '@/api/admin'
import { ElMessage } from 'element-plus'

const users = ref([])
const page = ref(1)
const total = ref(0)

// 充值
const showRecharge = ref(false)
const rechargeUser = ref(null)
const rechargeAmount = ref(1000000)

// 改密
const showResetPwd = ref(false)
const resetPwdUser = ref(null)
const pwdFormRef = ref(null)
const pwdForm = ref({ password: '', confirm: '' })
const pwdRules = {
  password: [{ required: true, min: 8, message: '至少 8 位', trigger: 'blur' }],
  confirm: [{
    validator: (rule, val, cb) =>
      val === pwdForm.value.password ? cb() : cb(new Error('两次密码不一致')),
    trigger: 'blur'
  }]
}

// 设置分组
const showSetGroup = ref(false)
const groupUser = ref(null)
const groupInput = ref('')

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

function openResetPwd(user) {
  resetPwdUser.value = user
  pwdForm.value = { password: '', confirm: '' }
  showResetPwd.value = true
}

async function doResetPwd() {
  await pwdFormRef.value.validate()
  await userApi.resetPassword(resetPwdUser.value.id, pwdForm.value.password)
  ElMessage.success(`已重置 ${resetPwdUser.value.email} 的密码`)
  showResetPwd.value = false
}

function openSetGroup(user) {
  groupUser.value = user
  groupInput.value = user.group || ''
  showSetGroup.value = true
}

async function doSetGroup() {
  await userApi.setGroup(groupUser.value.id, groupInput.value)
  ElMessage.success('已更新定价分组')
  showSetGroup.value = false
  fetchUsers()
}

function fmtTime(row, col, val) {
  return val ? new Date(val).toLocaleString('zh-CN') : '-'
}
</script>

<style scoped>
.users-page { max-width: 1320px; }
.hero-card { margin-bottom: 16px; }
.hero-row { display:flex;align-items:center;justify-content:space-between;gap:16px; }
.eyebrow { color:#1e66ff;font-size:.82rem;font-weight:700;text-transform:uppercase;letter-spacing:.08em; }
.hero-row h3 { margin:8px 0 10px;font-size:1.55rem; }
.hero-row p { margin:0;color:#617086; }
.hero-metric { min-width:120px;padding:16px;border-radius:16px;background:linear-gradient(180deg,#f7fbff,#eef5ff);border:1px solid #d8e6ff; }
.hero-metric strong { display:block;font-size:1.4rem;color:#0f172a; }
.hero-metric span { color:#72829a;font-size:.82rem; }
@media (max-width: 900px) { .hero-row { flex-direction:column;align-items:flex-start; } }
</style>
