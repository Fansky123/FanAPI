import { createHttpClient } from '@/lib/api/http'

const http = createHttpClient('user')

export type UserProfileResponse = {
  username?: string
  email?: string
  group?: string
}

export type UserBalanceResponse = {
  balance_credits?: number
}

export type UserStatsResponse = {
  total_consumed?: number
  today_consumed?: number
}

export type UserTransaction = {
  id?: number
  created_at?: string
  time?: string
  type?: string
  amount?: number
  credits?: number
  remark?: string
  description?: string
}

export type ApiKeyRecord = {
  id?: number
  name?: string
  key?: string
  masked_key?: string
  key_type?: string
}

export type UserChannel = {
  id?: number
  name?: string
  routing_model?: string
  model?: string
  description?: string
  type?: string
  category?: string
}

export type UserTask = {
  id?: number
  type?: string
  status?: string
  created_at?: string
  upstream_task_id?: string
}

export type UserLog = {
  id?: number
  model?: string
  created_at?: string
  corr_id?: string
  cost_credits?: number
  status?: string
}

export const userApi = {
  getProfile: () => http.get<UserProfileResponse>('/user/profile'),
  getBalance: () => http.get<UserBalanceResponse>('/user/balance'),
  getStats: () => http.get<UserStatsResponse>('/user/stats'),
  getTransactions: (page = 1, size = 20) =>
    http.get<{ items?: UserTransaction[]; transactions?: UserTransaction[] } | UserTransaction[]>(
      '/user/transactions',
      { params: { page, size } }
    ),
  listApiKeys: () =>
    http.get<{ api_keys?: ApiKeyRecord[]; keys?: ApiKeyRecord[] } | ApiKeyRecord[]>(
      '/user/apikeys'
    ),
  createApiKey: (name: string, keyType = 'low_price') =>
    http.post<Record<string, unknown>>('/user/apikeys', { name, key_type: keyType }),
  deleteApiKey: (id: number) =>
    http.delete<Record<string, unknown>>(`/user/apikeys/${id}`),
  listChannels: () =>
    http.get<{ channels?: UserChannel[] } | UserChannel[]>('/user/channels'),
  listTasks: (params: Record<string, unknown> = {}) =>
    http.get<{ items?: UserTask[]; tasks?: UserTask[] } | UserTask[]>('/v1/tasks', {
      params,
    }),
  listLogs: (params: Record<string, unknown> = {}) =>
    http.get<{ items?: UserLog[]; logs?: UserLog[] } | UserLog[]>('/v1/llm-logs', {
      params,
    }),
}
