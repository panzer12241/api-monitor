import axios from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'

// Create axios instance with base configuration
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Add request interceptor to include auth token
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('api-monitor-token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Add response interceptor to handle errors
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Token expired or invalid, redirect to login
      localStorage.removeItem('api-monitor-token')
      localStorage.removeItem('api-monitor-user')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export const authAPI = {
  login: (credentials) => apiClient.post('/auth/login', credentials),
  register: (userData) => apiClient.post('/auth/register', userData)
}

export const endpointsAPI = {
  getAll: () => apiClient.get('/endpoints'),
  create: (endpoint) => apiClient.post('/endpoints', endpoint),
  update: (id, endpoint) => apiClient.put(`/endpoints/${id}`, endpoint),
  delete: (id) => apiClient.delete(`/endpoints/${id}`),
  toggle: (id) => apiClient.post(`/endpoints/${id}/toggle`),
  getLogs: (id, params = {}) => apiClient.get(`/endpoints/${id}/logs`, { params }),
  manualCheck: (id) => apiClient.post(`/endpoints/${id}/check`)
}

export const proxiesAPI = {
  getAll: () => apiClient.get('/proxies'),
  create: (proxy) => apiClient.post('/proxies', proxy),
  update: (id, proxy) => apiClient.put(`/proxies/${id}`, proxy),
  delete: (id) => apiClient.delete(`/proxies/${id}`),
  toggle: (id) => apiClient.post(`/proxies/${id}/toggle`)
}

export const usersAPI = {
  getAll: () => apiClient.get('/users'),
  create: (user) => apiClient.post('/users', user),
  update: (id, user) => apiClient.put(`/users/${id}`, user),
  delete: (id) => apiClient.delete(`/users/${id}`)
}

export default apiClient
