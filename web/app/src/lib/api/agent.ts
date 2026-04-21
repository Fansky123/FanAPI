import { createHttpClient } from '@/lib/api/http'

const http = createHttpClient('agent')

export type AgentUser = {
  id?: number
  username?: string
  email?: string
  balance_credits?: number
}

export const agentApi = {
  listUsers: (page = 1, size = 50) =>
    http.get<{ items?: AgentUser[]; users?: AgentUser[] } | AgentUser[]>(
      '/agent/users',
      { params: { page, size } }
    ),
  getInvite: () =>
    http.get<Record<string, unknown>>('/agent/invite'),
}
