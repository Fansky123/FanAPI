<template>
  <div class="login-page">
    <div class="login-shell">
      <div class="hero-panel">
        <div class="brand">FanAPI</div>
        <h2>多模型统一接入控制台</h2>
        <p>稳定接入 LLM / 图像 / 视频 / 音频，透明计费，极速上线。</p>
        <div class="hero-grid">
          <div class="hero-item">
            <div class="v">99.9%</div>
            <div class="k">可用性</div>
          </div>
          <div class="hero-item">
            <div class="v">OpenAI 兼容</div>
            <div class="k">标准协议</div>
          </div>
        </div>
      </div>
      <div class="login-box">
        <h3>登录账户</h3>
        <el-form :model="form" @submit.prevent="handleLogin" label-position="top">
          <el-form-item label="邮箱">
            <el-input v-model="form.email" placeholder="your@email.com" clearable />
          </el-form-item>
          <el-form-item label="密码">
            <el-input v-model="form.password" type="password" show-password />
          </el-form-item>
          <el-button type="primary" native-type="submit" :loading="loading" style="width:100%;height:42px">
            进入控制台
          </el-button>
        </el-form>
        <div class="link-row">
          还没有账号？<router-link to="/register">立即注册</router-link>
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
    linear-gradient(160deg, rgba(30, 102, 255, 0.95), rgba(14, 197, 255, 0.92)),
    #1e66ff;
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
.hero-grid {
  margin-top: 32px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}
.hero-item {
  background: rgba(255, 255, 255, .12);
  border: 1px solid rgba(255, 255, 255, .25);
  border-radius: 14px;
  padding: 14px;
}
.hero-item .v {
  font-size: 1.15rem;
  font-weight: 700;
}
.hero-item .k {
  opacity: .85;
  margin-top: 6px;
  font-size: .88rem;
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
