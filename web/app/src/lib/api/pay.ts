import { createHttpClient } from './http'

const http = createHttpClient('user')

export const payApi = {
  createPayApplyOrder: (data: { amount: number; pay_flat: number; pay_from?: string }) =>
    http.post<{ out_trade_no?: string; pay_url?: string; wechat_qr?: string; alipay_qr?: string }>('/pay/apply/create', data),
  createEpayOrder: (amount: number, type: string) =>
    http.post<{ pay_url?: string; out_trade_no?: string }>('/pay/epay/create', { amount, type }),
  getOrderStatus: (outTradeNo: string) =>
    http.get<{ status: string }>('/pay/order/status', { params: { out_trade_no: outTradeNo } }),
}
