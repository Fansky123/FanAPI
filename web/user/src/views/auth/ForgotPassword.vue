<template>
  <div class="auth-page">
    <div class="auth-card">
      <div class="auth-logo">
        <div class="auth-logo-icon">{{ siteStore.siteName?.[0] ?? 'F' }}</div>
        <span class="auth-logo-name">{{ siteStore.siteName || 'FanAPI' }}</span>
      </div>

      <!-- Step 1: 输入邮箱 -->
      <template v-if="step === 1">
        <h2 class="auth-title">找回密码</h2>
        <p class="auth-sub">输入绑定的邮箱，我们将发送验证码</p>
        <el-form @submit.prevent="sendCode" label-position="top" class="auth-form">
          <el-form-item label="绑定邮箱">
            <el-input v-model="email" placeholder="your@email.com" clearable />
          </el-form-item>
          <el-button type="primary" native-type="submit" :loading="sending" class="auth-btn">
            发送验证码
          </el-button>
        </el-form>
      </template>

      <!-- Step 2: 验证码 + 新密码 -->
      <template v-else-if="step === 2">
        <h2 class="auth-title">重置密码</h2>
        <p class="auth-sub">验证码已发送到 <b>{{ email }}</b></p>
        <el-form @submit.prevent="doReset" label-position="top" class="auth-form">
          <el-form-item label="验证码">
            <div class="code-row">
              <el-input v-model="code" placeholder="6 位验证码" />
              <el-button :disabled="countdown > 0" @click="sendCode">
                {{ countdown > 0 ? `${countdown}s` : '重新发送' }}
              </el-button>
            </div>
          </el-form-item>
          <el-form-item label="新密码">
            <el-input v-model="newPassword" type="password" show-password placeholder="至少 8 位" />
          </el-form-item>
          <el-button type="primary" native-type="submit" :loading="resetting" class="auth-btn">
            确认重置
          </el-button>
        </el-form>
      </template>

      <!-- Step 3: 成功 -->
      <template v-else>
        <div class="success-icon">✓</div>
        <h2 class="auth-title">密码已重置</h2>
        <p class="auth-sub">请使用新密码登录</p>
        <el-button type="primary" class="auth-btn" @click="router.push('/login')">
          返回登录
        </el-button>
      </template>

      <div class="auth-footer">
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
import { useSiteStore } from '@/stores/site'

const router = useRouter()
const siteStore = useSiteStore()
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
.auth-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: linear-gradient(126.82deg, rgba(236,243,255,.7) 0.58%, rgba(232,247,251,.7) 86.28%), #f2f3f5;
  padding: 24px;
}
.auth-card {
  width: 100%;
  max-width: 400px;
  background: #fff;
  border-radius: 8px;
  padding: 40px 36px 32px;
  border: 1px solid var(--ow-border, #e5e6eb);
  box-shadow: 0 4px 20px rgba(22, 93, 255, .06);
}
.auth-logo {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 28px;
}
.auth-logo-icon {
  width: 28px; height: 28px;
  border-radius: 6px;
  background: var(--ow-primary, #165dff);
  color: #fff;
  font-weight: 800;
  font-size: .9rem;
  display: grid; place-items: center;
}
.auth-logo-name { font-weight: 700; font-size: .95rem; color: var(--ow-text, #1d2129); }
.auth-title { margin: 0 0 6px; font-size: 1.3rem; font-weight: 600; color: var(--ow-text, #1d2129); }
.auth-sub { margin: 0 0 24px; color: var(--ow-subtext, #86909c); font-size: .88rem; }
.auth-form :deep(.el-input__wrapper) { border-radius: 4px; }
.auth-btn { width: 100%; height: 40px; border-radius: 4px; font-size: .9rem; }
.code-row { display: flex; gap: 8px; width: 100%; }
.code-row .el-input { flex: 1; }
.success-icon {
  width: 52px; height: 52px;
  border-radius: 50%;
  background: #e6fff0;
  color: #26a65b;
  font-size: 1.6rem;
  display: grid; place-items: center;
  margin: 0 auto 16px;
}
.auth-footer {
  margin-top: 20px;
  text-align: center;
  font-size: .85rem;
  color: var(--ow-subtext, #86909c);
}
.auth-footer a { color: var(--ow-primary, #165dff); text-decoration: none; }
.auth-footer a:hover { text-decoration: underline; }
</style>
