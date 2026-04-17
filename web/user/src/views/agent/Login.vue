<template>
  <div class="login-page">
    <div class="login-shell">
      <div class="hero-panel">
        <div class="brand">客服工作台</div>
        <h2>客服登录</h2>
        <p>登录后可查看您邀请的用户充值与消费情况，并为用户充值积分。</p>
      </div>
      <div class="login-box">
        <h3>客服登录</h3>
        <el-form :model="form" @submit.prevent="handleLogin" label-position="top">
          <el-form-item label="用户名 / 邮箱">
            <el-input v-model="form.username" />
          </el-form-item>
          <el-form-item label="密码">
            <el-input v-model="form.password" type="password" show-password />
          </el-form-item>
          <el-button type="primary" native-type="submit" :loading="loading" style="width:100%;height:42px">
            进入工作台
          </el-button>
        </el-form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { agentAuthApi } from '@/api/agent'
import { ElMessage } from 'element-plus'

const router = useRouter()
const loading = ref(false)
const form = reactive({ username: '', password: '' })

async function handleLogin() {
  loading.value = true
  try {
    const res = await agentAuthApi.login(form)
    if (res.user?.role !== 'agent' && res.user?.role !== 'admin') {
      ElMessage.error('该账号无客服权限')
      return
    }
    localStorage.setItem('agent_token', res.token)
    localStorage.setItem('agent_username', res.user?.username || '')
    router.push('/agent/dashboard')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page { display:flex;align-items:center;justify-content:center;min-height:100vh;padding:24px }
.login-shell {
  width:min(960px,100%);
  display:grid;
  grid-template-columns:1fr 1fr;
  overflow:hidden;
  border-radius:20px;
  background:#fff;
  border:1px solid #dce7fa;
  box-shadow:0 20px 56px rgba(26,64,135,.12);
}
.hero-panel {
  padding:42px;
  background:linear-gradient(160deg,rgba(37,99,235,.96),rgba(16,185,129,.92));
  color:#fff;
}
.brand {
  display:inline-block;
  margin-bottom:24px;
  padding:5px 12px;
  border-radius:999px;
  background:rgba(255,255,255,.15);
  border:1px solid rgba(255,255,255,.24);
  font-weight:700;
  font-size:.85rem;
}
.hero-panel h2 { margin:0 0 12px;font-size:1.9rem;line-height:1.2 }
.hero-panel p { margin:0;color:rgba(255,255,255,.88);line-height:1.6 }
.login-box { padding:42px }
.login-box h3 { margin:0 0 22px;font-size:1.4rem }
@media (max-width:800px) {
  .login-shell { grid-template-columns:1fr }
  .hero-panel { padding:28px }
  .login-box { padding:28px }
}
</style>
