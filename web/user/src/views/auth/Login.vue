<template>
  <div class="auth-page">
    <div class="auth-card">
      <!-- Logo -->
      <div class="auth-logo">
        <div class="logo-icon">{{ site.siteName.charAt(0).toUpperCase() }}</div>
        <span class="logo-name">{{ site.siteName }}</span>
      </div>

      <h2 class="auth-title">登录账户</h2>
      <p class="auth-sub">欢迎回来，请输入登录信息</p>

      <el-form :model="form" @submit.prevent="handleLogin" label-position="top" class="auth-form">
        <el-form-item label="用户名 / 邮箱">
          <el-input v-model="form.username" placeholder="用户名或绑定邮箱" clearable size="large" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" show-password size="large" placeholder="请输入密码" />
        </el-form-item>
        <div class="forgot-row">
          <router-link to="/forgot-password" class="forgot-link">忘记密码？</router-link>
        </div>
        <el-button type="primary" native-type="submit" :loading="loading" class="submit-btn">
          登 录
        </el-button>
      </el-form>

      <div class="link-row">
        还没有账号？<router-link to="/register" class="link-a">立即注册</router-link>
      </div>
    </div>
  </div>

  <!-- 客服二维码弹窗（登录成功后） -->
  <el-dialog v-model="showInviterQR" title="专属客服" width="320px" :close-on-click-modal="false" align-center>
    <div class="qr-dialog">
      <p>欢迎回来！扫码添加您的专属客服微信获取使用帮助。</p>
      <img :src="inviterQR" alt="客服二维码" class="qr-dialog-img" />
    </div>
    <template #footer>
      <el-button type="primary" @click="closeQRAndGo">进入首页</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useSiteStore } from '@/stores/site'
import { authApi } from '@/api'

const router = useRouter()
const store = useUserStore()
const site = useSiteStore()
const loading = ref(false)
const form = reactive({ username: '', password: '' })

const showInviterQR = ref(false)
const inviterQR = ref('')

async function handleLogin() {
  loading.value = true
  try {
    const res = await authApi.login(form)
    store.setToken(res.token)
    store.setUsername(res.user?.username || form.username)
    if (res.inviter_wechat_qr) {
      inviterQR.value = res.inviter_wechat_qr
      showInviterQR.value = true
    } else {
      router.push('/models')
    }
  } catch {
    // 错误已由 HTTP 拦截器展示
  } finally {
    loading.value = false
  }
}

function closeQRAndGo() {
  showInviterQR.value = false
  router.push('/models')
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: radial-gradient(ellipse 80% 60% at 50% -10%, rgba(37,99,235,.12) 0%, transparent 70%),
              linear-gradient(180deg, #f1f5f9 0%, #e9f0fb 100%);
  padding: 24px;
}
.auth-card {
  width: 100%;
  max-width: 420px;
  background: rgba(255,255,255,.95);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(37,99,235,.1);
  border-radius: 16px;
  padding: 40px 36px;
  box-shadow: 0 8px 40px rgba(37,99,235,.1), 0 2px 8px rgba(0,0,0,.04);
}
.auth-logo {
  display: flex; align-items: center; gap: 10px;
  margin-bottom: 24px; justify-content: center;
}
.logo-icon {
  width: 36px; height: 36px; border-radius: 10px;
  background: linear-gradient(135deg, #2563eb, #3b82f6);
  color: #fff;
  display: grid; place-items: center; font-weight: 700; font-size: 16px;
  box-shadow: 0 2px 8px rgba(37,99,235,.35);
}
.logo-name { font-size: 18px; font-weight: 700; color: var(--ow-text, #0f172a); }
.auth-title {
  margin: 0 0 6px; font-size: 22px; font-weight: 700;
  color: var(--ow-text, #0f172a); text-align: center; letter-spacing: -.02em;
}
.auth-sub { margin: 0 0 28px; color: var(--ow-subtext, #94a3b8); font-size: 13.5px; text-align: center; }
.auth-form :deep(.el-form-item__label) { font-weight: 500; color: var(--ow-text-2, #475569); font-size: 13px; padding-bottom: 4px; }
.auth-form :deep(.el-input__wrapper) { height: 42px; }
.forgot-row { text-align: right; margin-bottom: 20px; margin-top: -4px; }
.forgot-link { font-size: 12.5px; color: var(--ow-subtext, #94a3b8); text-decoration: none; transition: color .15s; }
.forgot-link:hover { color: var(--ow-primary, #2563eb); }
.submit-btn {
  width: 100%; height: 44px; font-size: 15px;
  letter-spacing: .04em; border-radius: 10px !important;
  background: linear-gradient(135deg, #2563eb, #3b82f6) !important;
  border: none !important;
  box-shadow: 0 2px 12px rgba(37,99,235,.3) !important;
}
.submit-btn:hover { opacity: .9; transform: translateY(-1px); box-shadow: 0 4px 20px rgba(37,99,235,.4) !important; }
.link-row { margin-top: 24px; text-align: center; color: var(--ow-subtext, #94a3b8); font-size: 13.5px; }
.link-a { color: var(--ow-primary, #2563eb); text-decoration: none; font-weight: 600; transition: opacity .15s; }
.link-a:hover { opacity: .8; }
.qr-dialog { text-align: center; }
.qr-dialog p { margin: 0 0 16px; color: var(--ow-text-2, #475569); font-size: 13.5px; line-height: 1.6; }
.qr-dialog-img {
  width: 200px; height: 200px; object-fit: contain;
  border: 1px solid var(--ow-border); border-radius: 12px;
}
</style>
