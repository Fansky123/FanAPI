<template>
  <div class="ow-shell">
    <div v-if="site.headerHtml" class="ow-custom-header" v-html="site.headerHtml"></div>

    <aside class="ow-sidebar">
      <!-- Logo -->
      <div class="ow-logo">
        <img v-if="site.logoUrl && !logoImgErr" :src="site.logoUrl" class="logo-img" alt="logo" @error="logoImgErr = true" />
        <div v-else class="logo-icon">{{ site.siteName.charAt(0).toUpperCase() }}</div>
        <span class="logo-title">{{ site.siteName }}</span>
      </div>

      <!-- 导航 -->
      <nav class="ow-nav">
        <router-link to="/dashboard" class="ow-nav-item" :class="{ active: route.path === '/dashboard' }">
          <div class="menu-item"><span class="menu-item-title">数据看板</span></div>
        </router-link>
        <router-link to="/models" class="ow-nav-item" :class="{ active: route.path === '/models' }">
          <div class="menu-item"><span class="menu-item-title">模型列表</span></div>
        </router-link>
        <router-link to="/tasks" class="ow-nav-item" :class="{ active: route.path === '/tasks' || route.path === '/llm-logs' }">
          <div class="menu-item"><span class="menu-item-title">调用日志</span></div>
        </router-link>
        <router-link to="/stats" class="ow-nav-item" :class="{ active: route.path === '/stats' }">
          <div class="menu-item"><span class="menu-item-title">使用统计</span></div>
        </router-link>
        <router-link to="/docs" class="ow-nav-item" :class="{ active: route.path === '/docs' }">
          <div class="menu-item"><span class="menu-item-title">接口文档</span></div>
        </router-link>

        <template v-if="isLoggedIn">
          <div class="ow-nav-section">在线体验</div>
          <router-link to="/playground" class="ow-nav-item" :class="{ active: route.path === '/playground' }">
            <div class="menu-item"><span class="menu-item-title">文本对话</span></div>
          </router-link>
          <router-link to="/image-gen" class="ow-nav-item" :class="{ active: route.path === '/image-gen' }">
            <div class="menu-item"><span class="menu-item-title">图片生成</span></div>
          </router-link>
          <router-link to="/video-gen" class="ow-nav-item" :class="{ active: route.path === '/video-gen' }">
            <div class="menu-item"><span class="menu-item-title">视频生成</span></div>
          </router-link>

          <div class="ow-nav-section">账户管理</div>
          <router-link to="/keys" class="ow-nav-item" :class="{ active: route.path === '/keys' }">
            <div class="menu-item"><span class="menu-item-title">API 密钥</span></div>
          </router-link>
          <router-link to="/recharge" class="ow-nav-item" :class="{ active: route.path === '/recharge' }">
            <div class="menu-item"><span class="menu-item-title">积分充值</span></div>
          </router-link>
          <router-link to="/exchange" class="ow-nav-item" :class="{ active: route.path === '/exchange' }">
            <div class="menu-item"><span class="menu-item-title">兑换中心</span></div>
          </router-link>
          <router-link to="/billing" class="ow-nav-item" :class="{ active: route.path === '/billing' }">
            <div class="menu-item"><span class="menu-item-title">我的订单</span></div>
          </router-link>
          <router-link to="/profile" class="ow-nav-item" :class="{ active: route.path === '/profile' }">
            <div class="menu-item"><span class="menu-item-title">个人中心</span></div>
          </router-link>
          <router-link to="/invite" class="ow-nav-item" :class="{ active: route.path === '/invite' }">
            <div class="menu-item"><span class="menu-item-title">邀请中心</span></div>
          </router-link>
        </template>
      </nav>

      <!-- 侧边栏底部 -->
      <div class="ow-sidebar-footer">
        <template v-if="isLoggedIn">
          <div class="balance-row" @click="router.push('/billing')">
            <el-icon><Wallet /></el-icon>
            <span>{{ (store.balance / 1e6).toFixed(4) }}</span>
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
    </aside>

    <!-- 主区域 -->
    <div class="ow-main">
      <header class="ow-header">
        <div class="header-left">
          <span class="header-page">{{ pageTitle }}</span>
        </div>
        <div class="header-right">
          <el-popover v-if="site.qqGroupUrl" placement="bottom" :width="220" trigger="click">
            <template #reference>
              <div class="header-contact-btn qq-btn">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor" style="flex-shrink:0"><path d="M12 2C6.477 2 2 6.477 2 12c0 1.89.527 3.66 1.438 5.168L2 22l4.832-1.438A9.956 9.956 0 0 0 12 22c5.523 0 10-4.477 10-10S17.523 2 12 2zm0 2c4.418 0 8 3.582 8 8s-3.582 8-8 8a7.95 7.95 0 0 1-3.868-1l-.31-.183-3.2.952.952-3.2-.183-.31A7.95 7.95 0 0 1 4 12c0-4.418 3.582-8 8-8zm-1 4v2h2V8h-2zm0 4v6h2v-6h-2z"/></svg>
                QQ交流群
              </div>
            </template>
            <div style="text-align:center;padding:4px 0">
              <img :src="site.qqGroupUrl" style="width:180px;height:180px;object-fit:contain;border-radius:6px" alt="QQ交流群二维码" />
              <div style="margin-top:8px;font-size:13px;color:#606266">扫码加入 QQ 交流群</div>
            </div>
          </el-popover>
          <el-popover v-if="site.wechatCsUrl" placement="bottom" :width="220" trigger="click">
            <template #reference>
              <div class="header-contact-btn wechat-btn">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor" style="flex-shrink:0"><path d="M8.691 2.188C3.891 2.188 0 5.476 0 9.53c0 2.212 1.17 4.203 3.002 5.55a.59.59 0 0 1 .213.665l-.39 1.48c-.019.07-.048.141-.048.213 0 .163.13.295.29.295a.326.326 0 0 0 .167-.054l1.903-1.114a.864.864 0 0 1 .717-.098 10.16 10.16 0 0 0 2.837.403c.276 0 .543-.027.811-.05-.857-2.578.157-4.972 1.932-6.446 1.703-1.415 3.882-1.98 5.853-1.838-.576-3.583-4.196-6.348-8.596-6.348zM5.785 5.991c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 0 1-1.162 1.178A1.17 1.17 0 0 1 4.623 7.17c0-.651.52-1.18 1.162-1.18zm5.813 0c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 0 1-1.162 1.178 1.17 1.17 0 0 1-1.162-1.178c0-.651.52-1.18 1.162-1.18zm5.34 2.867c-1.797-.052-3.746.512-5.28 1.786-1.72 1.428-2.687 3.72-1.78 6.22.942 2.453 3.666 4.229 6.884 4.229.826 0 1.622-.12 2.361-.336a.722.722 0 0 1 .598.082l1.584.926a.272.272 0 0 0 .14.047c.134 0 .24-.111.24-.247 0-.06-.023-.12-.038-.177l-.327-1.233a.582.582 0 0 1-.023-.156.49.49 0 0 1 .201-.398C23.024 18.48 24 16.82 24 14.98c0-3.21-2.931-5.837-6.656-6.088V8.89c-.135-.01-.27-.027-.407-.03zm-2.53 3.274c.535 0 .969.44.969.982a.976.976 0 0 1-.969.983.976.976 0 0 1-.969-.983c0-.542.434-.982.97-.982zm4.844 0c.535 0 .969.44.969.982a.976.976 0 0 1-.969.983.976.976 0 0 1-.969-.983c0-.542.434-.982.969-.982z"/></svg>
                微信客服
              </div>
            </template>
            <div style="text-align:center;padding:4px 0">
              <img :src="site.wechatCsUrl" style="width:180px;height:180px;object-fit:contain;border-radius:6px" alt="微信客服二维码" />
              <div style="margin-top:8px;font-size:13px;color:#606266">扫码添加微信客服</div>
            </div>
          </el-popover>
          <template v-if="isLoggedIn">
            <div class="balance-chip" @click="router.push('/billing')">
              <el-icon><Wallet /></el-icon>
              <span>{{ (store.balance / 1e6).toFixed(4) }}</span>
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

      <main class="ow-content" :class="{ 'ow-content--flush': route.path === '/docs' }">
        <router-view />
      </main>

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
  Wallet, SwitchButton, ArrowDown
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const store = useUserStore()
const site = useSiteStore()

