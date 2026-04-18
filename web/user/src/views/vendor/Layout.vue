<template>
  <div class="vendor-shell">
    <header class="topbar">
      <div class="brand">
        <div class="brand-mark">VD</div>
        <span class="brand-name">号商工作台</span>
      </div>
      <nav class="topbar-nav">
        <router-link to="/vendor/dashboard" class="nav-link">仪表板</router-link>
        <router-link to="/vendor/keys" class="nav-link">我的 Key</router-link>
      </nav>
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
const username = ref(localStorage.getItem('vendor_username') || '号商')

function logout() {
  localStorage.removeItem('vendor_token')
  localStorage.removeItem('vendor_username')
  router.push('/vendor/login')
}
</script>

<style scoped>
.vendor-shell {
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
  background: linear-gradient(135deg, #7c3aed, #3b82f6);
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
.topbar-nav {
  display: flex;
  gap: 4px;
  flex: 1;
  justify-content: center;
}
.nav-link {
  padding: 6px 14px;
  border-radius: 6px;
  font-size: 14px;
  color: rgba(255,255,255,.7);
  text-decoration: none;
  transition: all .18s;
}
.nav-link:hover, .nav-link.router-link-active {
  color: #fff;
  background: rgba(255,255,255,.12);
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
.topbar-right :deep(.el-button) { color: rgba(255,255,255,.7); }
.topbar-right :deep(.el-button:hover) { color: #fff; }
.page-main {
  flex: 1;
  max-width: 1100px;
  width: 100%;
  margin: 0 auto;
  padding: 32px 20px;
}
</style>
