<template>
  <el-container style="height:100vh">
    <!-- 侧边栏 -->
    <el-aside width="220px" class="sidebar">
      <div class="logo">FanAPI</div>
      <el-menu
        :default-active="route.path"
        router
        background-color="#1d2129"
        text-color="#c9cdd4"
        active-text-color="#409eff"
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
      </el-menu>
      <div class="sidebar-bottom" @click="logout">
        <el-icon><SwitchButton /></el-icon> 退出登录
      </div>
    </el-aside>

    <!-- 主内容 -->
    <el-container>
      <el-header class="header">
        <span class="page-title">{{ pageTitle }}</span>
        <div class="header-right">
          <el-tag type="success">余额：¥{{ (store.balance / 1e6).toFixed(4) }}</el-tag>
        </div>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'

const route = useRoute()
const router = useRouter()
const store = useUserStore()

const titles = {
  '/dashboard': '概览',
  '/playground': '在线体验',
  '/channels': '渠道列表',
  '/keys': 'API 密钥',
  '/billing': '充值 & 账单',
}
const pageTitle = computed(() => titles[route.path] || 'FanAPI')

onMounted(() => store.fetchBalance())

function logout() {
  store.logout()
  router.push('/login')
}
</script>

<style scoped>
.sidebar {
  background: #1d2129; display: flex; flex-direction: column;
}
.logo {
  height: 60px; line-height: 60px; text-align: center;
  font-size: 1.2rem; font-weight: 700; color: #fff; letter-spacing: 2px;
}
.sidebar-bottom {
  margin-top: auto; padding: 16px 20px; color: #c9cdd4; cursor: pointer;
  display: flex; align-items: center; gap: 8px;
}
.sidebar-bottom:hover { color: #fff; }
.header {
  display: flex; align-items: center; justify-content: space-between;
  border-bottom: 1px solid #f0f0f0; background: #fff;
}
.page-title { font-weight: 600; font-size: 1rem; }
</style>
