<template>
  <v-container fluid>
    <v-row>
      <v-col cols="12">
        <v-card>
          <v-card-title>
            <v-icon left>mdi-api</v-icon>
            API Endpoints Management
            
            <v-spacer></v-spacer>
            
            <v-btn color="primary" @click="showCreateDialog">
              <v-icon left>mdi-plus</v-icon>
              Add Endpoint
            </v-btn>
          </v-card-title>
          
          <v-data-table
            :headers="headers"
            :items="endpoints"
            :loading="loading"
            class="elevation-1"
          >
            <template v-slot:item.status="{ item }">
              <v-switch
                v-model="item.is_active"
                @update:model-value="toggleEndpoint(item)"
                color="success"
                hide-details
              ></v-switch>
            </template>
            
            <template v-slot:item.check_interval_seconds="{ item }">
              {{ item.check_interval_seconds }}s
            </template>
            
            <template v-slot:item.timeout_seconds="{ item }">
              {{ item.timeout_seconds }}s
            </template>
            
            <template v-slot:item.actions="{ item }">
              <v-btn
                size="small"
                color="info"
                @click="viewLogs(item)"
                class="mr-2"
              >
                <v-icon left small>mdi-history</v-icon>
                Logs
              </v-btn>
              
              <v-btn
                size="small"
                color="primary"
                @click="editEndpoint(item)"
                class="mr-2"
              >
                <v-icon left small>mdi-pencil</v-icon>
                Edit
              </v-btn>
              
              <v-btn
                size="small"
                color="error"
                @click="deleteEndpoint(item)"
              >
                <v-icon left small>mdi-delete</v-icon>
                Delete
              </v-btn>
            </template>
          </v-data-table>
        </v-card>
      </v-col>
    </v-row>

    <!-- Create/Edit Dialog -->
    <v-dialog v-model="dialog" max-width="800px" persistent>
      <v-card>
        <v-card-title>
          <span class="text-h5">
            <v-icon left>{{ isEditMode ? 'mdi-pencil' : 'mdi-plus' }}</v-icon>
            {{ isEditMode ? 'Edit Endpoint' : 'Create Endpoint' }}
          </span>
        </v-card-title>
        
        <v-card-text>
          <v-form ref="endpointForm" v-model="valid">
            <v-text-field
              v-model="endpointForm.url"
              label="URL"
              :rules="urlRules"
              required
              outlined
              placeholder="https://api.example.com/health"
            ></v-text-field>
            
            <v-row>
              <v-col cols="6">
                <v-text-field
                  v-model.number="endpointForm.check_interval_seconds"
                  label="Check Interval (seconds)"
                  type="number"
                  :min="10"
                  :max="3600"
                  outlined
                ></v-text-field>
              </v-col>
              
              <v-col cols="6">
                <v-text-field
                  v-model.number="endpointForm.timeout_seconds"
                  label="Timeout (seconds)"
                  type="number"
                  :min="5"
                  :max="300"
                  outlined
                ></v-text-field>
              </v-col>
            </v-row>
            
            <v-switch
              v-model="endpointForm.is_active"
              label="Active"
              color="success"
            ></v-switch>
          </v-form>
        </v-card-text>
        
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn text @click="dialog = false">Cancel</v-btn>
          <v-btn 
            color="primary" 
            @click="saveEndpoint"
            :disabled="!valid"
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
        <v-card-title>
          <v-icon left color="error">mdi-delete</v-icon>
          Confirm Delete
        </v-card-title>
        
        <v-card-text>
          Are you sure you want to delete the endpoint "{{ deletingEndpoint?.name }}"?
          This action cannot be undone.
        </v-card-text>
        
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn text @click="deleteDialog = false">Cancel</v-btn>
          <v-btn 
            color="error" 
            @click="confirmDelete"
            :loading="deleting"
          >
            Delete
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Snackbar for notifications -->
    <v-snackbar
      v-model="snackbar"
      :color="snackbarColor"
      :timeout="3000"
      top
    >
      {{ snackbarText }}
      <template v-slot:actions>
        <v-btn color="white" variant="text" @click="snackbar = false">
          Close
        </v-btn>
      </template>
    </v-snackbar>

    <!-- Logs Dialog -->
    <v-dialog v-model="logsDialog" max-width="1200px">
      <v-card>
        <v-card-title>
          <v-icon left>mdi-history</v-icon>
          Endpoint Logs: {{ selectedEndpoint?.name || selectedEndpoint?.url }}
          <v-spacer></v-spacer>
          <v-btn 
            color="primary" 
            @click="fetchLogs" 
            :loading="logsLoading"
            size="small"
          >
            <v-icon left>mdi-refresh</v-icon>
            Refresh
          </v-btn>
        </v-card-title>
        
        <v-card-text style="max-height: 600px; overflow-y: auto;">
          <v-progress-linear v-if="logsLoading" indeterminate></v-progress-linear>
          
          <v-expansion-panels v-if="logs.length > 0" multiple>
            <v-expansion-panel v-for="log in logs" :key="log.id">
              <v-expansion-panel-title>
                <div class="d-flex align-center">
                  <v-chip 
                    :color="log.status_code >= 200 && log.status_code < 300 ? 'success' : 'error'"
                    size="small"
                    class="mr-3"
                  >
                    {{ log.status_code }}
                  </v-chip>
                  
                  <span class="mr-3">{{ log.response_time_ms }}ms</span>
                  
                  <v-chip 
                    variant="outlined" 
                    size="small"
                    class="mr-3"
                  >
                    {{ formatDate(log.checked_at) }}
                  </v-chip>
                  
                  <span v-if="log.error_message" class="text-error">
                    {{ log.error_message }}
                  </span>
                </div>
              </v-expansion-panel-title>
              
              <v-expansion-panel-text>
                <v-row>
                  <v-col cols="12" md="6">
                    <h4 class="mb-3">Response Headers</h4>
                    <v-sheet 
                      v-if="log.response_headers && log.response_headers.trim()" 
                      color="grey-lighten-4"
                      class="pa-3 rounded"
                      style="max-height: 400px; overflow-y: auto;"
                    >
                      <div 
                        v-for="(value, key) in parseHeaders(log.response_headers)" 
                        :key="key"
                        class="text-body-2 mb-1"
                        style="font-family: 'Courier New', monospace; line-height: 1.5;"
                      >
                        <span class="font-weight-bold text-blue">{{ key }}:</span> 
                        <span class="text-grey-darken-3">{{ value }}</span>
                      </div>
                    </v-sheet>
                    <v-alert v-else type="info" variant="outlined">No headers available</v-alert>
                  </v-col>
                  
                  <v-col cols="12" md="6">
                    <h4 class="mb-3">Response Body</h4>
                    <v-sheet 
                      v-if="log.response_body && log.response_body.trim()" 
                      color="black"
                      class="pa-3 rounded"
                      style="max-height: 400px; overflow-y: auto;"
                    >
                      <pre class="text-white text-body-2" style="font-family: 'Courier New', monospace; white-space: pre-wrap; word-break: break-word; line-height: 1.5; margin: 0;">{{ formatResponseBody(log.response_body) }}</pre>
                    </v-sheet>
                    <v-alert v-else type="info" variant="outlined">No response body</v-alert>
                  </v-col>
                </v-row>
                
                <!-- Debug info -->
                <v-row class="mt-4">
                  <v-col cols="12">
                    <details>
                      <summary>Debug Info</summary>
                      <pre class="text-caption">{{ JSON.stringify(log, null, 2) }}</pre>
                    </details>
                  </v-col>
                </v-row>
              </v-expansion-panel-text>
            </v-expansion-panel>
          </v-expansion-panels>
          
          <v-alert v-else-if="!logsLoading" type="info">
            No logs available for this endpoint
          </v-alert>
        </v-card-text>
        
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn @click="logsDialog = false">Close</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script>
import axios from 'axios'

