<template>
  <div class="profile-page">
    <div class="page-title-header">
      <h2>个人中心</h2>
    </div>

    <div class="content-card">
      <div class="basic-info-row">
        <div class="avatar-circle">{{ userInitial }}</div>
        <div class="user-meta">
          <div class="user-name">{{ store.username || '未设置用户名' }}</div>
          <div class="user-email">{{ store.email || '未绑定邮箱' }}</div>
          <el-tag v-if="store.group" type="warning" effect="light" size="small" style="margin-top:6px">{{ store.group }}</el-tag>
          <el-tag v-else effect="plain" size="small" style="margin-top:6px">默认分组</el-tag>
        </div>
        <div class="info-stats">
          <div class="info-stat-item">
            <div class="info-stat-label">当前余额</div>
            <div class="info-stat-val primary">{{ (store.balance / 1e6).toFixed(4) }} 积分</div>
          </div>
          <div class="info-stat-item">
            <div class="info-stat-label">定价分组</div>
            <div class="info-stat-val">{{ store.group || '默认' }}</div>
          </div>
          <div class="info-stat-item">
            <div class="info-stat-label">绑定邮箱</div>
            <div class="info-stat-val">{{ store.email || '未绑定' }}</div>
          </div>
        </div>
      </div>
    </div>

    <div class="content-card">
      <div class="card-section-title">修改密码</div>
      <el-form :model="pwdForm" label-position="top" @submit.prevent="changePassword" style="max-width:480px">
        <el-form-item label="当前密码">
          <el-input v-model="pwdForm.old_password" type="password" show-password placeholder="输入当前密码" />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input v-model="pwdForm.new_password" type="password" show-password placeholder="至少8位，包含字母和数字" />
        </el-form-item>
        <el-form-item label="确认新密码">
          <el-input v-model="pwdForm.confirm" type="password" show-password placeholder="再次输入新密码" />
        </el-form-item>
        <el-button type="primary" native-type="submit" :loading="pwdLoading">
          确认修改
        </el-button>
      </el-form>
      <el-alert v-if="pwdMsg.text" :type="pwdMsg.type" :title="pwdMsg.text" show-icon closable @close="pwdMsg.text=''" style="margin-top:12px;max-width:480px" />
    </div>

    <div class="content-card">
      <div class="card-section-title">{{ store.email ? '邮箱绑定' : '绑定邮箱' }}</div>
      <template v-if="!store.email">
        <p class="form-desc">绑定邮箱后可用于找回密码</p>
        <el-form :model="emailForm" label-position="top" @submit.prevent="bindEmail" style="max-width:480px">
          <el-form-item label="邮箱地址">
            <el-input v-model="emailForm.email" placeholder="your@email.com" />
          </el-form-item>
          <el-form-item label="验证码">
            <div style="display:flex;gap:8px">
              <el-input v-model="emailForm.code" placeholder="6位验证码" />
              <el-button :loading="sendingCode" :disabled="codeCooldown > 0" @click="sendCode" style="white-space:nowrap">
                {{ codeCooldown > 0 ? `${codeCooldown}s 后重发` : '发送验证码' }}
              </el-button>
            </div>
          </el-form-item>
          <el-button type="primary" native-type="submit" :loading="emailLoading">
            绑定邮箱
          </el-button>
        </el-form>
      </template>
      <template v-else>
        <div class="bound-email">
          <el-icon style="color:#00b42a"><CircleCheck /></el-icon>
          <span>{{ store.email }}</span>
        </div>
        <p class="form-desc" style="margin-top:10px">邮箱已绑定，可用于找回密码</p>
      </template>
      <el-alert v-if="emailMsg.text" :type="emailMsg.type" :title="emailMsg.text" show-icon closable @close="emailMsg.text=''" style="margin-top:12px;max-width:480px" />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import { userApi, authApi } from '@/api'
import { ElMessage } from 'element-plus'
import { CircleCheck } from '@element-plus/icons-vue'

const store = useUserStore()
const userInitial = computed(() => (store.username || store.email || 'U').charAt(0).toUpperCase())

const pwdForm = ref({ old_password: '', new_password: '', confirm: '' })
const pwdLoading = ref(false)
const pwdMsg = ref({ text: '', type: 'success' })

