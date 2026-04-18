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
  background: radial-gradient(ellipse 80% 60% at 50% -10%, rgba(37,99,235,.12) 0%, transparent 70%),
              linear-gradient(180deg, #f1f5f9 0%, #e9f0fb 100%);
  padding: 24px;
}
.auth-card {
  width: 100%;
  max-width: 420px;
  background: rgba(255,255,255,.95);
  backdrop-filter: blur(12px);
  border-radius: 16px;
  padding: 40px 36px 32px;
  border: 1px solid rgba(37,99,235,.1);
  box-shadow: 0 8px 40px rgba(37,99,235,.1), 0 2px 8px rgba(0,0,0,.04);
}
.auth-logo {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 28px;
  justify-content: center;
}
.auth-logo-icon {
  width: 36px; height: 36px;
  border-radius: 10px;
  background: linear-gradient(135deg, #2563eb, #3b82f6);
  color: #fff;
  font-weight: 800;
  font-size: 16px;
  display: grid; place-items: center;
  box-shadow: 0 2px 8px rgba(37,99,235,.35);
}
.auth-logo-name { font-weight: 700; font-size: 18px; color: var(--ow-text, #0f172a); }
.auth-title { margin: 0 0 6px; font-size: 22px; font-weight: 700; color: var(--ow-text, #0f172a); letter-spacing: -.02em; }
.auth-sub { margin: 0 0 28px; color: var(--ow-subtext, #94a3b8); font-size: 13.5px; }
.auth-form :deep(.el-form-item__label) { font-weight: 500; color: var(--ow-text-2, #475569); font-size: 13px; padding-bottom: 4px; }
.auth-form :deep(.el-input__wrapper) { height: 42px; border-radius: 10px !important; }
.auth-btn {
  width: 100%; height: 44px; border-radius: 10px !important; font-size: 15px;
  background: linear-gradient(135deg, #2563eb, #3b82f6) !important;
  border: none !important;
  box-shadow: 0 2px 12px rgba(37,99,235,.3) !important;
}
.auth-btn:hover { opacity: .9; transform: translateY(-1px); }
.code-row { display: flex; gap: 8px; width: 100%; }
.code-row .el-input { flex: 1; }
.success-icon {
  width: 56px; height: 56px;
  border-radius: 50%;
  background: var(--ow-success-bg, #ecfdf5);
  color: var(--ow-success, #10b981);
  font-size: 1.6rem;
  display: grid; place-items: center;
  margin: 0 auto 16px;
}
.auth-footer {
  margin-top: 24px;
  text-align: center;
  font-size: 13.5px;
  color: var(--ow-subtext, #94a3b8);
}
.auth-footer a { color: var(--ow-primary, #2563eb); text-decoration: none; font-weight: 500; transition: opacity .15s; }
.auth-footer a:hover { opacity: .8; }
</style>
