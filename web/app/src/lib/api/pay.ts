import { http } from './http'

export const payApi = {
  createEpayOrder: (amount: number, type: string) =>
    http.post<{ pay_url: string; out_trade_no: string }>('/pay/epay/create', {
      amount,
      type,
    }),
  createPayApplyOrder: (payload: { amount: number; pay_flat: number; pay_from: string }) =>
    http.post<{ pay_url: string; out_trade_no: string }>('/pay/apply/create', payload),
  getOrderStatus: (out_trade_no: string) =>
    http.get<{ status: number }>('/pay/order/status', { params: { out_trade_no } }),
}
