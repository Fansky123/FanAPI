<template>
  <div class="login-page">
    <div class="login-shell">
      <div class="hero-panel">
        <div class="brand">FanAPI Admin</div>
        <h2>管理后台控制中心</h2>
        <p>统一管理渠道、用户、利润和平台账单，保持与用户端一致的控制台视觉风格。</p>
      </div>
      <div class="login-box">
        <h3>管理员登录</h3>
        <el-form :model="form" @submit.prevent="handleLogin" label-position="top">
          <el-form-item label="邮箱 / 用户名"><el-input v-model="form.username" /></el-form-item>
          <el-form-item label="密码"><el-input v-model="form.password" type="password" show-password /></el-form-item>
          <el-button type="primary" native-type="submit" :loading="loading" style="width:100%;height:42px">进入后台</el-button>
        </el-form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { authApi } from '@/api/admin'
import { ElMessage } from 'element-plus'

const router = useRouter()
const loading = ref(false)
const form = reactive({ username: '', password: '' })

async function handleLogin() {
  loading.value = true
  try {
    const res = await authApi.login(form)
    if (res.user?.role !== 'admin') {
      ElMessage.error('该账号无管理员权限')
      return
    }
    localStorage.setItem('admin_token', res.token)
    router.push('/admin/dashboard')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page { display:flex;align-items:center;justify-content:center;min-height:100vh;padding:24px }
.login-shell {
  width:min(1040px,100%);
  display:grid;
  grid-template-columns:1.05fr .95fr;
  overflow:hidden;
  border-radius:24px;
  background:#fff;
  border:1px solid #dce7fa;
  box-shadow:0 20px 56px rgba(26,64,135,.15);
}
.hero-panel {
  padding:42px;
  background:linear-gradient(160deg,rgba(30,102,255,.95),rgba(14,197,255,.92));
  color:#fff;
}
.brand {
  display:inline-block;
  margin-bottom:24px;
  padding:6px 12px;
  border-radius:999px;
  background:rgba(255,255,255,.15);
  border:1px solid rgba(255,255,255,.24);
  font-weight:800;
}
.hero-panel h2 { margin:0 0 12px;font-size:2rem;line-height:1.2 }
.hero-panel p { margin:0;color:rgba(255,255,255,.9) }
.login-box { padding:42px }
.login-box h3 { margin:0 0 22px;font-size:1.45rem }
@media (max-width:900px) {
  .login-shell { grid-template-columns:1fr }
  .hero-panel,.login-box { padding:28px }
}
</style>
