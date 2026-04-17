import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

const http = axios.create({ baseURL: '/api', timeout: 30000 })

http.interceptors.request.use(cfg => {
  const token = localStorage.getItem('agent_token')
  if (token) cfg.headers.Authorization = `Bearer ${token}`
  return cfg
})

http.interceptors.response.use(
  res => res.data,
  err => {
    const msg = err.response?.data?.error || err.message || '请求失败'
    if (err.response?.status === 401) {
      localStorage.removeItem('agent_token')
      router.push('/agent/login')
    } else {
      ElMessage.error(msg)
    }
    return Promise.reject(err)
  }
)

export default http
