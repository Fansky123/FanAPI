type Role = 'user' | 'admin' | 'agent' | 'vendor'

const TOKEN_KEYS: Record<Role, string> = {
  user: 'token',
  admin: 'admin_token',
  agent: 'agent_token',
  vendor: 'vendor_token',
}

const MODE_KEY = 'fanapi_ui_mode'
const DARK_KEY = 'dark_mode'

export function getRoleToken(role: Role) {
  return window.localStorage.getItem(TOKEN_KEYS[role]) ?? ''
}

export function setRoleToken(role: Role, value: string) {
  window.localStorage.setItem(TOKEN_KEYS[role], value)
}

export function clearRoleToken(role: Role) {
  window.localStorage.removeItem(TOKEN_KEYS[role])
}

export function getSiteModePreference() {
  return window.localStorage.getItem(MODE_KEY) ?? 'user'
}

export function setSiteModePreference(mode: Role) {
  window.localStorage.setItem(MODE_KEY, mode)
}

export function isDarkModeEnabled() {
  return window.localStorage.getItem(DARK_KEY) === 'true'
}

export function setDarkMode(enabled: boolean) {
  window.localStorage.setItem(DARK_KEY, String(enabled))
  document.documentElement.classList.toggle('dark', enabled)
}
