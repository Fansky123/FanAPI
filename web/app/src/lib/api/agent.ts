import { createHttpClient } from '@/lib/api/http'

const http = createHttpClient('agent')

export type AgentUser = {
  id?: number
  username?: string
  email?: string
  balance?: number
  balance_credits?: number
  total_recharge?: number
  total_spend?: number
  created_at?: string
}

export type AgentInvite = {
  invite_code?: string
  wechat_qr?: string
}

export type AgentUsersResponse = {
  users?: AgentUser[]
  items?: AgentUser[]
  total?: number
}

export const agentApi = {
  listUsers: (page = 1, size = 50) =>
    http.get<AgentUsersResponse | AgentUser[]>(
      '/agent/users',
      { params: { page, size } }
    ),
  rechargeUser: (id: number, amount: number) =>
    http.post<void>(`/agent/users/${id}/recharge`, { amount }),
  getInvite: () =>
    http.get<AgentInvite>('/agent/invite'),
  updateWechatQR: (wechat_qr: string) =>
    http.put<void>('/agent/wechat-qr', { wechat_qr }),
}