async function changePassword() {
  if (!pwdForm.value.old_password || !pwdForm.value.new_password) return ElMessage.warning('请填写完整')
  if (pwdForm.value.new_password !== pwdForm.value.confirm) return ElMessage.warning('两次密码不一致')
  if (pwdForm.value.new_password.length < 8) return ElMessage.warning('密码不能少于8位')
  pwdLoading.value = true
  pwdMsg.value = { text: '', type: 'success' }
  try {
    await userApi.changePassword({ old_password: pwdForm.value.old_password, new_password: pwdForm.value.new_password })
    pwdMsg.value = { type: 'success', text: '密码修改成功' }
    pwdForm.value = { old_password: '', new_password: '', confirm: '' }
  } catch (e) {
    pwdMsg.value = { type: 'error', text: e?.response?.data?.error || '修改失败，请检查当前密码' }
  } finally {
    pwdLoading.value = false
  }
}

const emailForm = ref({ email: '', code: '' })
const emailLoading = ref(false)
const emailMsg = ref({ text: '', type: 'success' })
const sendingCode = ref(false)
const codeCooldown = ref(0)
let cooldownTimer = null

async function sendCode() {
  if (!emailForm.value.email) return ElMessage.warning('请先输入邮箱')
  sendingCode.value = true
  try {
    await authApi.sendCode(emailForm.value.email)
    codeCooldown.value = 60
    cooldownTimer = setInterval(() => {
      codeCooldown.value--
      if (codeCooldown.value <= 0) { clearInterval(cooldownTimer); cooldownTimer = null }
    }, 1000)
    ElMessage.success('验证码已发送')
  } catch (e) {
    ElMessage.error(e?.response?.data?.error || '发送失败')
  } finally {
    sendingCode.value = false
  }
}

async function bindEmail() {
  if (!emailForm.value.email || !emailForm.value.code) return ElMessage.warning('请填写邮箱和验证码')
  emailLoading.value = true
  emailMsg.value = { text: '', type: 'success' }
  try {
    await userApi.bindEmail({ email: emailForm.value.email, code: emailForm.value.code })
    store.setEmail(emailForm.value.email)
    emailMsg.value = { type: 'success', text: '邮箱绑定成功' }
    emailForm.value = { email: '', code: '' }
  } catch (e) {
    emailMsg.value = { type: 'error', text: e?.response?.data?.error || '绑定失败，验证码错误或已过期' }
  } finally {
    emailLoading.value = false
  }
}

onMounted(() => {
  if (store.token) { store.fetchBalance(); store.fetchProfile() }
})
</script>

<style scoped>
.profile-page { display: flex; flex-direction: column; }

.page-title-header {
  padding: 15px 24px;
  border-radius: 12px;
  background: #ffffff;
  border: 1px solid #ffffff;
  box-shadow: rgba(0,0,0,0.02) 0px 10px 20px 0px;
  margin-bottom: 15px;
}
.page-title-header h2 { margin: 0; font-size: 20px; font-weight: 600; color: rgb(26, 27, 28); }

.content-card { background: #ffffff; border-radius: 12px; padding: 24px; margin-bottom: 15px; }

.basic-info-row { display: flex; align-items: center; gap: 20px; flex-wrap: wrap; }

.avatar-circle {
  width: 80px; height: 80px; border-radius: 50%;
  background: #165dff; color: #fff;
  display: grid; place-items: center;
  font-size: 32px; font-weight: 700; flex-shrink: 0;
}

.user-meta { flex-shrink: 0; }
.user-name { font-size: 18px; font-weight: 700; color: #1d2129; margin-bottom: 4px; }
.user-email { font-size: 13px; color: #86909c; }

.info-stats { display: flex; gap: 32px; margin-left: auto; }
.info-stat-item { text-align: center; }
.info-stat-label { font-size: 12px; color: #86909c; margin-bottom: 4px; }
.info-stat-val { font-size: 16px; font-weight: 600; color: #1d2129; }
.info-stat-val.primary { color: #165dff; }

.card-section-title { font-size: 16px; font-weight: 600; color: #1d2129; margin-bottom: 20px; }
.form-desc { font-size: 13px; color: #86909c; margin: 0 0 14px; }
.bound-email { display: flex; align-items: center; gap: 8px; font-size: 15px; font-weight: 500; color: #1d2129; }

@media (max-width: 768px) {
  .info-stats { margin-left: 0; }
  .basic-info-row { align-items: flex-start; }
}
</style>
