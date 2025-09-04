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
        // Call the Go backend API
        const response = await axios.post('http://localhost:8080/api/v1/auth/login', {
          username: credentials.username,
          password: credentials.password
        })

        if (response.data && response.data.token) {
          const { token, user } = response.data
          
          this.isAuthenticated = true
          this.user = user
          this.token = token
          
          localStorage.setItem('api-monitor-token', token)
          localStorage.setItem('api-monitor-user', JSON.stringify(user))
          
          // Set axios default header
          axios.defaults.headers.common['Authorization'] = `Bearer ${token}`
          
          return { success: true }
        } else {
          throw new Error('Invalid response from server')
        }
      } catch (error) {
        const errorMessage = error.response?.data?.error || error.message || 'Login failed'
        return { 
          success: false, 
          error: errorMessage
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
