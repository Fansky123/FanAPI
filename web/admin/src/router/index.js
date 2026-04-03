import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { guest: true }
  },
  {
    path: '/',
    component: () => import('@/views/Layout.vue'),
    meta: { requiresAuth: true },
    children: [
      { path: '', redirect: '/channels' },
      { path: 'channels', name: 'Channels', component: () => import('@/views/channels/Index.vue') },
      { path: 'users', name: 'Users', component: () => import('@/views/users/Index.vue') },
      { path: 'billing', name: 'Billing', component: () => import('@/views/billing/Index.vue') },
      { path: 'tasks', name: 'Tasks', component: () => import('@/views/tasks/Index.vue') },
    ]
  },
  { path: '/:pathMatch(.*)*', redirect: '/' }
]

const router = createRouter({ history: createWebHistory(), routes })

router.beforeEach((to) => {
  const token = localStorage.getItem('admin_token')
  if (to.meta.requiresAuth && !token) return '/login'
  if (to.meta.guest && token) return '/'
})

export default router
