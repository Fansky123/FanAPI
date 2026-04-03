<template>
  <div class="login-page">
    <div class="login-shell">
      <div class="hero-panel">
        <div class="brand">FanAPI</div>
        <h2>创建你的 API 工作台</h2>
        <p>按量计费、明细可追踪、支持多渠道策略路由。</p>
        <ul>
          <li>兼容 OpenAI SDK 和主流应用</li>
          <li>多模型统一 Key 管理</li>
          <li>实时账单与任务状态可追踪</li>
        </ul>
      </div>
      <div class="login-box">
        <h3>注册账户</h3>
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
          <el-button type="primary" native-type="submit" :loading="loading" style="width:100%;height:42px">
            完成注册
          </el-button>
        </el-form>
        <div class="link-row">
          已有账号？<router-link to="/login">立即登录</router-link>
        </div>
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
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 24px;
}
.login-shell {
  width: min(1080px, 100%);
  display: grid;
  grid-template-columns: 1.1fr .9fr;
  overflow: hidden;
  border-radius: 24px;
  border: 1px solid #dce7fa;
  box-shadow: 0 20px 56px rgba(26, 64, 135, 0.15);
  background: #fff;
}
.hero-panel {
  padding: 42px;
  background:
    linear-gradient(160deg, rgba(10, 74, 214, 0.95), rgba(27, 137, 255, 0.92)),
    #0b4bd4;
  color: #fff;
}
.brand {
  display: inline-block;
  font-weight: 800;
  letter-spacing: .04em;
  margin-bottom: 26px;
  padding: 6px 12px;
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, .3);
  background: rgba(255, 255, 255, .15);
}
.hero-panel h2 {
  margin: 0 0 12px;
  font-size: 2rem;
  line-height: 1.2;
}
.hero-panel p {
  margin: 0;
  color: rgba(255, 255, 255, .9);
}
.hero-panel ul {
  margin: 20px 0 0;
  padding-left: 18px;
  color: rgba(255, 255, 255, .92);
  line-height: 1.9;
}
.login-box {
  padding: 42px;
}
.login-box h3 {
  margin: 0 0 22px;
  font-size: 1.45rem;
}
.link-row {
  margin-top: 16px;
  text-align: center;
  color: #909399;
}

@media (max-width: 900px) {
  .login-shell {
    grid-template-columns: 1fr;
  }
  .hero-panel {
    padding: 28px;
  }
  .login-box {
    padding: 28px;
  }
}
</style>
