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
          <el-form-item label="用户名">
            <el-input v-model="form.username" placeholder="3-32 个字符" clearable />
          </el-form-item>
          <el-form-item label="密码">
            <el-input v-model="form.password" type="password" show-password placeholder="至少 8 位" />
          </el-form-item>
          <el-button type="primary" native-type="submit" :loading="loading" style="width:100%;height:42px">
            完成注册
          </el-button>
        </el-form>
        <div class="hint-tip">
          <el-icon><InfoFilled /></el-icon>
          注册后可在账户设置中绑定邮箱，以便忘记密码时通过邮箱找回。
        </div>
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
import { InfoFilled } from '@element-plus/icons-vue'

const router = useRouter()
const store = useUserStore()
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
  background: linear-gradient(160deg, rgba(10, 74, 214, 0.95), rgba(27, 137, 255, 0.92)), #0b4bd4;
  color: #fff;
}
.brand {
  display: inline-block;
  font-weight: 800;
  letter-spacing: .04em;
  margin-bottom: 26px;
  padding: 6px 12px;
  border-radius: 999px;
  border: 1px solid rgba(255,255,255,.3);
  background: rgba(255,255,255,.15);
}
.hero-panel h2 { margin: 0 0 12px; font-size: 2rem; line-height: 1.2; }
.hero-panel p { margin: 0; color: rgba(255,255,255,.9); }
.hero-panel ul { margin: 20px 0 0; padding-left: 18px; color: rgba(255,255,255,.92); line-height: 1.9; }
.login-box { padding: 42px; }
.login-box h3 { margin: 0 0 22px; font-size: 1.45rem; }
.hint-tip {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  margin-top: 14px;
  padding: 10px 12px;
  background: #f0f7ff;
  border-radius: 8px;
  color: #4a7fd6;
  font-size: .82rem;
  line-height: 1.5;
}
.link-row { margin-top: 16px; text-align: center; color: #909399; }
@media (max-width: 900px) {
  .login-shell { grid-template-columns: 1fr; }
  .hero-panel { padding: 28px; }
  .login-box { padding: 28px; }
}
</style>

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
