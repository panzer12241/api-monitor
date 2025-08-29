<template>
  <v-container fluid>
    <!-- Stats Cards -->
    <v-row>
      <v-col cols="12" sm="6" md="3">
        <v-card color="primary" dark>
          <v-card-text>
            <div class="d-flex align-center">
              <v-icon size="40" class="mr-4">mdi-api</v-icon>
              <div>
                <div class="text-h4 font-weight-bold">{{ totalEndpoints }}</div>
                <div class="text-subtitle1">Total Endpoints</div>
              </div>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
      
      <v-col cols="12" sm="6" md="3">
        <v-card color="success" dark>
          <v-card-text>
            <div class="d-flex align-center">
              <v-icon size="40" class="mr-4">mdi-play-circle</v-icon>
              <div>
                <div class="text-h4 font-weight-bold">{{ activeEndpoints }}</div>
                <div class="text-subtitle1">Active Endpoints</div>
              </div>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
      
      <v-col cols="12" sm="6" md="3">
        <v-card color="info" dark>
          <v-card-text>
            <div class="d-flex align-center">
              <v-icon size="40" class="mr-4">mdi-check-circle</v-icon>
              <div>
                <div class="text-h4 font-weight-bold">{{ healthyEndpoints }}</div>
                <div class="text-subtitle1">Healthy Endpoints</div>
              </div>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
      
      <v-col cols="12" sm="6" md="3">
        <v-card color="error" dark>
          <v-card-text>
            <div class="d-flex align-center">
              <v-icon size="40" class="mr-4">mdi-alert-circle</v-icon>
              <div>
                <div class="text-h4 font-weight-bold">{{ unhealthyEndpoints }}</div>
                <div class="text-subtitle1">Unhealthy Endpoints</div>
              </div>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <!-- Charts Section -->
    <v-row class="mt-4">
      <v-col cols="12">
        <v-card>
          <v-card-title>
            <v-icon left>mdi-chart-line</v-icon>
            Response Time Trends
            
            <v-spacer></v-spacer>
            
            <v-btn 
              color="info" 
              @click="refreshChart" 
              size="small"
              :loading="chartLoading"
            >
              <v-icon left>mdi-refresh</v-icon>
              Refresh Chart
            </v-btn>
          </v-card-title>
          <v-card-text>
            <ResponseTimeChart 
              ref="responseTimeChart" 
              :endpoint-id="selectedChartEndpoint"
            />
          </v-card-text>
          <v-card-actions>
            <v-select
              v-model="selectedChartEndpoint"
              :items="endpointOptions"
              item-title="name"
              item-value="id"
              label="Select Endpoint"
              dense
              outlined
              class="mx-3"
            ></v-select>
          </v-card-actions>
        </v-card>
      </v-col>
    </v-row>

    <!-- Endpoints Management -->
    <v-row class="mt-4">
      <v-col cols="12">
        <v-card>
          <v-card-title>
            <v-icon left>mdi-api</v-icon>
            API Endpoints Management
            
            <v-spacer></v-spacer>
            
            <v-btn color="primary" @click="showCreateDialog" class="mr-2">
              <v-icon left>mdi-plus</v-icon>
              Add Endpoint
            </v-btn>
            
            <v-btn color="info" @click="refreshData" :loading="loading">
              <v-icon left>mdi-refresh</v-icon>
              Refresh
            </v-btn>
          </v-card-title>
          
          <v-data-table
            :headers="endpointHeaders"
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
            
            <template v-slot:item.proxy="{ item }">
              <v-chip
                v-if="item.proxy"
                :color="item.proxy.is_active ? 'success' : 'warning'"
                size="small"
                variant="outlined"
              >
                <v-icon left size="small">mdi-server-network</v-icon>
                {{ item.proxy.name }}
              </v-chip>
              <span v-else class="text-grey">No Proxy</span>
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
              v-model="endpointForm.name"
              label="Name"
              :rules="nameRules"
              required
              outlined
              placeholder="My API Endpoint"
            ></v-text-field>
            
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
            
            <v-select
              v-model="endpointForm.proxy_id"
              :items="proxyOptions"
              item-title="name"
              item-value="id"
              label="Proxy (Optional)"
              clearable
              outlined
              prepend-inner-icon="mdi-server-network"
            >
              <template v-slot:item="{ props, item }">
                <v-list-item v-bind="props" :disabled="!item.raw.is_active">
                  <v-list-item-title>{{ item.raw.name }}</v-list-item-title>
                  <v-list-item-subtitle>
                    {{ item.raw.host }}:{{ item.raw.port }}
                    <v-chip
                      v-if="!item.raw.is_active"
                      size="x-small"
                      color="error"
                      class="ml-2"
                    >
                      Inactive
                    </v-chip>
                  </v-list-item-subtitle>
                </v-list-item>
              </template>
            </v-select>
            
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
        
        <!-- Filters Section -->
        <v-card-text class="pb-0">
          <v-expansion-panels v-model="filtersExpanded">
            <v-expansion-panel>
              <v-expansion-panel-title>
                <v-icon left>mdi-filter</v-icon>
                Advanced Filters
              </v-expansion-panel-title>
              <v-expansion-panel-text>
                <v-row>
                  <v-col cols="12" md="3">
                    <v-text-field
                      v-model="filters.startDate"
                      label="Start Date"
                      type="datetime-local"
                      outlined
                      dense
                      @update:model-value="fetchLogs"
                    ></v-text-field>
                  </v-col>
                  
                  <v-col cols="12" md="3">
                    <v-text-field
                      v-model="filters.endDate"
                      label="End Date"
                      type="datetime-local"
                      outlined
                      dense
                      @update:model-value="fetchLogs"
                    ></v-text-field>
                  </v-col>
                  
                  <v-col cols="12" md="3">
                    <v-text-field
                      v-model.number="filters.minResponseTime"
                      label="Min Response Time (ms)"
                      type="number"
                      outlined
                      dense
                      placeholder="e.g. 1000"
                      @update:model-value="fetchLogs"
                    ></v-text-field>
                  </v-col>
                  
                  <v-col cols="12" md="3">
                    <v-select
                      v-model="filters.statusCode"
                      :items="statusCodeOptions"
                      label="Status Code"
                      outlined
                      dense
                      clearable
                      @update:model-value="fetchLogs"
                    ></v-select>
                  </v-col>
                </v-row>
                
                <v-row>
                  <v-col cols="auto">
                    <v-btn
                      color="secondary"
                      variant="outlined"
                      @click="clearFilters"
                      size="small"
                    >
                      <v-icon left>mdi-filter-remove</v-icon>
                      Clear Filters
                    </v-btn>
                  </v-col>
                </v-row>
              </v-expansion-panel-text>
            </v-expansion-panel>
          </v-expansion-panels>
        </v-card-text>
        
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
                        color="blue-grey-lighten-5"
                        class="pa-3 rounded"
                        style="max-height: 400px; overflow-y: auto;"
                    >
                        <div 
                        v-for="(value, key) in parseHeaders(log.response_headers)" 
                        :key="key"
                        class="text-body-2 mb-1 font-weight-medium"
                        style="font-family: 'Courier New', monospace; line-height: 1.5;"
                        >
                        <span class="font-weight-bold text-indigo-darken-2">{{ key }}:</span> 
                        <span class="text-grey-darken-4 font-weight-medium">{{ value }}</span>
                        </div>
                    </v-sheet>
                    <v-alert v-else type="info" variant="outlined">No headers available</v-alert>
                  </v-col>
                  
                  <v-col cols="12" md="6">
                    <h4 class="mb-3">Response Body</h4>
                    <v-sheet 
                      v-if="log.response_body && log.response_body.trim()" 
                      color="grey-darken-4"
                      class="pa-3 rounded"
                      style="max-height: 400px; overflow-y: auto;"
                    >
                      <pre class="text-green-lighten-2 text-body-2 font-weight-medium" style="font-family: 'Courier New', monospace; white-space: pre-wrap; word-break: break-word; line-height: 1.5; margin: 0;">{{ formatResponseBody(log.response_body) }}</pre>
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
        
        <!-- Pagination Controls -->
        <v-card-text v-if="logs.length > 0" class="pt-0">
          <v-row align="center" justify="space-between">
            <v-col cols="auto">
              <v-select
                v-model="logsPerPage"
                :items="[10, 25, 50, 100]"
                label="Items per page"
                density="compact"
                style="width: 120px;"
                @update:model-value="fetchLogs"
              ></v-select>
            </v-col>
            
            <v-col cols="auto">
              <span class="text-body-2">
                Showing {{ ((currentPage - 1) * logsPerPage) + 1 }} - 
                {{ Math.min(currentPage * logsPerPage, totalLogs) }} of {{ totalLogs }}
              </span>
            </v-col>
            
            <v-col cols="auto">
              <v-pagination
                v-model="currentPage"
                :length="totalPages"
                :total-visible="5"
                @update:model-value="fetchLogs"
              ></v-pagination>
            </v-col>
          </v-row>
        </v-card-text>
        
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn @click="logsDialog = false">Close</v-btn>
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
  </v-container>
