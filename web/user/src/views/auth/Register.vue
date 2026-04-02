<template>
  <div class="login-page">
    <div class="login-box">
      <h2>注册 FanAPI</h2>
      <el-form :model="form" @submit.prevent="handleRegister" label-position="top">
        <el-form-item label="邮箱">
          <el-input v-model="form.email" placeholder="your@email.com" clearable />
        </el-form-item>
        <el-form-item label="验证码">
          <div style="display:flex;gap:8px">
            <el-input v-model="form.code" placeholder="6 位验证码" />
            <el-button :disabled="countdown > 0" @click="sendCode">
              {{ countdown > 0 ? `${countdown}s` : '获取验证码' }}
            </el-button>
          </div>
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" show-password placeholder="至少 8 位" />
        </el-form-item>
        <el-button type="primary" native-type="submit" :loading="loading" style="width:100%">
          注册
        </el-button>
      </el-form>
      <div class="link-row">
        已有账号？<router-link to="/login">立即登录</router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { authApi } from '@/api'
import { ElMessage } from 'element-plus'

const router = useRouter()
const store = useUserStore()
const loading = ref(false)
const countdown = ref(0)
const form = reactive({ email: '', code: '', password: '' })

async function sendCode() {
  if (!form.email) return ElMessage.warning('请先填写邮箱')
  await authApi.sendCode(form.email)
  ElMessage.success('验证码已发送')
  countdown.value = 60
  const t = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) clearInterval(t)
  }, 1000)
}

async function handleRegister() {
  loading.value = true
  try {
    const res = await authApi.register(form)
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
  width: 420px; padding: 40px; background: #fff;
  border-radius: 12px; box-shadow: 0 4px 20px rgba(0,0,0,.08);
}
h2 { margin: 0 0 24px; font-size: 1.5rem; color: #303133; }
.link-row { margin-top: 16px; text-align: center; color: #909399; }
</style>
