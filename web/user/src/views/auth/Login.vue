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
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useSiteStore } from '@/stores/site'
import { authApi } from '@/api'
import { ElMessage } from 'element-plus'

const router = useRouter()
const store = useUserStore()
const site = useSiteStore()
const loading = ref(false)
const form = reactive({ username: '', password: '' })

async function handleLogin() {
  loading.value = true
  try {
    const res = await authApi.login(form)
    store.setToken(res.token)
    store.setUsername(res.user?.username || form.username)
    router.push('/models')
  } catch {
    // 错误已由 HTTP 拦截器展示
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(126.82deg, rgba(236,243,255,.8) 0.58%, rgba(232,247,251,.8) 86.28%), #f2f3f5;
  padding: 24px;
}
.auth-card {
  width: 100%;
  max-width: 400px;
  background: #fff;
  border: 1px solid #e5e6eb;
  border-radius: 8px;
  padding: 36px 32px;
  box-shadow: 0 4px 20px rgba(0,0,0,.06);
}
.auth-logo {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 28px;
  justify-content: center;
}
.logo-icon {
  width: 32px; height: 32px;
  border-radius: 8px;
  background: #165dff;
  color: #fff;
  display: grid; place-items: center;
  font-weight: 700; font-size: 15px;
}
.logo-name {
  font-size: 17px;
  font-weight: 700;
  color: #1d2129;
}
.auth-title {
  margin: 0 0 6px;
  font-size: 20px;
  font-weight: 700;
  color: #1d2129;
  text-align: center;
}
.auth-sub {
  margin: 0 0 24px;
  color: #86909c;
  font-size: 13px;
  text-align: center;
}
.auth-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: #1d2129;
  font-size: 13px;
}
.forgot-row {
  text-align: right;
  margin-bottom: 16px;
}
.forgot-link {
  font-size: 12px;
  color: #86909c;
  text-decoration: none;
}
.forgot-link:hover { color: #165dff; }
.submit-btn {
  width: 100%;
  height: 40px;
  font-size: 14px;
  letter-spacing: .04em;
}
.link-row {
  margin-top: 20px;
  text-align: center;
  color: #86909c;
  font-size: 13px;
}
.link-a {
  color: #165dff;
  text-decoration: none;
  font-weight: 500;
}
.link-a:hover { text-decoration: underline; }
</style>
