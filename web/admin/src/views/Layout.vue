<template>
  <el-container style="height:100vh">
    <el-aside width="200px" class="sidebar">
      <div class="logo">FanAPI Admin</div>
      <el-menu :default-active="route.path" router background-color="#1d2129" text-color="#c9cdd4" active-text-color="#409eff">
        <el-menu-item index="/channels"><el-icon><Connection /></el-icon>渠道管理</el-menu-item>
        <el-menu-item index="/users"><el-icon><User /></el-icon>用户管理</el-menu-item>
        <el-menu-item index="/billing"><el-icon><Tickets /></el-icon>账单流水</el-menu-item>
      </el-menu>
      <div class="sidebar-bottom" @click="logout"><el-icon><SwitchButton /></el-icon>退出</div>
    </el-aside>
    <el-container>
      <el-header class="header">
        <b>{{ pageTitle }}</b>
      </el-header>
      <el-main><router-view /></el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const titles = { '/channels': '渠道管理', '/users': '用户管理', '/billing': '账单流水' }
const pageTitle = computed(() => titles[route.path] ?? 'FanAPI 管理后台')

function logout() {
  localStorage.removeItem('admin_token')
  router.push('/login')
}
</script>

<style scoped>
.sidebar { background:#1d2129;display:flex;flex-direction:column }
.logo { height:60px;line-height:60px;text-align:center;font-size:1.1rem;font-weight:700;color:#fff }
.sidebar-bottom { margin-top:auto;padding:16px;color:#c9cdd4;cursor:pointer;display:flex;align-items:center;gap:8px }
.sidebar-bottom:hover { color:#fff }
.header { display:flex;align-items:center;border-bottom:1px solid #f0f0f0;background:#fff }
</style>
