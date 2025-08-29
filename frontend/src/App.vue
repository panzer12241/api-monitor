<template>
  <v-app>
    <v-app-bar
      v-if="authStore.isAuthenticated"
      app
      color="primary"
      dark
      elevation="1"
    >
      <v-app-bar-nav-icon @click="drawer = !drawer"></v-app-bar-nav-icon>
      
      <v-toolbar-title>
        <v-icon left class="mr-2">mdi-monitor-dashboard</v-icon>
        API Monitor
      </v-toolbar-title>
      
      <v-spacer></v-spacer>
      
      <v-menu bottom left>
        <template v-slot:activator="{ props }">
          <v-btn icon v-bind="props">
            <v-avatar size="32">
              <v-icon>mdi-account-circle</v-icon>
            </v-avatar>
          </v-btn>
        </template>
        
        <v-list>
          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>{{ authStore.user?.name }}</v-list-item-title>
              <v-list-item-subtitle>{{ authStore.user?.username }}</v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
          
          <v-divider></v-divider>
          
          <v-list-item @click="logout">
            <v-list-item-icon>
              <v-icon>mdi-logout</v-icon>
            </v-list-item-icon>
            <v-list-item-content>
              <v-list-item-title>Logout</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
        </v-list>
      </v-menu>
    </v-app-bar>

    <v-navigation-drawer
      v-if="authStore.isAuthenticated"
      v-model="drawer"
      app
      temporary
    >
      <v-list>
        <v-list-item
          v-for="item in menuItems"
          :key="item.title"
          :to="item.route"
          link
        >
          <v-list-item-icon>
            <v-icon>{{ item.icon }}</v-icon>
          </v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>{{ item.title }}</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
      </v-list>
    </v-navigation-drawer>

    <v-main>
      <router-view />
    </v-main>

    <v-snackbar
      v-model="showSnackbar"
      :color="snackbarColor"
      timeout="3000"
      top
    >
      {{ snackbarMessage }}
      <template v-slot:action="{ attrs }">
        <v-btn
          color="white"
          text
          v-bind="attrs"
          @click="showSnackbar = false"
        >
          Close
        </v-btn>
      </template>
    </v-snackbar>
  </v-app>
</template>

<script>
import { useAuthStore } from '@/stores/auth'

export default {
  name: 'App',

  data() {
    return {
      drawer: false,
      showSnackbar: false,
      snackbarMessage: '',
      snackbarColor: 'success',
      menuItems: [
        {
          title: 'Dashboard',
          icon: 'mdi-view-dashboard',
          route: '/'
        },
        {
          title: 'Proxy Management',
          icon: 'mdi-server-network',
          route: '/proxies'
        }
      ]
    }
  },

  setup() {
    const authStore = useAuthStore()
    return { authStore }
  },

  mounted() {
    // Check authentication status on app load
    this.authStore.checkAuth()
  },

  methods: {
    logout() {
      this.authStore.logout()
      this.$router.push('/login')
      this.showMessage('Logged out successfully', 'info')
    },

    showMessage(message, color = 'success') {
      this.snackbarMessage = message
      this.snackbarColor = color
      this.showSnackbar = true
    }
  }
}
</script>

<style>
.v-application {
  font-family: 'Roboto', sans-serif;
}
</style>
