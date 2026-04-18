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
  const plans = ref([])
  const allowCustom = ref(true)
  const qqGroupUrl = ref('')
  const wechatCsUrl = ref('')
  const defaultRebateRatio = ref(0)
  const defaultVendorCommission = ref(0)
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
      if (s.recharge_plans) {
        try { plans.value = JSON.parse(s.recharge_plans) } catch { plans.value = [] }
      }
      allowCustom.value = s.recharge_allow_custom !== 'false'
      if (s.qq_group_url !== undefined) qqGroupUrl.value = s.qq_group_url
      if (s.wechat_cs_url !== undefined) wechatCsUrl.value = s.wechat_cs_url
      if (s.default_rebate_ratio) defaultRebateRatio.value = parseFloat(s.default_rebate_ratio) || 0
      if (s.default_vendor_commission) defaultVendorCommission.value = parseFloat(s.default_vendor_commission) || 0
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

  return {
    siteName, logoUrl, headerHtml, footerHtml, epayEnabled, payApplyEnabled,
    noticeTitle, noticeContent, contactInfo, qrcodeUrl, plans, allowCustom,
    qqGroupUrl, wechatCsUrl, defaultRebateRatio, defaultVendorCommission,
    loaded, fetchSettings, darkMode, toggleDark,
  }
})
