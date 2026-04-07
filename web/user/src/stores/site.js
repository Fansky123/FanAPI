import { defineStore } from 'pinia'
import { ref } from 'vue'
import { publicApi } from '@/api'

export const useSiteStore = defineStore('site', () => {
  const siteName = ref('FanAPI')
  const logoUrl = ref('')
  const headerHtml = ref('')
  const footerHtml = ref('')
  const epayEnabled = ref(false)
  const loaded = ref(false)

  async function fetchSettings() {
    if (loaded.value) return
    try {
      const res = await publicApi.getSettings()
      const s = res.settings || {}
      if (s.site_name) siteName.value = s.site_name
      if (s.logo_url) logoUrl.value = s.logo_url
      if (s.header_html) headerHtml.value = s.header_html
      if (s.footer_html) footerHtml.value = s.footer_html
      epayEnabled.value = s.epay_enabled === 'true'
      loaded.value = true
    } catch {
      // 静默失败，使用默认值
    }
  }

  return { siteName, logoUrl, headerHtml, footerHtml, epayEnabled, loaded, fetchSettings }
})