</template>

<script>
import axios from 'axios'
import ResponseTimeChart from '@/components/ResponseTimeChart.vue'
import { wsClient } from '@/services/websocket.js'

const API_BASE = import.meta.env.DEV ? '/api/v1' : 'http://localhost:8080/api/v1'

export default {
  name: 'Dashboard',
  components: {
    ResponseTimeChart
  },
  
  data() {
    return {
      endpoints: [],
      proxies: [],
      logs: [],
      loading: false,
      logsLoading: false,
      chartLoading: false,
      
      // Pagination for logs
      currentPage: 1,
      logsPerPage: 25,
      totalLogs: 0,
      
      // Filters for logs
      filtersExpanded: [],
      filters: {
        startDate: '',
        endDate: '',
        minResponseTime: null,
        statusCode: null
      },
      
      // Dialog states
      dialog: false,
      deleteDialog: false,
      logsDialog: false,
      
      // Form states
      isEditMode: false,
      saving: false,
      deleting: false,
      valid: false,
      
      // Snackbar
      snackbar: false,
      snackbarText: '',
      snackbarColor: 'success',
      
      // Selected items
      deletingEndpoint: null,
      selectedEndpoint: null,
      
      // Form data
      endpointForm: {
        id: null,
        name: '',
        url: '',
        method: 'GET',
        headers: {},
        body: '',
        timeout_seconds: 30,
        check_interval_seconds: 60,
        is_active: true,
        proxy_id: null
      },
      
      // Validation rules
      nameRules: [
        v => !!v || 'Name is required',
        v => (v && v.length >= 3) || 'Name must be at least 3 characters',
        v => (v && v.length <= 100) || 'Name must be less than 100 characters'
      ],
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
      
      // Table headers
      endpointHeaders: [
        { title: 'Name', key: 'name', sortable: true },
        { title: 'URL', key: 'url', sortable: false },
        { title: 'Proxy', key: 'proxy', sortable: false },
        { title: 'Status', key: 'status', sortable: false },
        { title: 'Interval', key: 'check_interval_seconds', sortable: true },
        { title: 'Timeout', key: 'timeout_seconds', sortable: true },
        { title: 'Actions', key: 'actions', sortable: false }
      ],
      
      // Chart data
      selectedChartEndpoint: null
    }
  },
  
  computed: {
    totalEndpoints() {
      return this.endpoints.length
    },
    
    activeEndpoints() {
      return this.endpoints.filter(endpoint => endpoint.is_active).length
    },
    
    healthyEndpoints() {
      // For now, consider active endpoints as healthy
      // This could be enhanced with actual health status
      return this.activeEndpoints
    },
    
    unhealthyEndpoints() {
      return this.totalEndpoints - this.healthyEndpoints
    },
    
    endpointOptions() {
      return this.endpoints.map(endpoint => ({
        id: endpoint.id,
        name: endpoint.name || endpoint.url
      }))
    },
    
    proxyOptions() {
      return this.proxies.filter(proxy => proxy.is_active)
    },
    
    totalPages() {
      return Math.ceil(this.totalLogs / this.logsPerPage)
    },
    
    statusCodeOptions() {
      return [
        { title: 'Success (2xx)', value: '2xx' },
        { title: 'Redirect (3xx)', value: '3xx' },
        { title: 'Client Error (4xx)', value: '4xx' },
        { title: 'Server Error (5xx)', value: '5xx' },
        { title: '200 OK', value: '200' },
        { title: '404 Not Found', value: '404' },
        { title: '500 Internal Server Error', value: '500' }
      ]
    }
  },
  
  watch: {
    endpoints: {
      handler(newEndpoints) {
        if (newEndpoints && newEndpoints.length > 0) {
          // Only set if no endpoint is selected yet
          if (!this.selectedChartEndpoint) {
            this.selectedChartEndpoint = newEndpoints[0].id
          }
          // Ensure the selected endpoint exists in the current list
          else if (!newEndpoints.find(ep => ep.id === this.selectedChartEndpoint)) {
            this.selectedChartEndpoint = newEndpoints[0].id
          }
        }
      },
      immediate: true
    }
  },
  
  mounted() {
    this.refreshData()
    this.fetchProxies()
    
    // Setup WebSocket connection
    this.setupWebSocket()
  },
  
  beforeUnmount() {
    // Cleanup WebSocket connection
    this.cleanupWebSocket()
  },
  
  methods: {
    async refreshData() {
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
    
    async fetchProxies() {
      try {
        const response = await axios.get(`${API_BASE}/proxies`)
        this.proxies = response.data || []
      } catch (error) {
        console.error('Failed to fetch proxies:', error)
        this.proxies = []
      }
    },
    
    // Endpoint management methods
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
      
      const payload = {
        ...this.endpointForm,
        method: 'GET',
        headers: {},
        body: ''
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
        this.refreshData()
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
        this.refreshData()
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
        is_active: true,
        proxy_id: null
      }
    },
    
    // Logs methods
    viewLogs(endpoint) {
      this.selectedEndpoint = endpoint
      this.currentPage = 1  // Reset to first page
      this.clearFilters()   // Clear filters when opening new endpoint logs
      this.logsDialog = true
      this.fetchLogs()
    },
    
    async fetchLogs() {
      if (!this.selectedEndpoint) return
      
      this.logsLoading = true
      try {
        const offset = (this.currentPage - 1) * this.logsPerPage
        const params = new URLSearchParams({
          limit: this.logsPerPage.toString(),
          offset: offset.toString()
        })
        
        // Add filters if provided
        if (this.filters.startDate) {
          params.append('start_date', this.filters.startDate)
        }
        if (this.filters.endDate) {
          params.append('end_date', this.filters.endDate)
        }
        if (this.filters.minResponseTime) {
          params.append('min_response_time', this.filters.minResponseTime.toString())
        }
        if (this.filters.statusCode) {
          params.append('status_code', this.filters.statusCode)
        }
        
        const response = await axios.get(`${API_BASE}/endpoints/${this.selectedEndpoint.id}/logs?${params}`)
        
        if (response.data) {
          this.logs = response.data.logs || []
          this.totalLogs = response.data.total || 0
        } else {
          this.logs = []
          this.totalLogs = 0
        }
      } catch (error) {
        this.showSnackbar('Failed to fetch logs', 'error')
        console.error(error)
        this.logs = []
        this.totalLogs = 0
      } finally {
        this.logsLoading = false
      }
    },
    
    clearFilters() {
      this.filters = {
        startDate: '',
        endDate: '',
        minResponseTime: null,
        statusCode: null
      }
      this.currentPage = 1
      if (this.selectedEndpoint) {
        this.fetchLogs()
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
        const parsed = JSON.parse(body)
        return JSON.stringify(parsed, null, 2)
      } catch {
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
    },
    
    // Refresh chart manually
    refreshChart() {
      if (this.$refs.responseTimeChart) {
        this.chartLoading = true
        this.$refs.responseTimeChart.fetchLogs().finally(() => {
          this.chartLoading = false
        })
      }
    },
    
    // Update specific endpoint without full refresh
    updateEndpointStats(data) {
      // Find and update the specific endpoint
      const endpointIndex = this.endpoints.findIndex(ep => ep.id === data.endpoint_id)
      if (endpointIndex !== -1) {
        // Update endpoint with the latest check information
        const endpoint = this.endpoints[endpointIndex]
        
        // Update last check status (this can be used for health indicators)
        endpoint.last_check = {
          status_code: data.status_code,
          response_time: data.response_time,
          success: data.success,
          timestamp: data.timestamp
        }
        
        // Use Vue.set to ensure reactivity
        this.$set(this.endpoints, endpointIndex, { ...endpoint })
      }
    },
    
    // WebSocket methods
    setupWebSocket() {
      // Connect to WebSocket
      wsClient.connect()
      
      // Listen for endpoint check updates
      wsClient.on('endpoint_checked', (data) => {
        // Update specific endpoint stats without full refresh
        this.updateEndpointStats(data)
      })
      
      // Listen for endpoint created
      wsClient.on('endpoint_created', (data) => {
        console.log('New endpoint created:', data)
        // Add new endpoint to existing list instead of full refresh
        this.endpoints.push(data)
        this.showSnackbar(`New endpoint "${data.name}" created`, 'success')
      })
      
      // Listen for endpoint deleted
      wsClient.on('endpoint_deleted', (data) => {
        console.log('Endpoint deleted:', data)
        // Remove endpoint from list instead of full refresh
        this.endpoints = this.endpoints.filter(ep => ep.id !== data.id)
        this.showSnackbar('Endpoint deleted', 'info')
      })
    },
    
    cleanupWebSocket() {
      // Remove all listeners
      wsClient.off('endpoint_checked')
      wsClient.off('endpoint_created')
      wsClient.off('endpoint_deleted')
      // Disconnect WebSocket
      wsClient.disconnect()
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
