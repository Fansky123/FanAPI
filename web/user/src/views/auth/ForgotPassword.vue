<template>
  <div class="login-page">
    <div class="forgot-shell">
      <div class="brand-row">
        <div class="brand-icon">F</div>
        <span class="brand-name">FanAPI</span>
      </div>

      <!-- Step 1: 输入邮箱 -->
      <template v-if="step === 1">
        <h3>找回密码</h3>
        <p class="sub">输入绑定的邮箱，我们将发送验证码</p>
        <el-form @submit.prevent="sendCode" label-position="top">
          <el-form-item label="绑定邮箱">
            <el-input v-model="email" placeholder="your@email.com" clearable />
          </el-form-item>
          <el-button type="primary" native-type="submit" :loading="sending" style="width:100%;height:42px">
            发送验证码
          </el-button>
        </el-form>
      </template>

      <!-- Step 2: 验证码 + 新密码 -->
      <template v-else-if="step === 2">
        <h3>重置密码</h3>
        <p class="sub">验证码已发送到 <b>{{ email }}</b></p>
        <el-form @submit.prevent="doReset" label-position="top">
          <el-form-item label="验证码">
            <div style="display:flex;gap:8px">
              <el-input v-model="code" placeholder="6 位验证码" />
              <el-button :disabled="countdown > 0" @click="sendCode">
                {{ countdown > 0 ? `${countdown}s` : '重新发送' }}
              </el-button>
            </div>
          </el-form-item>
          <el-form-item label="新密码">
            <el-input v-model="newPassword" type="password" show-password placeholder="至少 8 位" />
          </el-form-item>
          <el-button type="primary" native-type="submit" :loading="resetting" style="width:100%;height:42px">
            确认重置
          </el-button>
        </el-form>
      </template>

      <!-- Step 3: 成功 -->
      <template v-else>
        <div class="success-icon">✓</div>
        <h3>密码已重置</h3>
        <p class="sub">请使用新密码登录</p>
        <el-button type="primary" style="width:100%;height:42px" @click="router.push('/login')">
          返回登录
        </el-button>
      </template>

      <div class="back-row">
        <router-link to="/login">← 返回登录</router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { authApi } from '@/api'
import { ElMessage } from 'element-plus'

const router = useRouter()
const step = ref(1)
const email = ref('')
const code = ref('')
const newPassword = ref('')
const sending = ref(false)
const resetting = ref(false)
const countdown = ref(0)

async function sendCode() {
  if (!email.value) return ElMessage.warning('请输入邮箱')
  sending.value = true
  try {
    await authApi.forgotPassword(email.value)
    ElMessage.success('如该邮箱已绑定账号，验证码已发送')
    step.value = 2
    startCountdown()
  } finally {
    sending.value = false
  }
}

function startCountdown() {
  countdown.value = 60
  const t = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) clearInterval(t)
  }, 1000)
}

async function doReset() {
  if (!code.value) return ElMessage.warning('请输入验证码')
  if (!newPassword.value || newPassword.value.length < 8) return ElMessage.warning('新密码至少 8 位')
  resetting.value = true
  try {
    await authApi.resetPassword({ email: email.value, code: code.value, password: newPassword.value })
    ElMessage.success('密码已重置')
    step.value = 3
  } finally {
    resetting.value = false
  }
}
</script>

<style scoped>
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: #f4f6fb;
  padding: 24px;
}
.forgot-shell {
  width: 420px;
  background: #fff;
  border-radius: 20px;
  padding: 40px 36px;
  border: 1px solid #e0e8f5;
  box-shadow: 0 12px 40px rgba(26,64,135,.1);
}
.brand-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 28px;
}
.brand-icon {
  width: 28px; height: 28px;
  border-radius: 8px;
  background: #1677ff;
  display: grid; place-items: center;
  color: #fff; font-weight: 800; font-size: .9rem;
}
.brand-name { font-weight: 700; font-size: .95rem; color: #0d1526; }
h3 { margin: 0 0 6px; font-size: 1.35rem; }
.sub { margin: 0 0 22px; color: #617086; font-size: .88rem; }
.success-icon {
  width: 52px; height: 52px;
  border-radius: 50%;
  background: #e6fff0;
  color: #26a65b;
  font-size: 1.6rem;
  display: grid; place-items: center;
  margin: 0 auto 16px;
}
.back-row {
  margin-top: 20px;
  text-align: center;
  font-size: .85rem;
  color: #909399;
}
</style>
