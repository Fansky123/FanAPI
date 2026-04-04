import http from './http'

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
  // 号池 CRUD（channel_id=0 表示查询全部）
  listPools: (channelId = 0) => http.get('/admin/key-pools', { params: { channel_id: channelId } }),
  createPool: (data) => http.post('/admin/key-pools', data),
  deletePool: (id) => http.delete(`/admin/key-pools/${id}`),
  // 号池内 Key CRUD
  listKeys: (poolId) => http.get(`/admin/key-pools/${poolId}/keys`),
  addKey: (poolId, data) => http.post(`/admin/key-pools/${poolId}/keys`, data),
  removeKey: (id) => http.delete(`/admin/pool-keys/${id}`),
}

export const userApi = {
  list: (page = 1, size = 20) => http.get('/admin/users', { params: { page, size } }),
  recharge: (id, amount) => http.post(`/admin/users/${id}/recharge`, { amount }),
  resetPassword: (id, password) => http.put(`/admin/users/${id}/password`, { password }),
}

export const txApi = {
  list: (params = {}) => http.get('/admin/transactions', { params }),
}

export const taskApi = {
  list: (params = {}) => http.get('/admin/tasks', { params }),
  get: (id) => http.get(`/admin/tasks/${id}`),
}
