import { createHttpClient } from '@/lib/api/http'

const http = createHttpClient('admin')

export type AdminLoginResponse = {
  token: string
}

export type AdminStatsResponse = {
  // flat fields (fallback)
  total_users?: number
  users?: number
  total_requests?: number
  requests?: number
  total_revenue?: number
  revenue?: number
  // structured fields (preferred)
  channels?: number
  active_channels?: number
  today?: {
    revenue?: number
    cost?: number
    profit?: number
    count?: number
  }
  total?: {
    revenue?: number
    cost?: number
    profit?: number
  }
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
  query_url?: string
  query_method?: string
  timeout_ms?: number
  query_timeout_ms?: number
  billing_type?: string
  headers?: Record<string, unknown>
  billing_config?: Record<string, unknown>
  billing_script?: string
  request_script?: string
  response_script?: string
  query_script?: string
  error_script?: string
  key_pool_id?: number
  auth_type?: string
  auth_param_name?: string
  auth_region?: string
  auth_service?: string
  passthrough_headers?: boolean
  passthrough_body?: boolean
  weight?: number
  priority?: number
  icon_url?: string
  description?: string
  is_active?: boolean
}

export type AdminUser = {
  id?: number
  username?: string
  email?: string
  role?: string
  group?: string
  balance_credits?: number
  balance?: number
  is_active?: boolean
  rebate_ratio?: number | null
  created_at?: string
}

export type AdminTransaction = {
  id?: number
  user_id?: number
  created_at?: string
  type?: string
  amount?: number
  credits?: number
  cost?: number
  channel_id?: number
  corr_id?: string
  remark?: string
  description?: string
}

export type AdminTransactionSummary = {
  revenue?: number
  cost?: number
  profit?: number
  transaction_count?: number
}

export type AdminTask = {
  id?: number
  user_id?: number
  channel_id?: number
  type?: string
  status?: string
  error_msg?: string
  credits_charged?: number
  upstream_task_id?: string
  corr_id?: string
  request?: Record<string, unknown>
  upstream_request?: Record<string, unknown>
  upstream_response?: Record<string, unknown>
  result?: Record<string, unknown>
  created_at?: string
  updated_at?: string
}

export type AdminLog = {
  id?: number
  user_id?: number
  channel_id?: number
  api_key_id?: number
  model?: string
  created_at?: string
  updated_at?: string
  corr_id?: string
  credits_charged?: number
  cost_charged?: number
  status?: string
  is_stream?: boolean
  error_msg?: string
  upstream_status?: number
  upstream_method?: string
  upstream_url?: string
  usage?: {
    prompt_tokens?: number
    completion_tokens?: number
    total_tokens?: number
    cache_creation_tokens?: number
    cache_read_tokens?: number
    estimated?: boolean
  }
  client_request?: Record<string, unknown>
  client_response?: Record<string, unknown>
  upstream_headers?: Record<string, unknown>
  upstream_request?: Record<string, unknown>
  upstream_response?: Record<string, unknown>
}

