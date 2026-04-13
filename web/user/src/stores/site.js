import { defineStore } from 'pinia'
import { ref } from 'vue'
import { publicApi } from '@/api'

export const useSiteStore = defineStore('site', () => {
  const siteName = ref('FanAPI')
  const logoUrl = ref('')
  const headerHtml = ref('')
  const footerHtml = ref('')
  const epayEnabled = ref(false)
  const payApplyEnabled = ref(false)
  const noticeTitle = ref('')
  const noticeContent = ref('')
  const contactInfo = ref('')
  const qrcodeUrl = ref('')
  const loaded = ref(false)

  async function fetchSettings() {
    if (loaded.value) return
    try {
      const res = await publicApi.getSettings()
      const s = res.settings || {}
      if (s.site_name) {
        siteName.value = s.site_name
        document.title = s.site_name
      }
      if (s.logo_url) logoUrl.value = s.logo_url
      if (s.header_html) headerHtml.value = s.header_html
      if (s.footer_html) footerHtml.value = s.footer_html
      epayEnabled.value = s.epay_enabled === 'true'
      payApplyEnabled.value = s.pay_apply_enabled === 'true'
      if (s.notice_title !== undefined) noticeTitle.value = s.notice_title
      if (s.notice_content !== undefined) noticeContent.value = s.notice_content
      if (s.contact_info !== undefined) contactInfo.value = s.contact_info
      if (s.qrcode_url !== undefined) qrcodeUrl.value = s.qrcode_url
      loaded.value = true
    } catch {
      // 静默失败，使用默认值
    }
  }

  const darkMode = ref(localStorage.getItem('dark_mode') === 'true')

  function toggleDark() {
    darkMode.value = !darkMode.value
    localStorage.setItem('dark_mode', darkMode.value)
    document.documentElement.classList.toggle('dark', darkMode.value)
  }

  return { siteName, logoUrl, headerHtml, footerHtml, epayEnabled, payApplyEnabled, noticeTitle, noticeContent, contactInfo, qrcodeUrl, loaded, fetchSettings, darkMode, toggleDark }
})
