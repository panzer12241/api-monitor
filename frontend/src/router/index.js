import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '../views/Dashboard.vue'
import Endpoints from '../views/Endpoints.vue'

const routes = [
  {
    path: '/',
    name: 'Dashboard',
    component: Dashboard
  },
  {
    path: '/endpoints',
    name: 'Endpoints',
    component: Endpoints
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
