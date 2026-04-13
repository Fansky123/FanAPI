import http from './http'

// 公开接口（无需登录）
export const publicApi = {
  listChannels: () => http.get('/public/channels'),
  getSettings: () => http.get('/public/settings'),
}

// 认证相关
export const authApi = {
  sendCode: (email) => http.post('/auth/send-code', { email }),
  register: (data) => http.post('/auth/register', data),
  login: (data) => http.post('/auth/login', data),
  forgotPassword: (email) => http.post('/auth/forgot-password', { email }),
  resetPassword: (data) => http.post('/auth/reset-password', data),
}

// 任务相关（需 API Key）
export const taskApi = {
  list: (params) => http.get('/v1/tasks', { params }),
  get: (id) => http.get(`/v1/tasks/${id}`),
}

// 用户相关（需 JWT）
export const userApi = {
  getProfile: () => http.get('/user/profile'),
  getBalance: () => http.get('/user/balance'),
  getStats: () => http.get('/user/stats'),
  getTransactions: (page = 1, size = 20) =>
    http.get('/user/transactions', { params: { page, size } }),
  listAPIKeys: () => http.get('/user/apikeys'),
  createAPIKey: (name) => http.post('/user/apikeys', { name }),
  deleteAPIKey: (id) => http.delete(`/user/apikeys/${id}`),
  listChannels: () => http.get('/user/channels'),
  redeemCard: (code) => http.post('/user/cards/redeem', { code }),
  bindEmail: (data) => http.post('/user/bind-email', data),
}

// 支付相关（需 JWT）
export const payApi = {
  createEpayOrder: (data) => http.post('/pay/epay/create', data),
  createPayApplyOrder: (data) => http.post('/pay/apply/create', data),
  getOrderStatus: (outTradeNo) => http.get('/pay/order/status', { params: { out_trade_no: outTradeNo } }),
  listOrders: (page = 1, size = 20) => http.get('/user/payment-orders', { params: { page, size } }),
}

// LLM 日志（需登录）
export const llmLogApi = {
  list: (params = {}) => http.get('/v1/llm-logs', { params }),
  get: (id) => http.get(`/v1/llm-logs/${id}`),
}

