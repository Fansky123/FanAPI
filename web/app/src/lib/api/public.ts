import { createHttpClient } from '@/lib/api/http'

const http = createHttpClient()

export type LoginResponse = {
  token: string
  inviter_wechat_qr?: string
  user?: {
    username?: string
  }
}

export const publicApi = {
  getSettings: () =>
    http.get<{ settings?: Record<string, string> } | Record<string, string>>(
      '/public/settings'
    ),
  listChannels: () => http.get<Record<string, unknown>>('/public/channels'),
}

export const authApi = {
  login: (payload: { username: string; password: string }) =>
    http.post<LoginResponse>('/auth/login', payload),
  register: (payload: { username: string; email?: string; password: string; invite_code?: string }) =>
    http.post<Record<string, unknown>>('/auth/register', payload),
  forgotPassword: (email: string) =>
    http.post<Record<string, unknown>>('/auth/forgot-password', { email }),
  resetPassword: (payload: { email: string; code: string; password: string }) =>
    http.post<Record<string, unknown>>('/auth/reset-password', payload),
  sendCode: (email: string) =>
    http.post<Record<string, unknown>>('/auth/send-code', { email }),
}
