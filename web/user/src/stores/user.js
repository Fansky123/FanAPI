import { defineStore } from 'pinia'
import { ref } from 'vue'
import { userApi } from '@/api'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const balance = ref(0)

  function setToken(t) {
    token.value = t
    localStorage.setItem('token', t)
  }

  function logout() {
    token.value = ''
    localStorage.removeItem('token')
  }

  async function fetchBalance() {
    const res = await userApi.getBalance()
    balance.value = res.balance_credits ?? 0
  }

  return { token, balance, setToken, logout, fetchBalance }
})
