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
  key?: string
  key_type?: string
  total_cost?: number
  total_profit?: number
}

export const vendorApi = {
  getProfile: () => http.get<VendorProfile>('/vendor/profile'),
  getKeys: () =>
    http.get<{ items?: VendorKey[]; keys?: VendorKey[] } | VendorKey[]>(
      '/vendor/keys'
    ),
}
