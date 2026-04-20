import { createHttpClient } from '@/lib/api/http'

const http = createHttpClient('admin')

export type AdminLoginResponse = {
  token: string
}

export type AdminStatsResponse = {
  total_users?: number
  users?: number
  total_requests?: number
  requests?: number
  total_revenue?: number
  revenue?: number
}

export type AdminChannel = {
  id?: number
  name?: string
  model?: string
  routing_model?: string
  type?: string
  protocol?: string
  base_url?: string
  method?: string
  timeout_ms?: number
  billing_type?: string
  headers?: Record<string, unknown>
  billing_config?: Record<string, unknown>
  is_active?: boolean
}

export type AdminUser = {
  id?: number
  username?: string
  email?: string
  group?: string
  balance_credits?: number
  balance?: number
}

export type AdminTransaction = {
  id?: number
  created_at?: string
  type?: string
  amount?: number
  credits?: number
  remark?: string
  description?: string
}

export type AdminTask = {
  id?: number
  type?: string
  status?: string
  created_at?: string
  upstream_task_id?: string
}

export type AdminLog = {
  id?: number
  model?: string
  created_at?: string
  corr_id?: string
  cost_credits?: number
  status?: string
}

export type AdminVendor = {
  id?: number
  name?: string
  username?: string
  email?: string
  is_active?: boolean
  enabled?: boolean
  commission_ratio?: number
  fee_ratio?: number
}

export type AdminCard = {
  id?: number
  code?: string
  credits?: number
  status?: string
  note?: string
  used_at?: string
  created_at?: string
}

export type AdminWithdrawal = {
  id?: number
  username?: string
  created_at?: string
  amount?: number
  payment_type?: string
  payment_qr?: string
  status?: string
  admin_remark?: string
}

export const adminAuthApi = {
  login: (payload: { username: string; password: string }) =>
    http.post<AdminLoginResponse>('/auth/login', payload),
}

export const adminApi = {
  getStats: () => http.get<AdminStatsResponse>('/admin/stats'),
  listChannels: () =>
    http.get<{ channels?: AdminChannel[]; items?: AdminChannel[] } | AdminChannel[]>(
      '/admin/channels'
    ),
  createChannel: (payload: Partial<AdminChannel>) =>
    http.post<AdminChannel>('/admin/channels', payload),
  updateChannel: (id: number, payload: Partial<AdminChannel>) =>
    http.put<AdminChannel>(`/admin/channels/${id}`, payload),
  toggleChannel: (id: number, isActive: boolean) =>
    http.patch<Record<string, unknown>>(`/admin/channels/${id}/active`, {
      is_active: isActive,
    }),
  deleteChannel: (id: number) =>
    http.delete<Record<string, unknown>>(`/admin/channels/${id}`),
  listUsers: (page = 1, size = 20) =>
    http.get<{ items?: AdminUser[]; users?: AdminUser[] } | AdminUser[]>(
      '/admin/users',
      { params: { page, size } }
    ),
  rechargeUser: (id: number, amount: number) =>
    http.post<Record<string, unknown>>(`/admin/users/${id}/recharge`, { amount }),
  resetUserPassword: (id: number, password: string) =>
    http.put<Record<string, unknown>>(`/admin/users/${id}/password`, { password }),
  setUserGroup: (id: number, group: string) =>
    http.put<Record<string, unknown>>(`/admin/users/${id}/group`, { group }),
  setUserRole: (id: number, role: string) =>
    http.put<Record<string, unknown>>(`/admin/users/${id}/role`, { role }),
  listTransactions: (params: Record<string, unknown> = {}) =>
    http.get<{ items?: AdminTransaction[]; transactions?: AdminTransaction[] } | AdminTransaction[]>(
      '/admin/transactions',
      { params }
    ),
  listTasks: (params: Record<string, unknown> = {}) =>
    http.get<{ items?: AdminTask[]; tasks?: AdminTask[] } | AdminTask[]>(
      '/admin/tasks',
      { params }
    ),
  listLogs: (params: Record<string, unknown> = {}) =>
    http.get<{ items?: AdminLog[]; logs?: AdminLog[] } | AdminLog[]>(
      '/admin/llm-logs',
      { params }
    ),
  getSettings: () =>
    http.get<{ settings?: Record<string, string> } | Record<string, string>>(
      '/admin/settings'
    ),
  updateSettings: (payload: Record<string, string>) =>
    http.put<Record<string, unknown>>('/admin/settings', payload),
  listVendors: (params: Record<string, unknown> = {}) =>
    http.get<{ items?: AdminVendor[]; vendors?: AdminVendor[] } | AdminVendor[]>(
      '/admin/vendors',
      { params }
    ),
  updateVendor: (id: number, payload: { is_active?: boolean; commission_ratio?: number }) =>
    http.patch<Record<string, unknown>>(`/admin/vendors/${id}`, payload),
  generateCards: (payload: { count: number; credits: number; note: string }) =>
    http.post<{ cards?: AdminCard[] }>('/admin/cards/generate', payload),
  listCards: (params: Record<string, unknown> = {}) =>
    http.get<{ cards?: AdminCard[]; total?: number }>('/admin/cards', { params }),
  deleteCard: (id: number) =>
    http.delete<Record<string, unknown>>(`/admin/cards/${id}`),
  listWithdrawals: (params: Record<string, unknown> = {}) =>
    http.get<{ records?: AdminWithdrawal[]; total?: number }>('/admin/withdrawals', {
      params,
    }),
  getPendingWithdrawCount: () =>
    http.get<{ count?: number }>('/admin/withdrawals/pending-count'),
  approveWithdrawal: (id: number, remark = '') =>
    http.post<Record<string, unknown>>(`/admin/withdrawals/${id}/approve`, { remark }),
  rejectWithdrawal: (id: number, remark = '') =>
    http.post<Record<string, unknown>>(`/admin/withdrawals/${id}/reject`, { remark }),
}
