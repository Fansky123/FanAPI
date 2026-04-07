import http from './admin-http'

export const authApi = {
  login: (data) => http.post('/auth/login', data),
  changePassword: (data) => http.put('/user/password', data),
}

export const channelApi = {
  list: () => http.get('/admin/channels'),
  create: (data) => http.post('/admin/channels', data),
  update: (id, data) => http.put(`/admin/channels/${id}`, data),
  delete: (id) => http.delete(`/admin/channels/${id}`),
}

export const keyPoolApi = {
  listPools: (channelId = 0) => http.get('/admin/key-pools', { params: { channel_id: channelId } }),
  createPool: (data) => http.post('/admin/key-pools', data),
  deletePool: (id) => http.delete(`/admin/key-pools/${id}`),
  listKeys: (poolId) => http.get(`/admin/key-pools/${poolId}/keys`),
  addKey: (poolId, data) => http.post(`/admin/key-pools/${poolId}/keys`, data),
  removeKey: (id) => http.delete(`/admin/pool-keys/${id}`),
}

export const userApi = {
  list: (page = 1, size = 20) => http.get('/admin/users', { params: { page, size } }),
  recharge: (id, amount) => http.post(`/admin/users/${id}/recharge`, { amount }),
  resetPassword: (id, password) => http.put(`/admin/users/${id}/password`, { password }),
  setGroup: (id, group) => http.put(`/admin/users/${id}/group`, { group }),
}

export const txApi = {
  list: (params = {}) => http.get('/admin/transactions', { params }),
}

export const taskApi = {
  list: (params = {}) => http.get('/admin/tasks', { params }),
  get: (id) => http.get(`/admin/tasks/${id}`),
}

export const statsApi = {
  get: () => http.get('/admin/stats'),
}

export const cardApi = {
  generate: (data) => http.post('/admin/cards/generate', data),
  list: (params = {}) => http.get('/admin/cards', { params }),
  remove: (id) => http.delete(`/admin/cards/${id}`),
}

export const llmLogApi = {
  list: (params = {}) => http.get('/admin/llm-logs', { params }),
  get: (id) => http.get(`/admin/llm-logs/${id}`),
}

export const settingsApi = {
  get: () => http.get('/admin/settings'),
  update: (data) => http.put('/admin/settings', data),
}

