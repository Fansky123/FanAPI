import { defineStore } from 'pinia'
import { ref } from 'vue'
import { userApi } from '@/api'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const username = ref(localStorage.getItem('user_username') || '')
  const email = ref(localStorage.getItem('user_email') || '')
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

  function logout() {
    token.value = ''
    username.value = ''
    email.value = ''
    localStorage.removeItem('token')
    localStorage.removeItem('user_username')
    localStorage.removeItem('user_email')
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
    } catch {
      // ignore — may not be logged in
    }
  }

  return { token, username, email, balance, setToken, setUsername, setEmail, logout, fetchBalance, fetchProfile }
})