const API_BASE = import.meta.env.DEV ? '/api/v1' : 'http://localhost:8080/api/v1'

export default {
  name: 'EndpointsView',
  
  data() {
    return {
      endpoints: [],
      dialog: false,
      deleteDialog: false,
      isEditMode: false,
      loading: false,
      saving: false,
      deleting: false,
      valid: false,
      snackbar: false,
      snackbarText: '',
      snackbarColor: 'success',
      deletingEndpoint: null,
      
      // Logs related
      logsDialog: false,
      logs: [],
      logsLoading: false,
      selectedEndpoint: null,
      
      endpointForm: {
        id: null,
        name: '',
        url: '',
        method: 'GET',
        headers: {},
        body: '',
        timeout_seconds: 30,
        check_interval_seconds: 60,
        is_active: true
      },
      
      urlRules: [
        v => !!v || 'URL is required',
        v => {
          try {
            new URL(v)
            return true
          } catch {
            return 'Please enter a valid URL'
          }
        }
      ],
      
      tableHeaders: [
        { title: 'URL', key: 'url', sortable: false },
        { title: 'Status', key: 'status', sortable: false },
        { title: 'Interval', key: 'check_interval_seconds', sortable: true },
        { title: 'Timeout', key: 'timeout_seconds', sortable: true },
        { title: 'Actions', key: 'actions', sortable: false }
      ]
    }
  },
  
  computed: {
    headers() {
      return this.tableHeaders
    }
  },
  
  mounted() {
    this.fetchEndpoints()
  },
  
  methods: {
    async fetchEndpoints() {
      this.loading = true
      try {
        const response = await axios.get(`${API_BASE}/endpoints`)
        this.endpoints = response.data || []
      } catch (error) {
        this.showSnackbar('Failed to fetch endpoints', 'error')
        console.error(error)
      } finally {
        this.loading = false
      }
    },
    
    showCreateDialog() {
      this.resetForm()
      this.isEditMode = false
      this.dialog = true
    },
    
    editEndpoint(endpoint) {
      this.endpointForm = { ...endpoint }
      this.isEditMode = true
      this.dialog = true
    },
    
    async saveEndpoint() {
      if (!this.valid) return
      
      this.saving = true
      
      // Set default values for simplified form
      const payload = {
        ...this.endpointForm,
        name: this.endpointForm.url, // Auto-generate name from URL
        method: 'GET', // Default to GET
        headers: {}, // Empty headers
        body: '' // Empty body
      }
      
      try {
        if (this.isEditMode) {
          await axios.put(`${API_BASE}/endpoints/${this.endpointForm.id}`, payload)
          this.showSnackbar('Endpoint updated successfully', 'success')
        } else {
          await axios.post(`${API_BASE}/endpoints`, payload)
          this.showSnackbar('Endpoint created successfully', 'success')
        }
        
        this.dialog = false
        this.fetchEndpoints()
      } catch (error) {
        this.showSnackbar('Failed to save endpoint', 'error')
        console.error(error)
      } finally {
        this.saving = false
      }
    },
    
    async toggleEndpoint(endpoint) {
      try {
        await axios.post(`${API_BASE}/endpoints/${endpoint.id}/toggle`)
        this.showSnackbar(`Endpoint ${endpoint.is_active ? 'activated' : 'deactivated'}`, 'success')
      } catch (error) {
        this.showSnackbar('Failed to toggle endpoint', 'error')
        // Revert the switch
        endpoint.is_active = !endpoint.is_active
        console.error(error)
      }
    },
    
    deleteEndpoint(endpoint) {
      this.deletingEndpoint = endpoint
      this.deleteDialog = true
    },
    
    async confirmDelete() {
      this.deleting = true
      
      try {
        await axios.delete(`${API_BASE}/endpoints/${this.deletingEndpoint.id}`)
        this.showSnackbar('Endpoint deleted successfully', 'success')
        this.deleteDialog = false
        this.fetchEndpoints()
      } catch (error) {
        this.showSnackbar('Failed to delete endpoint', 'error')
        console.error(error)
      } finally {
        this.deleting = false
      }
    },
    
    resetForm() {
      this.endpointForm = {
        id: null,
        name: '',
        url: '',
        method: 'GET',
        headers: {},
        body: '',
        timeout_seconds: 30,
        check_interval_seconds: 60,
        is_active: true
      }
    },
    
    // Logs methods
    viewLogs(endpoint) {
      this.selectedEndpoint = endpoint
      this.logsDialog = true
      this.fetchLogs()
    },
    
    async fetchLogs() {
      if (!this.selectedEndpoint) return
      
      this.logsLoading = true
      try {
        const response = await axios.get(`${API_BASE}/endpoints/${this.selectedEndpoint.id}/logs?limit=20`)
        this.logs = response.data || []
      } catch (error) {
        this.showSnackbar('Failed to fetch logs', 'error')
        console.error(error)
      } finally {
        this.logsLoading = false
      }
    },
    
    parseHeaders(headersStr) {
      try {
        if (!headersStr || headersStr.trim() === '') return {}
        return JSON.parse(headersStr)
      } catch (e) {
        console.error('Error parsing headers:', e, 'Headers string:', headersStr)
        return {}
      }
    },
    
    formatResponseBody(body) {
      try {
        // Try to format as JSON if possible
        const parsed = JSON.parse(body)
        return JSON.stringify(parsed, null, 2)
      } catch {
        // Return as-is if not JSON
        return body
      }
    },
    
    formatDate(dateStr) {
      return new Date(dateStr).toLocaleString()
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
/* Only keep essential overrides */
.log-expansion-panel >>> .v-expansion-panel-text__wrapper {
  padding: 16px;
}
</style>
