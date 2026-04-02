import http from './http'

// 认证相关
export const authApi = {
  sendCode: (email) => http.post('/auth/send-code', { email }),
  register: (data) => http.post('/auth/register', data),
  login: (data) => http.post('/auth/login', data),
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
}
