import { createHttpClient } from '@/lib/api/http'
import { uploadAuthedImage, uploadAuthedVideo, type UploadImageCategory, type UploadVideoCategory } from '@/lib/api/upload'

const http = createHttpClient('user')

export type UserProfileResponse = {
  id?: number
  username?: string
  email?: string
  group?: string
  balance?: number
}

export type UserBalanceResponse = {
  balance_cny?: number
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
  model_credit_charged?: number
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
  model_groups?: ApiKeyModelGroupBinding[]
  needs_group_binding?: boolean
  is_active?: boolean
  total_consumed?: number
  today_consumed?: number
  last_used_at?: string | null
  created_at?: string
}

export type UserChannel = {
  id?: number
  group_id?: number
  group_name?: string
  group_priority?: number
  name?: string
  routing_model?: string
  model?: string
  model_provider?: string
  description?: string
  type?: string
  category?: string
  protocol?: string
  billing_type?: string
  icon_url?: string
  price_display?: string
  group_price?: string
  group_prices?: Array<{
    group_id: number
    group_code: string
    group_name: string
    price_display?: string
    vip_prices?: Array<{ code: string; name: string; price_display?: string }>
  }>
}

export type ModelAvailabilityRecent = {
  status: string
  latency_ms: number
  created_at: string
}

export type ModelAvailability = {
  routing_model: string
  total: number
  success: number
  availability_percent: number
  p50_latency_ms: number
  recent: ModelAvailabilityRecent[]
}

export type UserTask = {
  id?: number
  task_id?: number
  task_type?: string
  type?: string
  code?: number
  url?: string        // top-level image/video URL (from TaskResult)
  items?: unknown[]   // top-level items array (music, multi-result tasks)
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
  requested_tier?: 'standard' | 'fast';
  actual_tier?: 'standard' | 'fast';
  tier_confirmed?: boolean;
  tier_downgraded?: boolean;
  tier_unconfirmed?: boolean;
  actual_service_tier?: string;
  actual_speed?: string;
}

export type ApiKeyModelGroup = {
  id?: number
  code?: string
  name?: string
  model_provider_id?: number
  model_provider?: string
  model_provider_active?: boolean
  model_provider_sort_order?: number
  description?: string
  is_active?: boolean
  model_count?: number
  official_discount_bps?: number
  official_discount_status?: 'available' | 'unavailable' | 'inconsistent'
}

export type ApiKeyModelGroupBinding = {
  id?: number
  api_key_id?: number
  group_id?: number
  priority?: number
  group?: ApiKeyModelGroup
}

export type TokenUsageRow = {
  model: string
  non_cached_input_tokens: number
  cache_read_tokens: number
  cache_creation_tokens: number
  output_tokens: number
  total_tokens: number
}

export type TokenStatsResponse = {
  items: TokenUsageRow[]
  page: number
  page_size: number
  total: number
  start_at: string
  end_at: string
}

export type UserLog = {
  id?: number
  model?: string
  created_at?: string
  updated_at?: string
  corr_id?: string
  input_price_per_1m_tokens?: number
  output_price_per_1m_tokens?: number
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
  inviter_id?: number | null
  can_bind?: boolean
  binding_deadline?: string
}

export type InviteeRecord = {
  id?: number
  username?: string
  total_recharge?: number
  total_spend?: number
  created_at?: string
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
  pay_channel?: string;
  pay_from?: string;
  pro_name?: string;
  amount: number;
  credits: number;
  status: string;
  trade_no?: string;
  created_at: string;
  paid_at?: string;
}

export type UserModelCredit = {
  id?: number
  model_name?: string
  credits?: number
}

export type ConversationMessage = {
  role: 'user' | 'assistant' | 'system'
  content: string
  images?: Array<{ src: string; width?: number; height?: number }>
  image_request?: { model: string; resolution: 'default' | '2K' | '4K'; ratio: '16:9' | '1:1'; edited: boolean }
}

export type ChatConversation = {
  id: number
  title: string
  model: string
  messages: ConversationMessage[]
  created_at: string
  updated_at: string
}

