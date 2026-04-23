import { createHttpClient } from '@/lib/api/http'

const http = createHttpClient('user')

export type UserProfileResponse = {
  id?: number
  username?: string
  email?: string
  group?: string
  balance?: number
}

export type UserBalanceResponse = {
  balance_credits?: number
}

export type UserStatsResponse = {
  total_consumed?: number
  today_consumed?: number
  daily_credits?: Array<{ day?: string; credits?: number }>
  daily_requests?: Array<{ day?: string; success?: number; failed?: number }>
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
  raw_key?: string
  key_prefix?: string
  viewable?: boolean
  masked_key?: string
  key_type?: string
  is_active?: boolean
  last_used_at?: string | null
  created_at?: string
}

export type UserChannel = {
  id?: number
  name?: string
  routing_model?: string
  model?: string
  description?: string
  type?: string
  category?: string
  protocol?: string
  billing_type?: string
  icon_url?: string
  price_display?: string
  group_price?: string
}

export type UserTask = {
  id?: number
  task_id?: number
  task_type?: string
  type?: string
  status?: number | string
  msg?: string
  error_msg?: string
  credits_charged?: number
  upstream_task_id?: string
  request?: Record<string, unknown>
  result?: Record<string, unknown>
  created_at?: string
  finished_at?: string
  updated_at?: string
}

export interface UserLogUsage {
  prompt_tokens?: number;
  completion_tokens?: number;
  total_tokens?: number;
  cache_read_tokens?: number;
  cache_creation_tokens?: number;
  estimated?: boolean;
}

export type UserLog = {
  id?: number
  model?: string
  created_at?: string
  updated_at?: string
  corr_id?: string
  /** 从 billing_transactions 聚合的净扣费积分，后端字段名 credits_charged */
  credits_charged?: number
  /** @deprecated 旧字段别名，兼容旧代码 */
  cost_credits?: number
  status?: string
  is_stream?: boolean
  error_msg?: string
  client_request?: Record<string, unknown>
  client_response?: Record<string, unknown>
  upstream_headers?: Record<string, unknown>
  upstream_request?: Record<string, unknown>
  upstream_response?: Record<string, unknown>
  latency_ms?: number
  usage?: UserLogUsage
}

export type InviteInfo = {
  invite_code?: string
  invite_count?: number
  frozen_balance?: number
}

export type RedeemRecord = {
  id?: number
  code?: string
  credits?: number
  status?: string
  note?: string
  used_at?: string
  created_at?: string
}

export type WithdrawRecord = {
  id?: number
  created_at?: string
  amount?: number
  payment_type?: string
  status?: string
  admin_remark?: string
}

export interface PaymentOrder {
  id: number;
  out_trade_no: string;
  pay_flat: number;
  pay_from?: string;
  pro_name?: string;
  amount: number;
  credits: number;
  status: string;
  trade_no?: string;
  created_at: string;
  paid_at?: string;
}

export const userApi = {
  getProfile: () => http.get<UserProfileResponse>('/user/profile'),
  getBalance: () => http.get<UserBalanceResponse>('/user/balance'),
  getStats: () => http.get<UserStatsResponse>('/user/stats'),
  getTransactions: (page = 1, size = 20, taskId?: string, corrId?: string) =>
    http.get<{ items?: UserTransaction[]; transactions?: UserTransaction[]; total?: number }>('/user/transactions', {
      params: { page, size, ...(taskId ? { task_id: taskId } : {}), ...(corrId ? { corr_id: corrId } : {}) },
    }),
  listApiKeys: () =>
    http.get<{ api_keys?: ApiKeyRecord[]; keys?: ApiKeyRecord[] } | ApiKeyRecord[]>('/user/apikeys'),
  createApiKey: (name: string, keyType = 'low_price') =>
    http.post<Record<string, unknown>>('/user/apikeys', { name, key_type: keyType }),
  deleteApiKey: (id: number) =>
    http.delete<Record<string, unknown>>(`/user/apikeys/${id}`),
  listChannels: () =>
    http.get<{ channels?: UserChannel[] } | UserChannel[]>('/user/channels'),
  redeemCard: (code: string) =>
    http.post<Record<string, unknown>>('/user/cards/redeem', { code }),
  getRedeemHistory: (page = 1, size = 20) =>
    http.get<{ records?: RedeemRecord[]; list?: RedeemRecord[] } | RedeemRecord[]>('/user/cards/redeem-history', { params: { page, size } }),
  getInviteInfo: () => http.get<InviteInfo>('/user/invite'),
  convertFrozen: (amount = 0) =>
    http.post<Record<string, unknown>>('/user/invite/convert', { amount }),
  getPaymentQR: () =>
    http.get<{ wechat_qr?: string; alipay_qr?: string }>('/user/payment-qr'),
  savePaymentQR: (payload: { wechat_qr?: string; alipay_qr?: string }) =>
    http.put<Record<string, unknown>>('/user/payment-qr', payload),
  submitWithdraw: (amount: number, paymentType: string) =>
    http.post<Record<string, unknown>>('/user/withdraw', { amount, payment_type: paymentType }),
  listWithdrawHistory: (page = 1, size = 20) =>
    http.get<{ records?: WithdrawRecord[]; list?: WithdrawRecord[]; total?: number } | WithdrawRecord[]>('/user/withdraw/history', { params: { page, size } }),
  listTasks: (params: Record<string, unknown> = {}) =>
    http.get<{ items?: UserTask[]; tasks?: UserTask[]; total?: number } | UserTask[]>('/v1/tasks', { params }),
  getTask: (id: number) =>
    http.get<{ task?: UserTask } | UserTask>(`/v1/tasks/${id}`),
  getTaskBilling: (id: number) =>
    http.get<{ transactions?: Array<{ id?: number; type?: string; credits?: number; balance_after?: number; metrics?: Record<string, unknown>; created_at?: string }>; total_charged?: number; total_refunded?: number; net_charged?: number; refunded?: boolean }>(`/v1/tasks/${id}/billing`),
  listLogs: (params: Record<string, unknown> = {}) =>
    http.get<{ items?: UserLog[]; logs?: UserLog[]; total?: number }>('/v1/llm-logs', { params }),
  getLog: (id: number) =>
    http.get<UserLog>(`/v1/llm-logs/${id}`),
  getPaymentOrders: (page = 1, size = 20) =>
    http.get<{ orders: PaymentOrder[]; total: number }>('/user/payment-orders', { params: { page, size } }),
  changePassword: (payload: { old_password: string; new_password: string }) =>
    http.put<Record<string, unknown>>('/user/password', payload),
  bindEmail: (payload: { email: string; code: string }) =>
    http.post<Record<string, unknown>>('/user/bind-email', payload),
}
