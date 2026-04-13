<template>
  <div class="ow-shell" :class="{ collapsed: sidebarCollapsed }">
    <!-- 自定义页眉 -->
    <div v-if="site.headerHtml" class="ow-custom-header" v-html="site.headerHtml"></div>

    <!-- 侧边栏 -->
    <aside class="ow-sidebar">
      <!-- Logo -->
      <div class="ow-logo">
        <img v-if="site.logoUrl && !logoImgErr" :src="site.logoUrl" class="logo-img" alt="logo" @error="logoImgErr = true" />
        <div v-else class="logo-icon">{{ site.siteName.charAt(0).toUpperCase() }}</div>
        <span class="logo-title" v-show="!sidebarCollapsed">{{ site.siteName }}</span>
      </div>

      <!-- 导航 -->
      <nav class="ow-nav">
        <router-link to="/dashboard" class="ow-nav-item" :class="{ active: route.path === '/dashboard' }">
          <el-icon><DataBoard /></el-icon>
          <span v-show="!sidebarCollapsed">数据看板</span>
        </router-link>
        <router-link to="/models" class="ow-nav-item" :class="{ active: route.path === '/models' }">
          <el-icon><Grid /></el-icon>
          <span v-show="!sidebarCollapsed">模型列表</span>
        </router-link>
        <router-link to="/docs" class="ow-nav-item" :class="{ active: route.path === '/docs' }">
          <el-icon><Document /></el-icon>
          <span v-show="!sidebarCollapsed">接口文档</span>
        </router-link>
        <router-link to="/tasks" class="ow-nav-item" :class="{ active: route.path === '/tasks' || route.path === '/llm-logs' }">
          <el-icon><List /></el-icon>
          <span v-show="!sidebarCollapsed">调用日志</span>
        </router-link>
        <template v-if="isLoggedIn">
          <div class="ow-nav-section" v-show="!sidebarCollapsed">在线体验</div>
          <router-link to="/playground" class="ow-nav-item" :class="{ active: route.path === '/playground' }">
            <el-icon><ChatDotRound /></el-icon>
            <span v-show="!sidebarCollapsed">文本对话</span>
          </router-link>
          <div class="ow-nav-section" v-show="!sidebarCollapsed">账户管理</div>
          <router-link to="/keys" class="ow-nav-item" :class="{ active: route.path === '/keys' }">
            <el-icon><Key /></el-icon>
            <span v-show="!sidebarCollapsed">API 密钥</span>
          </router-link>
          <router-link to="/recharge" class="ow-nav-item" :class="{ active: route.path === '/recharge' }">
            <el-icon><CreditCard /></el-icon>
            <span v-show="!sidebarCollapsed">充值积分</span>
          </router-link>
          <router-link to="/billing" class="ow-nav-item" :class="{ active: route.path === '/billing' }">
            <el-icon><Wallet /></el-icon>
            <span v-show="!sidebarCollapsed">我的订单</span>
          </router-link>
        </template>
      </nav>

      <!-- 侧边栏底部 -->
      <div class="ow-sidebar-footer" v-show="!sidebarCollapsed">
        <template v-if="isLoggedIn">
          <div class="balance-row" @click="router.push('/billing')">
            <el-icon><Wallet /></el-icon>
            <span>¥{{ (store.balance / 1e6).toFixed(4) }}</span>
          </div>
          <div class="logout-row" @click="logout">
            <el-icon><SwitchButton /></el-icon>
            <span>退出登录</span>
          </div>
        </template>
        <template v-else>
          <router-link to="/login" class="ow-footer-btn primary">登录</router-link>
          <router-link to="/register" class="ow-footer-btn outline">注册</router-link>
        </template>
      </div>

      <!-- 折叠按钮 -->
      <div class="ow-collapse-btn" @click="sidebarCollapsed = !sidebarCollapsed">
        <el-icon><DArrowLeft v-if="!sidebarCollapsed" /><DArrowRight v-else /></el-icon>
      </div>
    </aside>

    <!-- 主区域 -->
    <div class="ow-main">
      <!-- 顶部 Header -->
      <header class="ow-header">
        <div class="header-left">
          <span class="header-page">{{ pageTitle }}</span>
        </div>
        <div class="header-right">
          <template v-if="isLoggedIn">
            <div class="balance-chip" @click="router.push('/billing')">
              <el-icon><Wallet /></el-icon>
              <span>¥{{ (store.balance / 1e6).toFixed(4) }}</span>
            </div>
            <el-dropdown @command="handleCmd" trigger="click">
              <div class="avatar-btn">
                <div class="avatar-circle">{{ userInitial }}</div>
                <span class="avatar-name">{{ store.username || store.email || '用户' }}</span>
                <el-icon style="font-size:12px;opacity:.6"><ArrowDown /></el-icon>
              </div>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="billing">
                    <el-icon><Wallet /></el-icon> 钱包 &amp; 账单
                  </el-dropdown-item>
                  <el-dropdown-item command="logout" divided>
                    <el-icon><SwitchButton /></el-icon> 退出登录
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
          <template v-else>
            <router-link to="/login">
              <el-button size="small">登录</el-button>
            </router-link>
            <router-link to="/register">
              <el-button type="primary" size="small">免费注册</el-button>
            </router-link>
          </template>
         
        </div>
      </header>

      <!-- 内容 -->
      <main class="ow-content" :class="{ 'ow-content--flush': route.path === '/docs' }">
        <div v-if="site.headerHtml" style="display:none"></div>
        <router-view />
      </main>

      <!-- 自定义页脚 -->
      <footer v-if="site.footerHtml" class="ow-custom-footer" v-html="site.footerHtml"></footer>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useSiteStore } from '@/stores/site'
