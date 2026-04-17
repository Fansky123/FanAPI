import http from './agent-http'

export const agentAuthApi = {
  login: (data) => http.post('/auth/login', data),
}

export const agentUserApi = {
  list: (page = 1, size = 50) => http.get('/agent/users', { params: { page, size } }),
  recharge: (id, amount) => http.post(`/agent/users/${id}/recharge`, { amount }),
}

export const agentInviteApi = {
  get: () => http.get('/agent/invite'),
  updateWechatQR: (wechat_qr) => http.put('/agent/wechat-qr', { wechat_qr }),
}
