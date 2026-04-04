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
    path: '/',
    component: () => import('@/views/dashboard/Layout.vue'),
    meta: { requiresAuth: true },
    children: [
      { path: '', redirect: '/playground' },
      { path: 'playground', name: 'Playground', component: () => import('@/views/playground/Index.vue') },
      { path: 'models', name: 'Models', component: () => import('@/views/dashboard/Channels.vue') },
      { path: 'keys', name: 'APIKeys', component: () => import('@/views/keys/Index.vue') },
      { path: 'docs', name: 'Docs', component: () => import('@/views/docs/Index.vue') },
      { path: 'tasks', name: 'Tasks', component: () => import('@/views/tasks/Index.vue') },
      { path: 'billing', name: 'Billing', component: () => import('@/views/billing/Index.vue') },
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
  if (to.meta.requiresAuth && !token) return '/login'
  if (to.meta.guest && token) return '/'
  if (to.meta.requiresAdmin && !adminToken) return '/admin/login'
  if (to.meta.adminGuest && adminToken) return '/admin/dashboard'
})

export default router
