import { defineStore } from 'pinia'
import axios from 'axios'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    isAuthenticated: false,
    user: null,
    token: localStorage.getItem('api-monitor-token') || null
  }),

  getters: {
    isLoggedIn: (state) => state.isAuthenticated
  },

  actions: {
    async login(credentials) {
      try {
        // For demo purposes, we'll use simple validation
        // In production, this should call your backend API
        if (credentials.username === 'admin' && credentials.password === 'admin123') {
          const token = 'demo-token-' + Date.now()
          
          this.isAuthenticated = true
          this.user = {
            id: 1,
            username: credentials.username,
            name: 'Administrator'
          }
          this.token = token
          
          localStorage.setItem('api-monitor-token', token)
          localStorage.setItem('api-monitor-user', JSON.stringify(this.user))
          
          // Set axios default header
          axios.defaults.headers.common['Authorization'] = `Bearer ${token}`
          
          return { success: true }
        } else {
          throw new Error('Invalid credentials')
        }
      } catch (error) {
        return { 
          success: false, 
          error: error.message || 'Login failed' 
        }
      }
    },

    logout() {
      this.isAuthenticated = false
      this.user = null
      this.token = null
      
      localStorage.removeItem('api-monitor-token')
      localStorage.removeItem('api-monitor-user')
      
      delete axios.defaults.headers.common['Authorization']
    },

    checkAuth() {
      const token = localStorage.getItem('api-monitor-token')
      const userStr = localStorage.getItem('api-monitor-user')
      
      if (token && userStr) {
        try {
          this.token = token
          this.user = JSON.parse(userStr)
          this.isAuthenticated = true
          
          axios.defaults.headers.common['Authorization'] = `Bearer ${token}`
          return true
        } catch (error) {
          this.logout()
          return false
        }
      }
      return false
    }
  }
})
