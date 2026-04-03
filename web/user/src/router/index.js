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
      { path: '', redirect: '/dashboard' },
      { path: 'dashboard', name: 'Dashboard', component: () => import('@/views/dashboard/Index.vue') },
      { path: 'keys', name: 'APIKeys', component: () => import('@/views/keys/Index.vue') },
      { path: 'billing', name: 'Billing', component: () => import('@/views/billing/Index.vue') },
      { path: 'playground', name: 'Playground', component: () => import('@/views/playground/Index.vue') },
      { path: 'channels', name: 'Channels', component: () => import('@/views/dashboard/Channels.vue') },
      { path: 'docs', name: 'Docs', component: () => import('@/views/docs/Index.vue') },
      { path: 'tasks', name: 'Tasks', component: () => import('@/views/tasks/Index.vue') },
    ]
  },
  { path: '/:pathMatch(.*)*', redirect: '/' }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫：未登录跳 login，已登录不能访问 guest 路由
router.beforeEach((to) => {
  const token = localStorage.getItem('token')
  if (to.meta.requiresAuth && !token) return '/login'
  if (to.meta.guest && token) return '/'
})

export default router
