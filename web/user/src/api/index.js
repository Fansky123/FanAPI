import http from './http'

// 认证相关
export const authApi = {
  sendCode: (email) => http.post('/auth/send-code', { email }),
  register: (data) => http.post('/auth/register', data),
  login: (data) => http.post('/auth/login', data),
}

// 任务相关（需 API Key）
export const taskApi = {
  list: (params) => http.get('/v1/tasks', { params }),
  get: (id) => http.get(`/v1/tasks/${id}`),
}

// 用户相关（需 JWT）
export const userApi = {
  getBalance: () => http.get('/user/balance'),
  getTransactions: (page = 1, size = 20) =>
    http.get('/user/transactions', { params: { page, size } }),
  listAPIKeys: () => http.get('/user/apikeys'),
  createAPIKey: (name) => http.post('/user/apikeys', { name }),
  deleteAPIKey: (id) => http.delete(`/user/apikeys/${id}`),
  listChannels: () => http.get('/user/channels'),
  redeemCard: (code) => http.post('/user/cards/redeem', { code }),
}
