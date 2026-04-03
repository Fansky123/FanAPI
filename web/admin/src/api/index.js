import http from './http'

export const authApi = {
  login: (data) => http.post('/auth/login', data),
}

export const channelApi = {
  list: () => http.get('/admin/channels'),
  create: (data) => http.post('/admin/channels', data),
  update: (id, data) => http.put(`/admin/channels/${id}`, data),
  delete: (id) => http.delete(`/admin/channels/${id}`),
}

export const userApi = {
  list: (page = 1, size = 20) => http.get('/admin/users', { params: { page, size } }),
  recharge: (id, amount) => http.post(`/admin/users/${id}/recharge`, { amount }),
}

export const txApi = {
  list: (params = {}) => http.get('/admin/transactions', { params }),
}

export const taskApi = {
  list: (params = {}) => http.get('/admin/tasks', { params }),
  get: (id) => http.get(`/admin/tasks/${id}`),
}
