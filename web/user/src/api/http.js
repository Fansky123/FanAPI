import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

const http = axios.create({ baseURL: '/api', timeout: 30000 })

// 请求拦截：自动附带 token（用户端用 JWT）
http.interceptors.request.use(cfg => {
  const token = localStorage.getItem('token')
  if (token) cfg.headers.Authorization = `Bearer ${token}`
  return cfg
})

// 响应拦截：统一错误提示，401 跳登录（仅当用户已登录时才跳转）
http.interceptors.response.use(
  res => res.data,
  err => {
    const msg = err.response?.data?.error || err.message || '请求失败'
    if (err.response?.status === 401) {
      localStorage.removeItem('token')
      // 只有先前有 token 的情况（即 token 失效）才跳登录页，避免未登录用户被强制跳转
      const wasLoggedIn = !!err.config?.headers?.Authorization
      if (wasLoggedIn) router.push('/login')
    } else {
      ElMessage.error(msg)
    }
    return Promise.reject(err)
  }
)

export default http
