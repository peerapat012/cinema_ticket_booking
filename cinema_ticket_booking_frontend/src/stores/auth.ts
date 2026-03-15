import { defineStore } from 'pinia'
import api from '@/services/api'
import type { User } from '@/types'

interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    user: null,
    token: localStorage.getItem('token'),
    isAuthenticated: !!localStorage.getItem('token'),
  }),

  actions: {
    async checkAuth() {
      const token = localStorage.getItem('token')
      console.log('checkAuth - token from localStorage:', token ? 'yes' : 'no')
      if (token) {
        api.setToken(token)
        try {
          const user = await api.getCurrentUser()
          this.user = user
          this.isAuthenticated = true
          console.log('checkAuth - user authenticated:', user.username)
        } catch (e) {
          console.log('checkAuth - failed to get user, logging out')
          this.logout()
        }
      } else {
        console.log('checkAuth - no token found')
      }
    },

    logout() {
      this.user = null
      this.token = null
      this.isAuthenticated = false
      localStorage.removeItem('token')
      api.setToken(null)
    },

    setAuth(token: string, user: User) {
      this.token = token
      this.user = user
      this.isAuthenticated = true
      localStorage.setItem('token', token)
      api.setToken(token)
    },
  },
})