const logoImgErr = ref(false)
const isLoggedIn = computed(() => !!store.token)

const titles = {
  '/dashboard': '数据看板',
  '/playground': '文本对话',
  '/models':     '模型列表',
  '/keys':       'API 密钥',
  '/recharge':   '积分充值',
  '/billing':    '我的订单',
  '/docs':       '接口文档',
  '/tasks':      '调用日志',
  '/llm-logs':   '调用日志',
  '/stats':      '使用统计',
  '/image-gen':  '图片生成',
  '/video-gen':  '视频生成',
  '/exchange':   '兑换中心',
  '/profile':    '个人中心',
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
.ow-shell {
  display: flex;
  min-height: 100vh;
}

.ow-custom-header { width: 100%; order: -1; }

/* Sidebar */
.ow-sidebar {
  width: 210px;
  flex-shrink: 0;
  background: var(--ow-layout-sidebar, #ffffff);
  border-right: 1px solid var(--ow-layout-border, #e5e6eb);
  display: flex;
  flex-direction: column;
  position: fixed;
  top: 0; left: 0;
  height: 100vh;
  z-index: 100;
  overflow: hidden;
}

/* Logo */
.ow-logo {
  height: 54px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 16px;
  border-bottom: 1px solid var(--ow-layout-border, #e5e6eb);
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

/* Nav */
.ow-nav {
  flex: 1;
  padding: 4px 8px;
  overflow-y: auto;
  overflow-x: hidden;
  display: flex;
  flex-direction: column;
}
.ow-nav::-webkit-scrollbar { width: 0; }

.ow-nav-section {
  height: 40px;
  display: flex;
  align-items: center;
  padding: 10px 0 10px 8px;
  font-size: 13px;
  font-weight: 400;
  color: rgb(126, 131, 142);
  background: transparent;
  user-select: none;
}

.ow-nav-item {
  display: flex;
  align-items: center;
  height: 42px;
  margin: 0 0 4px 0;
  border-radius: 8px;
  color: rgb(24, 24, 24);
  text-decoration: none;
  font-size: 14px;
  font-weight: 400;
  transition: background .15s, color .15s;
  cursor: pointer;
}
.menu-item {
  display: flex;
  align-items: center;
  width: 100%;
  padding: 0 12px 0 20px;
}
.menu-item-title {
  display: block;
}
.ow-nav-item:hover {
  background: rgba(42, 85, 229, 0.06);
  color: rgb(22, 93, 255);
}
.ow-nav-item.active {
  background: rgba(42, 85, 229, 0.06);
  color: rgb(22, 93, 255);
  font-weight: 700;
}

/* Sidebar Footer */
.ow-sidebar-footer {
  padding: 10px 8px;
  border-top: 1px solid var(--ow-layout-border, #e5e6eb);
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
  border-radius: 8px;
  font-size: 13px;
  cursor: pointer;
  transition: background .15s, color .15s;
  color: rgb(24, 24, 24);
}
.balance-row:hover { background: rgba(42, 85, 229, 0.06); color: var(--ow-primary, #165dff); }
.logout-row:hover { background: rgba(42, 85, 229, 0.06); color: #f53f3f; }

.ow-footer-btn {
  display: block;
  text-align: center;
  padding: 6px 12px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 600;
  text-decoration: none;
  transition: all .15s;
}
.ow-footer-btn.primary {
  background: var(--ow-primary, #165dff);
  color: #fff;
}
.ow-footer-btn.primary:hover { opacity: .9; }
.ow-footer-btn.outline {
  border: 1px solid var(--ow-layout-border, #e5e6eb);
  color: rgb(24, 24, 24);
}
.ow-footer-btn.outline:hover {
  border-color: var(--ow-primary, #165dff);
  color: var(--ow-primary, #165dff);
}

/* Main area */
.ow-main {
  flex: 1;
  margin-left: 210px;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

/* Header */
.ow-header {
  height: 54px;
  background: var(--ow-layout-sidebar, #ffffff);
  border-bottom: 1px solid var(--ow-layout-border, #e5e6eb);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  flex-shrink: 0;
  position: sticky; top: 0; z-index: 50;
}
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
.header-contact-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 4px 12px;
  border-radius: 999px;
  font-size: 13px;
  font-weight: 500;
  text-decoration: none;
  cursor: pointer;
  transition: opacity .15s;
}
.header-contact-btn:hover { opacity: .85; }
.qq-btn { background: #ff6a00; color: #fff; }
.wechat-btn { background: #07c160; color: #fff; }
.balance-chip {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 4px 12px;
  border-radius: 999px;
  color: var(--ow-primary, #165dff);
  font-size: 13px;
  cursor: pointer;
  transition: background .15s;
  font-weight: 500;
  border: 1px solid var(--ow-layout-border, #e5e6eb);
}
.balance-chip:hover { background: rgba(42, 85, 229, 0.06); }
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
.avatar-btn:hover { background: rgba(42, 85, 229, 0.06); }
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

/* Content */
.ow-content {
  flex: 1;
  padding: 15px;
  overflow-y: auto;
  background: var(--ow-content-bg, #f2f3f5);
}
.ow-content--flush {
  padding: 0;
  overflow: hidden;
}

.ow-custom-footer { padding: 12px 24px; }

@media (max-width: 768px) {
  .ow-sidebar { transform: translateX(-100%); }
  .ow-main { margin-left: 0 !important; }
}
</style>