export const userApi = {
  getProfile: () => http.get<UserProfileResponse>('/user/profile'),
  getBalance: () => http.get<UserBalanceResponse>('/user/balance'),
  getStats: () => http.get<UserStatsResponse>('/user/stats'),
  getTokenStats: (params: { start_at: string; end_at: string; model?: string; page: number; page_size: number }) =>
    http.get<TokenStatsResponse>('/user/stats/tokens', { params }),
  getTransactions: (page = 1, size = 20, taskId?: string, corrId?: string, apiKeyId?: number) =>
    http.get<{ items?: UserTransaction[]; transactions?: UserTransaction[]; total?: number }>('/user/transactions', {
      params: { page, size, ...(taskId ? { task_id: taskId } : {}), ...(corrId ? { corr_id: corrId } : {}), ...(apiKeyId ? { api_key_id: apiKeyId } : {}) },
    }),
  listApiKeys: () =>
    http.get<{ api_keys?: ApiKeyRecord[]; keys?: ApiKeyRecord[] } | ApiKeyRecord[]>('/user/apikeys'),
  createApiKey: (name: string, groupIds: number[]) =>
    http.post<Record<string, unknown>>('/user/apikeys', { name, group_ids: groupIds }),
  listModelGroups: () =>
    http.get<{ groups?: ApiKeyModelGroup[] }>('/user/model-groups'),
  listApiKeyModelGroups: (id: number) =>
    http.get<{ groups?: ApiKeyModelGroupBinding[] }>(`/user/apikeys/${id}/model-groups`),
  replaceApiKeyModelGroups: (id: number, groupIds: number[]) =>
    http.put<{ groups?: ApiKeyModelGroupBinding[] }>(`/user/apikeys/${id}/model-groups`, { group_ids: groupIds }),
  deleteApiKey: (id: number) =>
    http.delete<Record<string, unknown>>(`/user/apikeys/${id}`),
  listChannels: () =>
    http.get<{ channels?: UserChannel[] } | UserChannel[]>('/user/channels'),
  getModelAvailability: () =>
    http.get<{ models: ModelAvailability[] }>('/user/model-availability'),
  redeemCard: (code: string) =>
    http.post<{ credits?: number; message?: string }>('/user/cards/redeem', { code }),
  getRedeemHistory: (page = 1, size = 20) =>
    http.get<{ records?: RedeemRecord[]; list?: RedeemRecord[] } | RedeemRecord[]>('/user/cards/redeem-history', { params: { page, size } }),
  getInviteInfo: () => http.get<InviteInfo>('/user/invite'),
  bindInviteCode: (inviteCode: string) =>
    http.post<{ message?: string; inviter_id?: number }>('/user/invite/bind', { invite_code: inviteCode }),
  getInviteeList: () => http.get<{ invitees: InviteeRecord[] }>('/user/invite/list'),
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
  validateCoupon: (code: string, amount: number) =>
    http.get<{ valid: boolean; coupon_id: number; discount_yuan: number; final_amount: number; discount_type: string; discount_value: number }>('/user/coupons/validate', { params: { code, amount } }),
  listTasks: (params: Record<string, unknown> = {}) =>
    http.get<{ items?: UserTask[]; tasks?: UserTask[]; total?: number } | UserTask[]>('/v1/tasks', { params }),
  clearTaskHistory: (type?: string) =>
    http.delete<Record<string, unknown>>('/v1/tasks/history', { params: type ? { type } : {} }),
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
  changePassword: (payload: { new_password: string }) =>
    http.put<Record<string, unknown>>('/user/password', payload),
  bindEmail: (payload: { email: string; code: string }) =>
    http.post<Record<string, unknown>>('/user/bind-email', payload),
  uploadReferenceImage: (file: File) => {
    return uploadAuthedImage('user', file, 'reference')
  },
  uploadImage: (file: File, category: UploadImageCategory) =>
    uploadAuthedImage('user', file, category),
  uploadVideo: (file: File, category: UploadVideoCategory) =>
    uploadAuthedVideo('user', file, category),
  getModelCredits: () =>
    http.get<{ model_credits?: UserModelCredit[] }>('/user/model-credits'),
  listConversations: (size = 50) =>
    http.get<{ items: ChatConversation[] }>('/v1/conversations', { params: { size } }),
  saveConversation: (payload: { id?: number; title: string; model: string; messages: ConversationMessage[] }) =>
    http.post<ChatConversation>('/v1/conversations', payload),
  deleteConversation: (id: number) =>
    http.delete<{ ok: boolean }>(`/v1/conversations/${id}`),
}
