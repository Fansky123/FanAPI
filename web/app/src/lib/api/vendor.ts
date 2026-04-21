import { createHttpClient } from '@/lib/api/http'

const http = createHttpClient('vendor')

export type VendorProfile = {
  name?: string
  username?: string
  email?: string
  balance?: number
  commission_ratio?: number
}

export type VendorKey = {
  id?: number
  pool_id?: number
  channel_id?: number
  channel_name?: string
  masked_value?: string
  key?: string
  key_type?: string
  total_cost?: number
  total_profit?: number
  my_earn?: number
  is_active?: boolean
  created_at?: string
}

export type VendorPool = {
  id?: number
  name?: string
  channel_id?: number
  channel_name?: string
  channel_type?: string
}

export const vendorApi = {
  getProfile: () => http.get<VendorProfile>('/vendor/profile'),
  getKeys: () =>
    http.get<{ items?: VendorKey[]; keys?: VendorKey[] } | VendorKey[]>(
      '/vendor/keys'
    ),
  getPools: () =>
    http.get<{ pools?: VendorPool[] } | VendorPool[]>('/vendor/pools'),
  submitKey: (payload: { pool_id?: number | null; channel_id?: number; value: string }) =>
    http.post<Record<string, unknown>>('/vendor/keys', payload),
}
