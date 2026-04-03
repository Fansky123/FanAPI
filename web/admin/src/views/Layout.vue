<template>
  <el-container class="shell">
    <el-aside width="246px" class="sidebar">
      <div class="logo-wrap">
        <div class="logo-mark">A</div>
        <div>
          <div class="logo">FanAPI Admin</div>
          <div class="logo-sub">Control Panel</div>
        </div>
      </div>
      <el-menu :default-active="route.path" router class="side-menu">
        <el-menu-item index="/channels"><el-icon><Connection /></el-icon>渠道管理</el-menu-item>
        <el-menu-item index="/users"><el-icon><User /></el-icon>用户管理</el-menu-item>
        <el-menu-item index="/billing"><el-icon><Tickets /></el-icon>账单流水</el-menu-item>
        <el-menu-item index="/tasks"><el-icon><Document /></el-icon>任务中心</el-menu-item>
      </el-menu>
      <div class="sidebar-bottom" @click="logout"><el-icon><SwitchButton /></el-icon>退出</div>
    </el-aside>
    <el-container class="content-wrap">
      <el-header class="header">
        <div>
          <div class="page-title">{{ pageTitle }}</div>
          <div class="page-subtitle">平台管理、利润分析与运营控制</div>
        </div>
      </el-header>
      <el-main class="page-main"><router-view /></el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const titles = { '/channels': '渠道管理', '/users': '用户管理', '/billing': '账单流水', '/tasks': '任务中心' }
const pageTitle = computed(() => titles[route.path] ?? 'FanAPI 管理后台')

function logout() {
  localStorage.removeItem('admin_token')
  router.push('/login')
}
</script>

<style scoped>
.shell { min-height:100vh }
.sidebar {
  background:linear-gradient(200deg,#0b1227 0%,#102145 45%,#163575 100%);
  display:flex;
  flex-direction:column;
  padding:16px 14px;
}
.logo-wrap { display:flex;align-items:center;gap:10px;padding:8px 8px 16px }
.logo-mark {
  width:36px;height:36px;border-radius:10px;display:grid;place-items:center;
  font-weight:800;color:#fff;background:linear-gradient(140deg,#1e66ff,#00b4ff)
}
.logo { font-size:1.05rem;font-weight:700;color:#fff }
.logo-sub { color:rgba(255,255,255,.72);font-size:.76rem }
.side-menu { border:none;background:transparent }
:deep(.side-menu .el-menu-item) {
  border-radius:10px;margin:4px 0;color:rgba(232,239,255,.84)
}
:deep(.side-menu .el-menu-item:hover) { background:rgba(255,255,255,.1) }
:deep(.side-menu .el-menu-item.is-active) {
  background:linear-gradient(90deg,rgba(30,102,255,.36),rgba(14,197,255,.24));color:#fff
}
.sidebar-bottom {
  margin-top:auto;padding:14px 12px;color:rgba(235,242,255,.82);cursor:pointer;display:flex;align-items:center;gap:8px;border-radius:10px
}
.sidebar-bottom:hover { background:rgba(255,255,255,.08);color:#fff }
.header {
  display:flex;align-items:center;justify-content:space-between;border-bottom:1px solid #e7edf5;background:rgba(255,255,255,.84);backdrop-filter:blur(8px);padding:0 24px;height:76px
}
.page-title { font-weight:700;font-size:1.1rem }
.page-subtitle { color:#6b7a90;font-size:.82rem;margin-top:3px }
.page-main { padding:22px }
@media (max-width:900px) {
  .shell { display:block }
  .sidebar { width:100% !important;padding:10px }
  .page-main { padding:12px }
}
</style>
