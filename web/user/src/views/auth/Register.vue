<template>
  <div class="auth-page">
    <div class="auth-card">
      <!-- Logo -->
      <div class="auth-logo">
        <div class="logo-icon">{{ site.siteName.charAt(0).toUpperCase() }}</div>
        <span class="logo-name">{{ site.siteName }}</span>
      </div>

      <h2 class="auth-title">创建账户</h2>
      <p class="auth-sub">填写以下信息完成注册</p>

      <el-form :model="form" @submit.prevent="handleRegister" label-position="top" class="auth-form">
        <el-form-item label="用户名">
          <el-input v-model="form.username" placeholder="3-32 个字符" clearable size="large" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" show-password placeholder="至少 8 位" size="large" />
        </el-form-item>
        <el-button type="primary" native-type="submit" :loading="loading" class="submit-btn">
          完成注册
        </el-button>
      </el-form>

      <div class="hint-tip">
        <el-icon><InfoFilled /></el-icon>
        注册后可在账户设置中绑定邮箱，以便通过邮箱找回密码。
      </div>

      <div class="link-row">
        已有账号？<router-link to="/login" class="link-a">立即登录</router-link>
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
import { InfoFilled } from '@element-plus/icons-vue'

const router = useRouter()
const store = useUserStore()
const site = useSiteStore()
const loading = ref(false)
const form = reactive({ username: '', password: '' })

async function handleRegister() {
  if (!form.username || form.username.length < 3) return ElMessage.warning('用户名至少 3 个字符')
  if (!form.password || form.password.length < 8) return ElMessage.warning('密码至少 8 位')
  loading.value = true
  try {
    const res = await authApi.register(form)
    store.setToken(res.token)
    store.setUsername(res.user?.username || form.username)
    ElMessage.success('注册成功，欢迎！')
    router.push('/models')
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
.submit-btn {
  width: 100%;
  height: 40px;
  font-size: 14px;
  letter-spacing: .04em;
  margin-top: 4px;
}
.hint-tip {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  margin-top: 16px;
  padding: 10px 12px;
  background: #f0f4ff;
  border-radius: 4px;
  color: #165dff;
  font-size: 12px;
  line-height: 1.6;
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
