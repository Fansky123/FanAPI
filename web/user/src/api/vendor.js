import axios from 'axios'
import { ElMessage } from 'element-plus'

const vendorHttp = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 30000,
})

vendorHttp.interceptors.request.use((config) => {
  const token = localStorage.getItem('vendor_token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

vendorHttp.interceptors.response.use(
  (res) => res.data,
  (err) => {
    const msg = err.response?.data?.error || err.response?.data?.message || '请求失败'
    if (err.response?.status === 401) {
      localStorage.removeItem('vendor_token')
      window.location.href = '/vendor/login'
    } else {
      ElMessage.error(msg)
    }
    return Promise.reject(err)
  }
)

export const vendorAuthApi = {
  login: (data) => vendorHttp.post('/vendor/auth/login', data),
  register: (data) => vendorHttp.post('/vendor/auth/register', data),
}

export const vendorApi = {
  getProfile: () => vendorHttp.get('/vendor/profile'),
  getKeys: () => vendorHttp.get('/vendor/keys'),
  getPools: () => vendorHttp.get('/vendor/pools'),
  submitKey: (data) => vendorHttp.post('/vendor/keys', data),
}