import {
  ChatDotRound, Grid, Key, Wallet, List, Document, SwitchButton, ArrowDown,
  ChatLineSquare, DArrowLeft, DArrowRight, DataBoard, CreditCard, Moon, Sunny
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const store = useUserStore()
const site = useSiteStore()

const logoImgErr = ref(false)
const sidebarCollapsed = ref(false)

const isLoggedIn = computed(() => !!store.token)

const titles = {
  '/dashboard': '数据看板',
  '/playground': '文本对话',
  '/models':     '模型列表',
  '/keys':       'API 密钥',
  '/recharge':   '充值积分',
  '/billing':    '我的订单',
  '/docs':       '接口文档',
  '/tasks':      '调用日志',
  '/llm-logs':   '调用日志',
}
const pageTitle = computed(() => titles[route.path] ?? site.siteName)
const userInitial = computed(() => {
  const name = store.username || localStorage.getItem('user_username') || store.email || 'U'
  return name.charAt(0).toUpperCase()
})

onMounted(() => {
  site.fetchSettings()
  if (isLoggedIn.value) {
    store.fetchBalance()
    store.fetchProfile()
  }
})

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
/* ── Shell ── */
.ow-shell {
  display: flex;
  min-height: 100vh;
}

/* ── 自定义页眉 ── */
.ow-custom-header { width: 100%; order: -1; }

/* ── 侧边栏 ── */
.ow-sidebar {
  width: var(--ow-sidebar-w, 220px);
  flex-shrink: 0;
  background: var(--ow-layout-sidebar);
  border-right: 1px solid var(--ow-layout-border);
  display: flex;
  flex-direction: column;
  position: fixed;
  top: 0; left: 0;
  height: 100vh;
  z-index: 100;
  transition: width .25s ease;
  overflow: hidden;
}

.ow-shell.collapsed .ow-sidebar {
  width: 60px;
}
.ow-shell.collapsed .ow-main {
  margin-left: 60px;
}

/* ── Logo ── */
.ow-logo {
  height: var(--ow-header-h, 48px);
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 16px;
  border-bottom: 1px solid var(--ow-layout-border);
  flex-shrink: 0;
  overflow: hidden;
  white-space: nowrap;
}
.logo-img {
  width: 26px; height: 26px;
  border-radius: 6px;
  object-fit: contain;
  flex-shrink: 0;
}
.logo-icon {
  width: 26px; height: 26px;
  border-radius: 6px;
  background: var(--ow-primary, #165dff);
  color: #fff;
  display: grid; place-items: center;
  font-weight: 700; font-size: 13px;
  flex-shrink: 0;
}
.logo-title {
  font-weight: 700;
  font-size: 15px;
  color: var(--ow-text, #1d2129);
  letter-spacing: .01em;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* ── Nav ── */
.ow-nav {
  flex: 1;
  padding: 8px 8px;
  overflow-y: auto;
  overflow-x: hidden;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.ow-nav::-webkit-scrollbar { width: 0; }
.ow-nav-section {
  padding: 12px 12px 4px;
  font-size: 11px;
  font-weight: 600;
  color: var(--ow-subtext, #86909c);
  text-transform: uppercase;
  letter-spacing: .08em;
  user-select: none;
}

.ow-nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 12px;
  height: 40px;
  border-radius: var(--ow-radius, 4px);
  color: var(--ow-nav-color);
  text-decoration: none;
  font-size: 14px;
  font-weight: 400;
  transition: all .15s;
  position: relative;
  white-space: nowrap;
  cursor: pointer;
}
.ow-nav-item .el-icon {
  font-size: 16px;
  flex-shrink: 0;
  color: inherit;
}
.ow-nav-item:hover {
  background: var(--ow-nav-hover);
  color: var(--ow-primary, #165dff);
}
.ow-nav-item.active {
  background: var(--ow-nav-active);
  color: var(--ow-primary, #165dff);
  font-weight: 600;
}
.ow-nav-item.active::before {
  content: '';
  position: absolute;
  left: 0; top: 25%; bottom: 25%;
  width: 3px;
  background: var(--ow-primary, #165dff);
  border-radius: 0 3px 3px 0;
}

/* ── Sidebar Footer ── */
.ow-sidebar-footer {
  padding: 10px 8px;
  border-top: 1px solid var(--ow-layout-border);
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.balance-row, .logout-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 12px;
  height: 36px;
  border-radius: var(--ow-radius, 4px);
  font-size: 13px;
  cursor: pointer;
  transition: all .15s;
  color: var(--ow-nav-color);
}
.balance-row:hover { background: var(--ow-nav-hover); color: var(--ow-primary); }
.logout-row:hover { background: var(--ow-nav-hover); color: var(--ow-danger, #f53f3f); }

.ow-footer-btn {
  display: block;
  text-align: center;
  padding: 6px 12px;
  border-radius: var(--ow-radius, 4px);
  font-size: 13px;
  font-weight: 600;
  text-decoration: none;
  transition: all .15s;
}
.ow-footer-btn.primary {
  background: var(--ow-primary, #165dff);
  color: #fff;
}
.ow-footer-btn.primary:hover { background: var(--ow-primary-hover, #4080ff); }
.ow-footer-btn.outline {
  border: 1px solid var(--ow-border, #e5e6eb);
  color: var(--ow-nav-color);
}
.ow-footer-btn.outline:hover { border-color: var(--ow-primary); color: var(--ow-primary); }

/* ── Collapse button ── */
.ow-collapse-btn {
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: var(--ow-subtext);
  border-top: 1px solid var(--ow-layout-border);
  flex-shrink: 0;
  transition: color .15s;
}
.ow-collapse-btn:hover { color: var(--ow-primary); }

/* ── 主区域 ── */
.ow-main {
  flex: 1;
  margin-left: var(--ow-sidebar-w, 220px);
  display: flex;
  flex-direction: column;
  min-width: 0;
  transition: margin-left .25s ease;
}

/* ── Header ── */
.ow-header {
  height: var(--ow-header-h, 48px);
  background: var(--ow-layout-sidebar);
  border-bottom: 1px solid var(--ow-layout-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  flex-shrink: 0;
  position: sticky; top: 0; z-index: 50;
}
.header-left {}
.header-page {
  color: var(--ow-text, #1d2129);
  font-size: 14px;
  font-weight: 600;
}
.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}
.balance-chip {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 4px 12px;
  border-radius: 999px;
  background: var(--ow-chip-bg);
  color: var(--ow-primary, #165dff);
  font-size: 13px;
  cursor: pointer;
  transition: background .15s;
  font-weight: 500;
  border: 1px solid var(--ow-layout-border);
}
.balance-chip:hover { background: var(--ow-nav-active); }
.avatar-btn {
  display: flex;
  align-items: center;
  gap: 7px;
  padding: 4px 10px;
  border-radius: 6px;
  cursor: pointer;
  color: var(--ow-text, #1d2129);
  transition: background .15s;
}
.avatar-btn:hover { background: var(--ow-chip-bg); }
.avatar-circle {
  width: 26px; height: 26px;
  border-radius: 50%;
  background: var(--ow-primary, #165dff);
  color: #fff;
  display: grid; place-items: center;
  font-size: 12px;
  font-weight: 700;
  flex-shrink: 0;
}
.avatar-name {
  font-size: 13px;
  color: var(--ow-text, #1d2129);
  max-width: 100px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* ── 内容 ── */
.ow-content {
  flex: 1;
  padding: 20px 24px;
  overflow-y: auto;
  background: var(--ow-content-bg);
}
.ow-content--flush {
  padding: 0;
  overflow: hidden;
}

/* 深色模式切换按鈕 */
.theme-toggle {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  display: grid;
  place-items: center;
  cursor: pointer;
  color: var(--ow-nav-color);
  transition: background .15s, color .15s;
}
.theme-toggle:hover {
  background: var(--ow-nav-hover);
  color: var(--ow-primary);
}

/* ── 自定义页脚 ── */
.ow-custom-footer { padding: 12px 24px; }

/* ── Mobile ── */
@media (max-width: 768px) {
  .ow-sidebar {
    transform: translateX(-100%);
    width: var(--ow-sidebar-w, 220px) !important;
  }
  .ow-main {
    margin-left: 0 !important;
  }
  .ow-collapse-btn { display: none; }
}
</style>
