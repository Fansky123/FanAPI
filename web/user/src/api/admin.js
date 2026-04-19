import http from './admin-http'

export const authApi = {
  login: (data) => http.post('/auth/login', data),
  changePassword: (data) => http.put('/user/password', data),
}

export const channelApi = {
  list: () => http.get('/admin/channels'),
  create: (data) => http.post('/admin/channels', data),
  update: (id, data) => http.put(`/admin/channels/${id}`, data),
  patchActive: (id, isActive) => http.patch(`/admin/channels/${id}/active`, { is_active: isActive }),
  delete: (id) => http.delete(`/admin/channels/${id}`),
}

export const keyPoolApi = {
  listPools: (channelId) => http.get('/admin/key-pools', channelId ? { params: { channel_id: channelId } } : undefined),
  createPool: (data) => http.post('/admin/key-pools', data),
  deletePool: (id) => http.delete(`/admin/key-pools/${id}`),
  togglePool: (id) => http.patch(`/admin/key-pools/${id}/toggle`),
  toggleVendorSubmittable: (id) => http.patch(`/admin/key-pools/${id}/vendor-toggle`),
  listKeys: (poolId) => http.get(`/admin/key-pools/${poolId}/keys`),
  addKey: (poolId, data) => http.post(`/admin/key-pools/${poolId}/keys`, data),
  removeKey: (id) => http.delete(`/admin/pool-keys/${id}`),
}

export const userApi = {
  list: (page = 1, size = 20) => http.get('/admin/users', { params: { page, size } }),
  recharge: (id, amount) => http.post(`/admin/users/${id}/recharge`, { amount }),
  resetPassword: (id, password) => http.put(`/admin/users/${id}/password`, { password }),
  setGroup: (id, group) => http.put(`/admin/users/${id}/group`, { group }),
  setRole: (id, role) => http.put(`/admin/users/${id}/role`, { role }),
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

export const withdrawApi = {
  list: (params = {}) => http.get('/admin/withdrawals', { params }),
  pendingCount: () => http.get('/admin/withdrawals/pending-count'),
  approve: (id, remark = '') => http.post(`/admin/withdrawals/${id}/approve`, { remark }),
  reject: (id, remark = '') => http.post(`/admin/withdrawals/${id}/reject`, { remark }),
}

export const llmLogApi = {
  list: (params = {}) => http.get('/admin/llm-logs', { params }),
  get: (id) => http.get(`/admin/llm-logs/${id}`),
}

export const settingsApi = {
  get: () => http.get('/admin/settings'),
  update: (data) => http.put('/admin/settings', data),
}

export const vendorAdminApi = {
  list: (params = {}) => http.get('/admin/vendors', { params }),
  update: (id, data) => http.patch(`/admin/vendors/${id}`, data),
  setPoolKeyVendor: (poolKeyId, data) => http.patch(`/admin/pool-keys/${poolKeyId}/vendor`, data),
  setUserRebateRatio: (userId, data) => http.put(`/admin/users/${userId}/rebate-ratio`, data),
}

export const ocpcApi = {
  upload: () => http.post('/admin/ocpc/upload'),
  getSchedule: () => http.get('/admin/ocpc/schedule'),
  updateSchedule: (data) => http.put('/admin/ocpc/schedule', data),
}

export const ocpcPlatformApi = {
  list: () => http.get('/admin/ocpc/platforms'),
  create: (data) => http.post('/admin/ocpc/platforms', data),
  update: (id, data) => http.put(`/admin/ocpc/platforms/${id}`, data),
  delete: (id) => http.delete(`/admin/ocpc/platforms/${id}`),
  toggle: (id) => http.patch(`/admin/ocpc/platforms/${id}/toggle`),
}

