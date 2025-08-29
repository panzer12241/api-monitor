import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import Dashboard from '@/views/Dashboard.vue'
import Endpoints from '@/views/Endpoints.vue'
import Login from '@/views/Login.vue'
import Grafana from '@/views/Grafana.vue'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { requiresGuest: true }
  },
  {
    path: '/',
    name: 'Dashboard',
    component: Dashboard,
    meta: { requiresAuth: true }
  },
  {
    path: '/endpoints',
    name: 'Endpoints',
    component: Endpoints,
    meta: { requiresAuth: true }
  },
  {
    path: '/grafana',
    name: 'Grafana',
    component: Grafana,
    meta: { requiresAuth: true }
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

// Navigation guards
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  
  // Check if route requires authentication
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    // Check if user has valid token in localStorage
    if (!authStore.checkAuth()) {
      next('/login')
      return
    }
  }
  
  // Check if route is for guests only (like login)
  if (to.meta.requiresGuest && authStore.isAuthenticated) {
    next('/')
    return
  }
  
  next()
})

export default router
