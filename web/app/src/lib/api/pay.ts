import { createHttpClient } from './http'

const http = createHttpClient('user')

export const payApi = {
  createPayApplyOrder: (data: { amount: number; pay_flat?: number; pay_from?: string }) =>
    http.post<{ out_trade_no?: string; pay_url?: string; wechat_qr?: string; alipay_qr?: string }>('/user/pay-apply/create', data),
  createEpayOrder: (amount: number, pay_type: string) =>
    http.post<{ pay_url?: string; out_trade_no?: string }>('/user/epay/create', { amount, pay_type }),
  getOrderStatus: (outTradeNo: string) =>
    http.get<{ status: number; credits?: number }>(`/user/pay/status/${outTradeNo}`),
}
