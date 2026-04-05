import { defineStore } from 'pinia'
import { ref } from 'vue'
import { userApi } from '@/api'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const email = ref(localStorage.getItem('user_email') || '')
  const balance = ref(0)

  function setToken(t) {
    token.value = t
    localStorage.setItem('token', t)
  }

  function setEmail(e) {
    email.value = e
    localStorage.setItem('user_email', e)
  }

  function logout() {
    token.value = ''
    email.value = ''
    localStorage.removeItem('token')
    localStorage.removeItem('user_email')
  }

  async function fetchBalance() {
    const res = await userApi.getBalance()
    balance.value = res.balance_credits ?? 0
  }

  return { token, email, balance, setToken, setEmail, logout, fetchBalance }
})
