import { defineStore } from 'pinia'
import { ref } from 'vue'
import { userApi } from '@/api'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const username = ref(localStorage.getItem('user_username') || '')
  const email = ref(localStorage.getItem('user_email') || '')
  const group = ref(localStorage.getItem('user_group') || '')
  const balance = ref(0)

  function setToken(t) {
    token.value = t
    localStorage.setItem('token', t)
  }

  function setUsername(u) {
    username.value = u
    localStorage.setItem('user_username', u)
  }

  function setEmail(e) {
    email.value = e
    localStorage.setItem('user_email', e)
  }

  function setGroup(g) {
    group.value = g
    localStorage.setItem('user_group', g)
  }

  function logout() {
    token.value = ''
    username.value = ''
    email.value = ''
    group.value = ''
    localStorage.removeItem('token')
    localStorage.removeItem('user_username')
    localStorage.removeItem('user_email')
    localStorage.removeItem('user_group')
  }

  async function fetchBalance() {
    const res = await userApi.getBalance()
    balance.value = res.balance_credits ?? 0
  }

  async function fetchProfile() {
    try {
      const res = await userApi.getProfile()
      if (res.username) setUsername(res.username)
      if (res.email) setEmail(res.email)
      setGroup(res.group ?? '')
    } catch {
      // 忽略——用户可能尚未登录
    }
  }

  return { token, username, email, group, balance, setToken, setUsername, setEmail, setGroup, logout, fetchBalance, fetchProfile }
})
