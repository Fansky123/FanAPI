import { useStore } from './store'
import { createDefaultOpenAIProfile } from './lib/apiProfiles'

export async function bootstrapMidCode() {
  const fragment = new URLSearchParams(window.location.hash.slice(1))
  const code = fragment.get('midcode_code')
  if (!code) return
  // Remove the ticket before network requests, rendering, or history persistence.
  window.history.replaceState(null, '', window.location.pathname + window.location.search)
  if (!/^[a-f0-9]{64}$/.test(code)) throw new Error('跳转凭证无效，请从 MidCode 重新进入。')
  const response = await fetch('/midcode/exchange', {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ code }), cache: 'no-store', credentials: 'omit',
    signal: AbortSignal.timeout(15000),
  })
  if (!response.ok) throw new Error('跳转已过期或暂时不可用，请从 MidCode 重新进入。')
  const data = await response.json()
  if (typeof data.apiKey !== 'string' || !data.apiKey || typeof data.model !== 'string' || !data.model) {
    throw new Error('未收到有效的工作台配置，请从 MidCode 重新进入。')
  }
  const entries = Array.isArray(data.profiles) && data.profiles.length ? data.profiles : [{id:'midcode',name:'MidCode',apiKey:data.apiKey,model:data.model}]
  const profiles = entries.map((entry: {id:string;name:string;apiKey:string;model:string;models?:string[]}) => {
    if (!entry.id || !entry.apiKey || !entry.model) throw new Error('工作台密钥配置无效，请重新进入。')
    return createDefaultOpenAIProfile({
      id: entry.id, name: entry.name + (entry.models?.length === 0 ? '（未绑定图片模型）' : ''), baseUrl: `${window.location.origin}/midcode-api/v1`,
      apiKey: entry.apiKey, model: entry.model, apiMode: 'images',
      apiProxy: false, responseFormatB64Json: true, streamImages: false, streamPartialImages: 0, codexCli: false,
    })
  })
  const activeProfileId = profiles.some((profile: {id:string}) => profile.id === data.activeProfileId) ? data.activeProfileId : profiles[0].id
  useStore.getState().setSettings({ profiles, activeProfileId })
  useStore.getState().setAppMode('gallery')
  useStore.getState().clearInputImages()
  useStore.getState().clearMaskDraft()
  useStore.getState().setPrompt(typeof data.prompt === 'string' ? data.prompt : '')
}

export function showMidCodeError(error: unknown) {
  const root = document.getElementById('root')!
  const message = document.createElement('p')
  message.setAttribute('role', 'alert')
  message.textContent = error instanceof Error ? error.message : '无法进入图片工作台，请返回 MidCode 重试。'
  root.replaceChildren(message)
}
