<template>
  <div class="login-page">
    <div class="login-box">
      <h2>登录 FanAPI</h2>
      <el-form :model="form" @submit.prevent="handleLogin" label-position="top">
        <el-form-item label="邮箱">
          <el-input v-model="form.email" placeholder="your@email.com" clearable />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" show-password />
        </el-form-item>
        <el-button type="primary" native-type="submit" :loading="loading" style="width:100%">
          登录
        </el-button>
      </el-form>
      <div class="link-row">
        还没有账号？<router-link to="/register">立即注册</router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { authApi } from '@/api'

const router = useRouter()
const store = useUserStore()
const loading = ref(false)
const form = reactive({ email: '', password: '' })

async function handleLogin() {
  loading.value = true
  try {
    const res = await authApi.login(form)
    store.setToken(res.token)
    router.push('/dashboard')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  display: flex; align-items: center; justify-content: center;
  min-height: 100vh; background: #f5f7fa;
}
.login-box {
  width: 380px; padding: 40px; background: #fff;
  border-radius: 12px; box-shadow: 0 4px 20px rgba(0,0,0,.08);
}
h2 { margin: 0 0 24px; font-size: 1.5rem; color: #303133; }
.link-row { margin-top: 16px; text-align: center; color: #909399; }
</style>
