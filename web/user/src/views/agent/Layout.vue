<template>
  <div class="agent-shell">
    <header class="topbar">
      <div class="brand">
        <div class="brand-mark">CS</div>
        <span class="brand-name">客服工作台</span>
      </div>
      <div class="topbar-right">
        <span class="username-tag">{{ username }}</span>
        <el-button size="small" text @click="logout">
          <el-icon><SwitchButton /></el-icon> 退出
        </el-button>
      </div>
    </header>
    <main class="page-main">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { SwitchButton } from '@element-plus/icons-vue'

const router = useRouter()
const username = ref(localStorage.getItem('agent_username') || '客服')

function logout() {
  localStorage.removeItem('agent_token')
  localStorage.removeItem('agent_username')
  router.push('/agent/login')
}
</script>

<style scoped>
.agent-shell {
  min-height: 100vh;
  background: #f4f6fb;
  display: flex;
  flex-direction: column;
}
.topbar {
  background: #1a1d2e;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 28px;
  position: sticky;
  top: 0;
  z-index: 100;
  box-shadow: 0 2px 8px rgba(0,0,0,.25);
}
.brand {
  display: flex;
  align-items: center;
  gap: 10px;
}
.brand-mark {
  width: 32px; height: 32px;
  border-radius: 8px;
  background: linear-gradient(135deg, #10b981, #3b82f6);
  color: #fff;
  font-weight: 800;
  font-size: 12px;
  display: grid;
  place-items: center;
}
.brand-name {
  font-size: 15px;
  font-weight: 700;
  color: #fff;
  letter-spacing: .02em;
}
.topbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}
.username-tag {
  font-size: 13px;
  color: rgba(255,255,255,.7);
}
.topbar-right :deep(.el-button) {
  color: rgba(255,255,255,.7);
}
.topbar-right :deep(.el-button:hover) {
  color: #fff;
}
.page-main {
  flex: 1;
  max-width: 960px;
  width: 100%;
  margin: 0 auto;
  padding: 32px 20px;
}
</style>
