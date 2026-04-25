import { useEffect, useState } from 'react'
import { publicApi } from '@/lib/api/public'

export type Plan = {
  credits: number
  amount: number
  origin_amount?: number
  desc?: string
  bonus?: number
}

export type SiteSettings = {
  siteName: string
  logoUrl: string
  plans: Plan[]
  epayEnabled: boolean
  payApplyEnabled: boolean
  allowCustom: boolean
  noticeTitle: string
  noticeContent: string
  contactInfo: string
  qqGroupUrl: string
  wechatCsUrl: string
  qrCodeUrl: string
  headerHtml: string
  footerHtml: string
}

const defaultSettings: SiteSettings = {
  siteName: 'FanAPI',
  logoUrl: '',
  plans: [],
  epayEnabled: false,
  payApplyEnabled: false,
  allowCustom: false,
  noticeTitle: '',
  noticeContent: '',
  contactInfo: '',
  qqGroupUrl: '',
  wechatCsUrl: '',
  qrCodeUrl: '',
  headerHtml: '',
  footerHtml: '',
}

export function useSiteSettings() {
  const [settings, setSettings] = useState<SiteSettings>(defaultSettings)
  const [loaded, setLoaded] = useState(false)

  useEffect(() => {
    async function load() {
      try {
        const response = await publicApi.getSettings()
        const maybeSettings = (response as { settings?: unknown }).settings
        const record =
          maybeSettings && typeof maybeSettings === 'object'
            ? (maybeSettings as Record<string, any>)
            : (response as Record<string, any>)
        setSettings({
          siteName: record.site_name || 'FanAPI',
          logoUrl: record.logo_url || '',
          plans: (() => {
            try { return JSON.parse(record.recharge_plans || '[]') } catch { return [] }
          })(),
          epayEnabled: record.epay_enabled === 'true',
          payApplyEnabled: record.pay_apply_enabled === 'true',
          allowCustom: record.recharge_allow_custom !== 'false',
          noticeTitle: record.notice_title || '',
          noticeContent: record.notice_content || '',
          contactInfo: record.contact_info || '',
          qqGroupUrl: record.qq_group_url || '',
          wechatCsUrl: record.wechat_cs_url || '',
          qrCodeUrl: record.qrcode_url || '',
          headerHtml: record.header_html || '',
          footerHtml: record.footer_html || '',
        })
      } catch {
        setSettings(defaultSettings)
      } finally {
        setLoaded(true)
      }
    }

    void load()
  }, [])

  useEffect(() => {
    document.title = settings.siteName
  }, [settings.siteName])

  return { settings, loaded }
}
