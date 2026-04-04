<template>
  <div class="app-shell">
    <!-- 侧边栏 -->
    <aside class="sidebar">
      <div class="brand">
        <div class="brand-icon">F</div>
        <span class="brand-name">FanAPI</span>
      </div>

      <nav class="nav">
        <router-link v-for="item in navItems" :key="item.to" :to="item.to" class="nav-item" :class="{ active: route.path === item.to }">
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.label }}</span>
        </router-link>
      </nav>

      <div class="sidebar-footer">
        <div class="balance-mini">
          <span class="balance-label">余额</span>
          <span class="balance-val">¥{{ (store.balance / 1e6).toFixed(4) }}</span>
        </div>
        <div class="logout-btn" @click="logout">
          <el-icon><SwitchButton /></el-icon>
          <span>退出</span>
        </div>
      </div>
    </aside>

    <!-- 主区域 -->
    <div class="main-area">
      <!-- 顶部栏 -->
      <header class="topbar">
        <div class="topbar-left">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item>FanAPI</el-breadcrumb-item>
            <el-breadcrumb-item>{{ pageTitle }}</el-breadcrumb-item>
          </el-breadcrumb>
          <h1 class="page-title">{{ pageTitle }}</h1>
        </div>
        <div class="topbar-right">
          <el-tag size="large" type="info" effect="plain" class="balance-tag" @click="router.push('/billing')" style="cursor:pointer">
            <el-icon><Wallet /></el-icon>
            余额 ¥{{ (store.balance / 1e6).toFixed(4) }}
          </el-tag>
          <el-dropdown @command="handleCmd">
            <div class="avatar-btn">
              <div class="avatar-circle">{{ userInitial }}</div>
              <el-icon><ArrowDown /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="billing">钱包 & 账单</el-dropdown-item>
                <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>

      <!-- 内容 -->
      <main class="content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import {
  ChatDotRound, Grid, Key, Wallet, List, Document, SwitchButton, ArrowDown
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const store = useUserStore()

const navItems = [
  { to: '/playground', label: '在线体验',  icon: 'ChatDotRound' },
  { to: '/models',     label: '模型列表',  icon: 'Grid' },
  { to: '/keys',       label: 'API 密钥', icon: 'Key' },
  { to: '/billing',    label: '钱包 & 账单', icon: 'Wallet' },
  { to: '/tasks',      label: '任务日志',  icon: 'List' },
  { to: '/docs',       label: '接口文档',  icon: 'Document' },
]

const titles = {
  '/playground': '在线体验',
  '/models':     '模型列表',
  '/keys':       'API 密钥',
  '/billing':    '钱包 & 账单',
  '/docs':       '接口文档',
  '/tasks':      '任务日志',
}
const pageTitle = computed(() => titles[route.path] ?? 'FanAPI')
const userInitial = computed(() => {
  const email = store.email || localStorage.getItem('user_email') || ''
  return email.charAt(0).toUpperCase() || 'U'
})

onMounted(() => store.fetchBalance())

function logout() {
  store.logout()
  router.push('/login')
}

function handleCmd(cmd) {
  if (cmd === 'logout') logout()
  else if (cmd === 'billing') router.push('/billing')
}
</script>

<style scoped>
/* ---- 整体布局 ---- */
.app-shell {
  display: flex;
  min-height: 100vh;
  background: #f4f6fb;
}

/* ---- 侧边栏 ---- */
.sidebar {
  width: 220px;
  flex-shrink: 0;
  background: #0d1526;
  display: flex;
  flex-direction: column;
  padding: 0;
}

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 20px 20px 16px;
  border-bottom: 1px solid rgba(255,255,255,.06);
}
.brand-icon {
  width: 30px; height: 30px;
  border-radius: 8px;
  background: #1677ff;
  display: grid; place-items: center;
  font-weight: 800; color: #fff; font-size: .9rem;
}
.brand-name {
  font-weight: 700;
  color: #fff;
  font-size: .95rem;
  letter-spacing: .02em;
}

.nav {
  flex: 1;
  padding: 12px 10px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 12px;
  border-radius: 8px;
  color: rgba(255,255,255,.6);
  text-decoration: none;
  font-size: .875rem;
  transition: all .15s;
}
.nav-item:hover {
  background: rgba(255,255,255,.08);
  color: rgba(255,255,255,.9);
}
.nav-item.active {
  background: rgba(22,119,255,.25);
  color: #5ba4ff;
}
.nav-item .el-icon { font-size: 1rem; flex-shrink: 0; }

.sidebar-footer {
  padding: 12px 10px;
  border-top: 1px solid rgba(255,255,255,.06);
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.balance-mini {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: rgba(255,255,255,.05);
  border-radius: 8px;
}
.balance-label { color: rgba(255,255,255,.45); font-size: .75rem; }
.balance-val { color: #5ba4ff; font-weight: 600; font-size: .85rem; }
.logout-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 8px;
  color: rgba(255,255,255,.45);
  cursor: pointer;
  font-size: .85rem;
  transition: all .15s;
}
.logout-btn:hover { background: rgba(255,100,100,.15); color: #ff7875; }

/* ---- 主区域 ---- */
.main-area {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

/* ---- 顶部栏 ---- */
.topbar {
  height: 56px;
  background: #fff;
  border-bottom: 1px solid #e8ecf4;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  flex-shrink: 0;
}
.topbar-left { display: flex; flex-direction: column; gap: 1px; }
.page-title { font-size: 1rem; font-weight: 600; color: #0d1526; margin: 0; line-height: 1.2; }
:deep(.el-breadcrumb) { font-size: .72rem; opacity: .55; }
.topbar-right { display: flex; align-items: center; gap: 12px; }
.balance-tag { font-weight: 600; gap: 4px; }
.avatar-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 999px;
  color: #454f63;
}
.avatar-btn:hover { background: #f0f2f7; }
.avatar-circle {
  width: 28px; height: 28px;
  border-radius: 50%;
  background: #1677ff;
  color: #fff;
  display: grid; place-items: center;
  font-size: .8rem;
  font-weight: 700;
}

/* ---- 内容区 ---- */
.content {
  flex: 1;
  padding: 20px 24px;
  overflow-y: auto;
}

@media (max-width: 768px) {
  .app-shell { flex-direction: column; }
  .sidebar { width: 100%; flex-direction: row; flex-wrap: wrap; height: auto; padding: 8px; }
  .nav { flex-direction: row; padding: 0; flex-wrap: wrap; }
  .sidebar-footer { display: none; }
}
</style>
