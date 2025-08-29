<template>
  <v-app>
    <v-main>
      <v-container fluid class="fill-height">
        <v-row align="center" justify="center">
          <v-col cols="12" sm="8" md="4">
            <v-card class="elevation-12">
              <v-toolbar color="primary" dark flat>
                <v-toolbar-title>
                  <v-icon left class="mr-2">mdi-monitor-dashboard</v-icon>
                  API Monitor Login
                </v-toolbar-title>
              </v-toolbar>
              
              <v-card-text>
                <v-form ref="loginForm" v-model="valid" @submit.prevent="handleLogin">
                  <v-text-field
                    v-model="credentials.username"
                    label="Username"
                    prepend-icon="mdi-account"
                    :rules="usernameRules"
                    required
                    outlined
                  ></v-text-field>
                  
                  <v-text-field
                    v-model="credentials.password"
                    label="Password"
                    prepend-icon="mdi-lock"
                    :type="showPassword ? 'text' : 'password'"
                    :append-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
                    @click:append="showPassword = !showPassword"
                    :rules="passwordRules"
                    required
                    outlined
                  ></v-text-field>
                  
                  <v-alert
                    v-if="errorMessage"
                    type="error"
                    dismissible
                    v-model="showError"
                    class="mb-4"
                  >
                    {{ errorMessage }}
                  </v-alert>
                </v-form>
              </v-card-text>
              
              <v-card-actions>
                <v-spacer></v-spacer>
                <v-btn
                  color="primary"
                  :loading="loading"
                  :disabled="!valid"
                  large
                  @click="handleLogin"
                >
                  Login
                  <v-icon right>mdi-login</v-icon>
                </v-btn>
              </v-card-actions>
              
              <v-card-text>
                <v-divider class="mb-4"></v-divider>
                <v-alert type="info" outlined dense>
                  <strong>Demo Credentials:</strong><br>
                  Username: admin<br>
                  Password: admin123
                </v-alert>
              </v-card-text>
            </v-card>
          </v-col>
        </v-row>
      </v-container>
    </v-main>
  </v-app>
</template>

<script>
import { useAuthStore } from '@/stores/auth'

export default {
  name: 'LoginView',
  
  data() {
    return {
      valid: false,
      loading: false,
      showPassword: false,
      showError: false,
      errorMessage: '',
      credentials: {
        username: '',
        password: ''
      },
      usernameRules: [
        v => !!v || 'Username is required',
        v => v.length >= 3 || 'Username must be at least 3 characters'
      ],
      passwordRules: [
        v => !!v || 'Password is required',
        v => v.length >= 6 || 'Password must be at least 6 characters'
      ]
    }
  },

  setup() {
    const authStore = useAuthStore()
    return { authStore }
  },

  async mounted() {
    // Check if already authenticated
    if (this.authStore.checkAuth()) {
      this.$router.push('/')
    }
  },

  methods: {
    async handleLogin() {
      if (!this.valid) return
      
      this.loading = true
      this.showError = false
      
      try {
        const result = await this.authStore.login(this.credentials)
        
        if (result.success) {
          this.$router.push('/')
        } else {
          this.errorMessage = result.error
          this.showError = true
        }
      } catch (error) {
        this.errorMessage = 'Login failed. Please try again.'
        this.showError = true
      } finally {
        this.loading = false
      }
    }
  }
}
</script>

<style scoped>
.fill-height {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.v-card {
  border-radius: 16px;
}
</style>