export type AdminVendor = {
  id?: number
  name?: string
  username?: string
  email?: string
  invite_code?: string
  is_active?: boolean
  enabled?: boolean
  commission_ratio?: number
  fee_ratio?: number
  balance?: number
  balance_credits?: number
  created_at?: string
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

export type AdminKeyPool = {
  id?: number
  name?: string
  channel_id?: number
  is_active?: boolean
  vendor_submittable?: boolean
}

export type AdminPoolKey = {
  id?: number
  pool_id?: number
  vendor_id?: number | null
  value?: string
  priority?: number
  is_active?: boolean
}

export type AdminOcpcPlatform = {
  id?: number
  platform?: string
  name?: string
  enabled?: boolean
  baidu_page_url?: string
  baidu_token?: string
  baidu_reg_type?: number
  baidu_order_type?: number
  e360_key?: string
  e360_secret?: string
  e360_jzqs?: string
  e360_so_type?: string
  e360_reg_event?: string
  e360_order_event?: string
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
    http.get<{ items?: AdminUser[]; users?: AdminUser[]; total?: number } | AdminUser[]>(
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
    http.get<{ items?: AdminTransaction[]; transactions?: AdminTransaction[]; total?: number; summary?: AdminTransactionSummary } | AdminTransaction[]>(
      '/admin/transactions',
      { params }
    ),
  listTasks: (params: Record<string, unknown> = {}) =>
    http.get<{ items?: AdminTask[]; tasks?: AdminTask[]; total?: number } | AdminTask[]>(
      '/admin/tasks',
      { params }
    ),
  getAdminTask: (id: number) =>
    http.get<{ task?: AdminTask } | AdminTask>(`/admin/tasks/${id}`),
  listLogs: (params: Record<string, unknown> = {}) =>
    http.get<{ logs?: AdminLog[]; items?: AdminLog[]; total?: number }>('/admin/llm-logs', { params }),
  getLog: (id: number) =>
    http.get<AdminLog>(`/admin/llm-logs/${id}`),
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
  listKeyPools: (channelId?: number) =>
    http.get<{ pools?: AdminKeyPool[] } | AdminKeyPool[]>('/admin/key-pools', {
      params: channelId ? { channel_id: channelId } : undefined,
    }),
  createKeyPool: (payload: { channel_id: number; name: string }) =>
    http.post<Record<string, unknown>>('/admin/key-pools', payload),
  deleteKeyPool: (id: number) =>
    http.delete<Record<string, unknown>>(`/admin/key-pools/${id}`),
  toggleKeyPool: (id: number) =>
    http.patch<Record<string, unknown>>(`/admin/key-pools/${id}/toggle`, {}),
  toggleVendorSubmittable: (id: number) =>
    http.patch<Record<string, unknown>>(`/admin/key-pools/${id}/vendor-toggle`, {}),
  listPoolKeys: (poolId: number) =>
    http.get<{ keys?: AdminPoolKey[] } | AdminPoolKey[]>(`/admin/key-pools/${poolId}/keys`),
  addPoolKey: (poolId: number, payload: { value: string; priority: number }) =>
    http.post<Record<string, unknown>>(`/admin/key-pools/${poolId}/keys`, payload),
  removePoolKey: (id: number) =>
    http.delete<Record<string, unknown>>(`/admin/pool-keys/${id}`),
  updatePoolKey: (id: number, payload: { priority: number; is_active: boolean }) =>
    http.patch<Record<string, unknown>>(`/admin/pool-keys/${id}`, payload),
  setPoolKeyVendor: (id: number, vendorId: number | null) =>
    http.patch<Record<string, unknown>>(`/admin/pool-keys/${id}/vendor`, { vendor_id: vendorId }),
  setUserRebateRatio: (id: number, ratio: number | null) =>
    http.put<Record<string, unknown>>(`/admin/users/${id}/rebate-ratio`, { rebate_ratio: ratio }),
  triggerOcpcUpload: () =>
    http.post<Record<string, unknown>>('/admin/ocpc/upload', {}),
  getOcpcSchedule: () =>
    http.get<{ schedule?: Record<string, string> }>('/admin/ocpc/schedule'),
  updateOcpcSchedule: (payload: { enabled: boolean; interval: number }) =>
    http.put<Record<string, unknown>>('/admin/ocpc/schedule', payload),
  listOcpcPlatforms: () =>
    http.get<{ list?: AdminOcpcPlatform[] } | AdminOcpcPlatform[]>('/admin/ocpc/platforms'),
  createOcpcPlatform: (payload: Record<string, unknown>) =>
    http.post<Record<string, unknown>>('/admin/ocpc/platforms', payload),
  updateOcpcPlatform: (id: number, payload: Record<string, unknown>) =>
    http.put<Record<string, unknown>>(`/admin/ocpc/platforms/${id}`, payload),
  deleteOcpcPlatform: (id: number) =>
    http.delete<Record<string, unknown>>(`/admin/ocpc/platforms/${id}`),
  toggleOcpcPlatform: (id: number) =>
    http.patch<Record<string, unknown>>(`/admin/ocpc/platforms/${id}/toggle`, {}),
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
