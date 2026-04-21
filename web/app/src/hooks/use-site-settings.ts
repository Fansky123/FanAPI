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
}

const defaultSettings: SiteSettings = {
  siteName: 'FanAPI',
  logoUrl: '',
  plans: [],
  epayEnabled: false,
  payApplyEnabled: false,
  allowCustom: false,
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
          plans: Array.isArray(record.plans) ? record.plans : [],
          epayEnabled: !!record.epay_enabled,
          payApplyEnabled: !!record.pay_apply_enabled,
          allowCustom: !!record.allow_custom,
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
