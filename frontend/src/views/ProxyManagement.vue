<template>
  <v-container fluid>
    <v-row>
      <v-col>
        <v-card>
          <v-card-title class="d-flex align-center">
            <v-icon left>mdi-server-network</v-icon>
            Proxy Management
            <v-spacer></v-spacer>
            <v-btn
              color="primary"
              @click="showCreateDialog"
              prepend-icon="mdi-plus"
            >
              Add Proxy
            </v-btn>
          </v-card-title>

          <v-card-text>
            <v-data-table
              :headers="headers"
              :items="proxies"
              :loading="loading"
              class="elevation-1"
              no-data-text="No proxies configured"
            >
              <template v-slot:item.is_active="{ item }">
                <v-chip
                  :color="item.is_active ? 'success' : 'error'"
                  size="small"
                  variant="flat"
                >
                  {{ item.is_active ? 'Active' : 'Inactive' }}
                </v-chip>
              </template>

              <template v-slot:item.endpoint="{ item }">
                {{ item.host }}:{{ item.port }}
              </template>

              <template v-slot:item.auth="{ item }">
                <v-chip
                  v-if="item.username"
                  color="info"
                  size="small"
                  variant="outlined"
                >
                  {{ item.username }}
                </v-chip>
                <span v-else class="text-grey">No Auth</span>
              </template>

              <template v-slot:item.actions="{ item }">
                <v-btn
                  icon="mdi-pencil"
                  size="small"
                  variant="text"
                  @click="editProxy(item)"
                ></v-btn>
                <v-btn
                  :icon="item.is_active ? 'mdi-pause' : 'mdi-play'"
                  size="small"
                  variant="text"
                  @click="toggleProxy(item)"
                ></v-btn>
                <v-btn
                  icon="mdi-delete"
                  size="small"
                  variant="text"
                  color="error"
                  @click="deleteProxy(item)"
                ></v-btn>
              </template>
            </v-data-table>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <!-- Create/Edit Dialog -->
    <v-dialog v-model="dialog" max-width="600px" persistent>
      <v-card>
        <v-card-title>
          <span class="text-h5">{{ isEditMode ? 'Edit' : 'Create' }} Proxy</span>
        </v-card-title>

        <v-card-text>
          <v-container>
            <v-row>
              <v-col cols="12">
                <v-text-field
                  v-model="form.name"
                  label="Proxy Name*"
                  required
                  :rules="[v => !!v || 'Name is required']"
                ></v-text-field>
              </v-col>

              <v-col cols="12" md="8">
                <v-text-field
                  v-model="form.host"
                  label="Host*"
                  required
                  :rules="[v => !!v || 'Host is required']"
                  placeholder="127.0.0.1 or proxy.example.com"
                ></v-text-field>
              </v-col>

              <v-col cols="12" md="4">
                <v-text-field
                  v-model.number="form.port"
                  label="Port*"
                  required
                  type="number"
                  :rules="[
                    v => !!v || 'Port is required',
                    v => (v >= 1 && v <= 65535) || 'Port must be between 1-65535'
                  ]"
                  placeholder="8080"
                ></v-text-field>
              </v-col>

              <v-col cols="12" md="6">
                <v-text-field
                  v-model="form.username"
                  label="Username"
                  placeholder="Optional"
                ></v-text-field>
              </v-col>

              <v-col cols="12" md="6">
                <v-text-field
                  v-model="form.password"
                  label="Password"
                  type="password"
                  placeholder="Optional"
                ></v-text-field>
              </v-col>

              <v-col cols="12">
                <v-switch
                  v-model="form.is_active"
                  label="Active"
                  color="primary"
                ></v-switch>
              </v-col>
            </v-row>
          </v-container>
        </v-card-text>

        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="grey" variant="text" @click="closeDialog">
            Cancel
          </v-btn>
          <v-btn
            color="primary"
            variant="flat"
            @click="saveProxy"
            :loading="saving"
          >
            {{ isEditMode ? 'Update' : 'Create' }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Delete Confirmation Dialog -->
    <v-dialog v-model="deleteDialog" max-width="400px">
      <v-card>
        <v-card-title class="text-h5">Confirm Delete</v-card-title>
        <v-card-text>
          Are you sure you want to delete proxy "{{ selectedProxy?.name }}"?
          <v-alert v-if="selectedProxy?.usage_count > 0" type="warning" class="mt-3">
            This proxy is being used by {{ selectedProxy.usage_count }} endpoint(s).
          </v-alert>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="grey" variant="text" @click="deleteDialog = false">
            Cancel
          </v-btn>
          <v-btn
            color="error"
            variant="flat"
            @click="confirmDelete"
            :loading="deleting"
          >
            Delete
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Snackbar -->
    <v-snackbar
      v-model="snackbar"
      :color="snackbarColor"
      timeout="3000"
    >
      {{ snackbarText }}
    </v-snackbar>
  </v-container>
</template>

<script>
import axios from 'axios'

const API_BASE = import.meta.env.DEV ? '/api/v1' : 'http://localhost:8080/api/v1'

export default {
  name: 'ProxyManagement',
  
  data() {
    return {
      proxies: [],
      loading: false,
      saving: false,
      deleting: false,
      
      // Dialog states
      dialog: false,
      deleteDialog: false,
      isEditMode: false,
      
      // Form data
      form: {
        name: '',
        host: '',
        port: 8080,
        username: '',
        password: '',
        is_active: true
      },
      
      selectedProxy: null,
      
      // Snackbar
      snackbar: false,
      snackbarText: '',
      snackbarColor: 'success',
      
      headers: [
        { title: 'Name', value: 'name', sortable: true },
        { title: 'Endpoint', value: 'endpoint', sortable: false },
        { title: 'Authentication', value: 'auth', sortable: false },
        { title: 'Status', value: 'is_active', sortable: true },
        { title: 'Created', value: 'created_at', sortable: true },
        { title: 'Actions', value: 'actions', sortable: false, width: 120 }
      ]
    }
  },
  
  mounted() {
    this.fetchProxies()
  },
  
  methods: {
    async fetchProxies() {
      this.loading = true
      try {
        const response = await axios.get(`${API_BASE}/proxies`)
        this.proxies = response.data || []
      } catch (error) {
        this.showSnackbar('Failed to fetch proxies', 'error')
        console.error(error)
      } finally {
        this.loading = false
      }
    },
    
    showCreateDialog() {
      this.form = {
        name: '',
        host: '',
        port: 8080,
        username: '',
        password: '',
        is_active: true
      }
      this.isEditMode = false
      this.dialog = true
    },
    
    editProxy(proxy) {
      this.form = { ...proxy }
      this.isEditMode = true
      this.dialog = true
    },
    
    closeDialog() {
      this.dialog = false
      this.form = {
        name: '',
        host: '',
        port: 8080,
        username: '',
        password: '',
        is_active: true
      }
    },
    
    async saveProxy() {
      this.saving = true
      try {
        if (this.isEditMode) {
          await axios.put(`${API_BASE}/proxies/${this.form.id}`, this.form)
          this.showSnackbar('Proxy updated successfully')
        } else {
          await axios.post(`${API_BASE}/proxies`, this.form)
          this.showSnackbar('Proxy created successfully')
        }
        
        this.closeDialog()
        this.fetchProxies()
      } catch (error) {
        this.showSnackbar(error.response?.data?.error || 'Failed to save proxy', 'error')
        console.error(error)
      } finally {
        this.saving = false
      }
    },
    
    deleteProxy(proxy) {
      this.selectedProxy = proxy
      this.deleteDialog = true
    },
    
    async confirmDelete() {
      this.deleting = true
      try {
        await axios.delete(`${API_BASE}/proxies/${this.selectedProxy.id}`)
        this.showSnackbar('Proxy deleted successfully')
        this.deleteDialog = false
        this.fetchProxies()
      } catch (error) {
        this.showSnackbar(error.response?.data?.error || 'Failed to delete proxy', 'error')
        console.error(error)
      } finally {
        this.deleting = false
      }
    },
    
    async toggleProxy(proxy) {
      try {
        await axios.post(`${API_BASE}/proxies/${proxy.id}/toggle`)
        this.showSnackbar(`Proxy ${proxy.is_active ? 'deactivated' : 'activated'}`)
        this.fetchProxies()
      } catch (error) {
        this.showSnackbar('Failed to toggle proxy status', 'error')
        console.error(error)
      }
    },
    
    showSnackbar(text, color = 'success') {
      this.snackbarText = text
      this.snackbarColor = color
      this.snackbar = true
    }
  }
}
</script>

<style scoped>
.v-data-table {
  margin-top: 16px;
}
</style>
