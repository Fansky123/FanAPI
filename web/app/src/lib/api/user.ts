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

export type InviteInfo = {
  invite_code?: string
  invite_count?: number
  frozen_balance?: number
}

export type RedeemRecord = {
  code?: string
  credits?: number
  amount?: number
  created_at?: string
  redeemed_at?: string
}

export type WithdrawRecord = {
  id?: number
  created_at?: string
  amount?: number
  payment_type?: string
  status?: string
  admin_remark?: string
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
  redeemCard: (code: string) =>
    http.post<Record<string, unknown>>('/user/cards/redeem', { code }),
  getRedeemHistory: (page = 1, size = 20) =>
    http.get<{ records?: RedeemRecord[]; list?: RedeemRecord[] } | RedeemRecord[]>(
      '/user/cards/redeem-history',
      { params: { page, size } }
    ),
  getInviteInfo: () => http.get<InviteInfo>('/user/invite'),
  convertFrozen: (amount = 0) =>
    http.post<Record<string, unknown>>('/user/invite/convert', { amount }),
  getPaymentQR: () =>
    http.get<{ wechat_qr?: string; alipay_qr?: string }>('/user/payment-qr'),
  savePaymentQR: (payload: { wechat_qr?: string; alipay_qr?: string }) =>
    http.put<Record<string, unknown>>('/user/payment-qr', payload),
  submitWithdraw: (amount: number, paymentType: string) =>
    http.post<Record<string, unknown>>('/user/withdraw', {
      amount,
      payment_type: paymentType,
    }),
  listWithdrawHistory: (page = 1, size = 20) =>
    http.get<{ records?: WithdrawRecord[]; list?: WithdrawRecord[] } | WithdrawRecord[]>(
      '/user/withdraw/history',
      { params: { page, size } }
    ),
  listTasks: (params: Record<string, unknown> = {}) =>
    http.get<{ items?: UserTask[]; tasks?: UserTask[] } | UserTask[]>('/v1/tasks', {
      params,
    }),
  listLogs: (params: Record<string, unknown> = {}) =>
    http.get<{ items?: UserLog[]; logs?: UserLog[] } | UserLog[]>('/v1/llm-logs', {
      params,
    }),
}

export interface PaymentOrder {
  id: number;
  out_trade_no: string;
  pay_type: string;
  amount: number;
  credits: number;
  status: number;
  created_at: string;
  paid_at?: string;
}

export const getPaymentOrders = (page = 1, size = 20) =>
  http.get<{ items: PaymentOrder[]; total: number }>('/user/payment-orders', {
    params: { page, size },
  })

userApi.getPaymentOrders = getPaymentOrders;
