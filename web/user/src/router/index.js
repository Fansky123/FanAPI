import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/auth/Login.vue'),
    meta: { guest: true }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/auth/Register.vue'),
    meta: { guest: true }
  },
  {
    path: '/forgot-password',
    name: 'ForgotPassword',
    component: () => import('@/views/auth/ForgotPassword.vue'),
  },
  {
    path: '/',
    component: () => import('@/views/dashboard/Layout.vue'),
    children: [
      { path: '', redirect: '/dashboard' },
      { path: 'dashboard', name: 'Dashboard', component: () => import('@/views/dashboard/Index.vue') },
      { path: 'playground', name: 'Playground', component: () => import('@/views/playground/Index.vue'), meta: { requiresAuth: true } },
      { path: 'models', name: 'Models', component: () => import('@/views/dashboard/Channels.vue') },
      { path: 'keys', name: 'APIKeys', component: () => import('@/views/keys/Index.vue'), meta: { requiresAuth: true } },
      { path: 'docs', name: 'Docs', component: () => import('@/views/docs/Index.vue') },
      { path: 'tasks', name: 'Tasks', component: () => import('@/views/tasks/Index.vue'), meta: { requiresAuth: true } },
      { path: 'llm-logs', name: 'LLMLogs', component: () => import('@/views/llm-logs/Index.vue'), meta: { requiresAuth: true } },
      { path: 'recharge', name: 'Recharge', component: () => import('@/views/billing/Index.vue'), meta: { requiresAuth: true } },
      { path: 'billing', name: 'Billing', component: () => import('@/views/billing/Index.vue'), meta: { requiresAuth: true } },
      { path: 'stats', name: 'Stats', component: () => import('@/views/stats/Index.vue'), meta: { requiresAuth: true } },
      { path: 'image-gen', name: 'ImageGen', component: () => import('@/views/image-gen/Index.vue'), meta: { requiresAuth: true } },
      { path: 'video-gen', name: 'VideoGen', component: () => import('@/views/video-gen/Index.vue'), meta: { requiresAuth: true } },
      { path: 'exchange', name: 'Exchange', component: () => import('@/views/exchange/Index.vue'), meta: { requiresAuth: true } },
      { path: 'profile', name: 'Profile', component: () => import('@/views/profile/Index.vue'), meta: { requiresAuth: true } },
      { path: 'invite', name: 'Invite', component: () => import('@/views/invite/Index.vue'), meta: { requiresAuth: true } },
    ]
  },
  // 管理端路由
  {
    path: '/admin/login',
    component: () => import('@/views/admin/Login.vue'),
    meta: { adminGuest: true }
  },
  {
    path: '/admin',
    component: () => import('@/views/admin/Layout.vue'),
    meta: { requiresAdmin: true },
    children: [
      { path: '', redirect: '/admin/dashboard' },
      { path: 'dashboard', component: () => import('@/views/admin/dashboard/Index.vue') },
      { path: 'channels', component: () => import('@/views/admin/channels/Index.vue') },
      { path: 'key-pools', component: () => import('@/views/admin/keypools/Index.vue') },
      { path: 'users', component: () => import('@/views/admin/users/Index.vue') },
      { path: 'billing', component: () => import('@/views/admin/billing/Index.vue') },
      { path: 'tasks', component: () => import('@/views/admin/tasks/Index.vue') },
      { path: 'cards', component: () => import('@/views/admin/cards/Index.vue') },
      { path: 'ocpc', component: () => import('@/views/admin/ocpc/Index.vue') },
      { path: 'llm-logs', component: () => import('@/views/admin/llm-logs/Index.vue') },
      { path: 'settings', component: () => import('@/views/admin/settings/Index.vue') },
      { path: 'vendors', component: () => import('@/views/admin/vendors/Index.vue') },
      { path: 'withdraw', component: () => import('@/views/admin/withdraw/Index.vue') },
    ]
  },
  // 客服端路由
  {
    path: '/agent/login',
    component: () => import('@/views/agent/Login.vue'),
    meta: { agentGuest: true }
  },
  {
    path: '/agent',
    component: () => import('@/views/agent/Layout.vue'),
    meta: { requiresAgent: true },
    children: [
      { path: '', redirect: '/agent/dashboard' },
      { path: 'dashboard', component: () => import('@/views/agent/Dashboard.vue') },
    ]
  },
  // 号商端路由
  {
    path: '/vendor/login',
    component: () => import('@/views/vendor/Login.vue'),
    meta: { vendorGuest: true }
  },
  {
    path: '/vendor',
    component: () => import('@/views/vendor/Layout.vue'),
    meta: { requiresVendor: true },
    children: [
      { path: '', redirect: '/vendor/dashboard' },
      { path: 'dashboard', component: () => import('@/views/vendor/Dashboard.vue') },
      { path: 'keys', component: () => import('@/views/vendor/Keys.vue') },
    ]
  },
  { path: '/:pathMatch(.*)*', redirect: '/' }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to) => {
  const token = localStorage.getItem('token')
  const adminToken = localStorage.getItem('admin_token')
  const agentToken = localStorage.getItem('agent_token')
  const vendorToken = localStorage.getItem('vendor_token')
  if (to.meta.requiresAuth && !token) return '/login'
  if (to.meta.guest && token) return '/dashboard'
  if (to.meta.requiresAdmin && !adminToken) return '/admin/login'
  if (to.meta.adminGuest && adminToken) return '/admin/dashboard'
  if (to.meta.requiresAgent && !agentToken) return '/agent/login'
  if (to.meta.agentGuest && agentToken) return '/agent/dashboard'
  if (to.meta.requiresVendor && !vendorToken) return '/vendor/login'
  if (to.meta.vendorGuest && vendorToken) return '/vendor/dashboard'
})

export default router
