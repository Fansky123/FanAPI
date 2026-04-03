<template>
  <el-container class="shell">
    <!-- 侧边栏 -->
    <el-aside width="246px" class="sidebar">
      <div class="logo-wrap">
        <div class="logo-mark">F</div>
        <div>
          <div class="logo">FanAPI</div>
          <div class="logo-sub">Control Panel</div>
        </div>
      </div>
      <el-menu
        :default-active="route.path"
        router
        class="side-menu"
      >
        <el-menu-item index="/dashboard">
          <el-icon><Odometer /></el-icon> 概览
        </el-menu-item>
        <el-menu-item index="/playground">
          <el-icon><ChatDotRound /></el-icon> 在线体验
        </el-menu-item>
        <el-menu-item index="/channels">
          <el-icon><Grid /></el-icon> 渠道列表
        </el-menu-item>
        <el-menu-item index="/keys">
          <el-icon><Key /></el-icon> API 密钥
        </el-menu-item>
        <el-menu-item index="/billing">
          <el-icon><Wallet /></el-icon> 充值 & 账单
        </el-menu-item>
        <el-menu-item index="/tasks">
          <el-icon><List /></el-icon> 我的任务
        </el-menu-item>
        <el-menu-item index="/docs">
          <el-icon><Document /></el-icon> 接口文档
        </el-menu-item>
      </el-menu>
      <div class="sidebar-bottom" @click="logout">
        <el-icon><SwitchButton /></el-icon> 退出登录
      </div>
    </el-aside>

    <!-- 主内容 -->
    <el-container class="content-wrap">
      <el-header class="header">
        <div>
          <div class="page-title">{{ pageTitle }}</div>
          <div class="page-subtitle">Chatfire 风格控制台 UI</div>
        </div>
        <div class="header-right">
          <div class="balance-pill">
            <span>余额</span>
            <strong>¥{{ (store.balance / 1e6).toFixed(4) }}</strong>
          </div>
        </div>
      </el-header>
      <el-main class="page-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { Document, List } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const store = useUserStore()

const titles = {
  '/dashboard': '概览',
  '/playground': '在线体验',
  '/channels': '渠道列表',
  '/keys': 'API 密钥',
  '/billing': '充值 & 账单',
  '/docs': '接口文档',
  '/tasks': '我的任务',
}
const pageTitle = computed(() => titles[route.path] || 'FanAPI')

onMounted(() => store.fetchBalance())

function logout() {
  store.logout()
  router.push('/login')
}
</script>

<style scoped>
.shell {
  min-height: 100vh;
}
.sidebar {
  background:
    linear-gradient(200deg, #0b1227 0%, #102145 45%, #163575 100%);
  display: flex;
  flex-direction: column;
  padding: 16px 14px;
  box-shadow: inset -1px 0 0 rgba(255, 255, 255, .06);
}
.logo-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 8px 16px;
}
.logo-mark {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: grid;
  place-items: center;
  font-weight: 800;
  color: #fff;
  background: linear-gradient(140deg, #1e66ff, #00b4ff);
}
.logo {
  font-size: 1.05rem;
  font-weight: 700;
  color: #fff;
  letter-spacing: .02em;
}
.logo-sub {
  color: rgba(255, 255, 255, .72);
  font-size: .76rem;
}
.side-menu {
  border: none;
  background: transparent;
}
:deep(.side-menu .el-menu-item) {
  border-radius: 10px;
  margin: 4px 0;
  color: rgba(232, 239, 255, .84);
}
:deep(.side-menu .el-menu-item:hover) {
  background: rgba(255, 255, 255, .1);
}
:deep(.side-menu .el-menu-item.is-active) {
  background: linear-gradient(90deg, rgba(30, 102, 255, .36), rgba(14, 197, 255, .24));
  color: #fff;
}
.sidebar-bottom {
  margin-top: auto;
  padding: 14px 12px;
  color: rgba(235, 242, 255, .82);
  cursor: pointer;
  display: flex; align-items: center; gap: 8px;
  border-radius: 10px;
}
.sidebar-bottom:hover {
  background: rgba(255, 255, 255, .08);
  color: #fff;
}
.content-wrap {
  min-width: 0;
}
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #e7edf5;
  background: rgba(255, 255, 255, .84);
  backdrop-filter: blur(8px);
  padding: 0 24px;
  height: 76px;
}
.page-title {
  font-weight: 700;
  font-size: 1.1rem;
}
.page-subtitle {
  color: #6b7a90;
  font-size: .82rem;
  margin-top: 3px;
}
.balance-pill {
  border-radius: 999px;
  border: 1px solid #d7e6ff;
  padding: 8px 14px;
  display: flex;
  align-items: baseline;
  gap: 8px;
  background: linear-gradient(90deg, #f2f8ff, #eefaff);
  color: #1248ab;
}
.balance-pill strong {
  font-size: 1rem;
}
.page-main {
  padding: 22px;
}

@media (max-width: 900px) {
  .shell {
    display: block;
  }
  .sidebar {
    width: 100% !important;
    padding: 10px;
  }
  .header {
    padding: 0 12px;
  }
  .page-main {
    padding: 12px;
  }
}
</style>
