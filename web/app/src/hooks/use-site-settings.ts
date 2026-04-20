import { useEffect, useState } from 'react'

import { publicApi } from '@/lib/api/public'

type SiteSettings = {
  siteName: string
  logoUrl: string
}

const defaultSettings: SiteSettings = {
  siteName: 'FanAPI',
  logoUrl: '',
}

export function useSiteSettings() {
  const [settings, setSettings] = useState<SiteSettings>(defaultSettings)

  useEffect(() => {
    async function load() {
      try {
        const response = await publicApi.getSettings()
        const maybeSettings = (response as { settings?: unknown }).settings
        const record =
          maybeSettings && typeof maybeSettings === 'object'
            ? (maybeSettings as Record<string, string>)
            : (response as Record<string, string>)
        setSettings({
          siteName: record.site_name || 'FanAPI',
          logoUrl: record.logo_url || '',
        })
      } catch {
        setSettings(defaultSettings)
      }
    }

    void load()
  }, [])

  useEffect(() => {
    document.title = settings.siteName
  }, [settings.siteName])

  return settings
}
