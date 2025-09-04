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
            :items="endpointsData"
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
    <v-dialog v-model="logsDialog" max-width="1400px" fullscreen-mobile>
      <v-card class="logs-dialog-card">
        <v-card-title class="pa-6 pb-4 position-relative">
          <v-icon left>mdi-history</v-icon>
          Endpoint Logs: {{ selectedEndpoint?.name || selectedEndpoint?.url }}
          <v-spacer></v-spacer>
          <v-btn 
            color="primary" 
            @click="resetAndFetchLogs" 
            :loading="logsLoading"
            size="small"
            class="mr-2"
          >
            <v-icon left>mdi-refresh</v-icon>
            Refresh
          </v-btn>
          
          <!-- Close button in top right corner -->
          <v-btn
            icon
            @click="logsDialog = false"
            class="close-btn"
            size="large"
            color="amber"
            variant="elevated"
          >
            <v-icon size="24" color="black">mdi-close</v-icon>
          </v-btn>
        </v-card-title>
        
        <!-- Filters Section -->
        <v-card-text class="pa-6 pt-0">
          <v-expansion-panels v-model="filtersExpanded" class="mb-4">
            <v-expansion-panel>
              <v-expansion-panel-title class="pa-4">
                <v-icon left>mdi-filter</v-icon>
                Advanced Filters
              </v-expansion-panel-title>
              <v-expansion-panel-text class="pa-4">
                <v-row class="mb-4">
                  <v-col cols="12" md="3">
                    <v-text-field
                      v-model="filters.startDate"
                      label="Start Date"
                      type="datetime-local"
                      outlined
                      dense
                      @update:model-value="resetAndFetchLogs"
                    ></v-text-field>
                  </v-col>
                  
                  <v-col cols="12" md="3">
                    <v-text-field
                      v-model="filters.endDate"
                      label="End Date"
                      type="datetime-local"
                      outlined
                      dense
                      @update:model-value="resetAndFetchLogs"
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
                      @update:model-value="resetAndFetchLogs"
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
                      @update:model-value="resetAndFetchLogs"
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
        
        <v-card-text class="pa-6 pt-0" style="max-height: 70vh; overflow-y: auto;">
          <v-progress-linear v-if="logsLoading" indeterminate class="mb-4"></v-progress-linear>
          
          <v-expansion-panels v-if="logs.length > 0" multiple class="logs-expansion-panels">
            <v-expansion-panel v-for="log in logs" :key="log.id" class="mb-3">
              <v-expansion-panel-title class="pa-4">
                <div class="d-flex align-center w-100">
                  <v-chip 
                    :color="log.status_code >= 200 && log.status_code < 300 ? 'success' : 'error'"
                    size="small"
                    class="mr-4"
                  >
                    {{ log.status_code }}
                  </v-chip>
                  
                  <span class="mr-4 font-weight-medium">{{ log.response_time_ms }}ms</span>
                  
                  <v-chip 
                    variant="outlined" 
                    size="small"
                    class="mr-4"
                  >
                    {{ formatDate(log.checked_at) }}
                  </v-chip>
                  
                  <span v-if="log.error_message" class="text-error text-truncate flex-grow-1">
                    {{ log.error_message }}
                  </span>
                </div>
              </v-expansion-panel-title>
              
              <v-expansion-panel-text class="pa-6">
                <v-row class="mb-4">
                  <v-col cols="12" md="6">
                    <h4 class="mb-4">Response Headers</h4>
                    <v-sheet 
                        v-if="log.response_headers && log.response_headers.trim()" 
                        color="blue-grey-lighten-5"
                        class="pa-4 rounded"
                        style="max-height: 400px; overflow-y: auto;"
                    >
                        <div 
                        v-for="(value, key) in parseHeaders(log.response_headers)" 
                        :key="key"
                        class="text-body-2 mb-2 font-weight-medium"
                        style="font-family: 'Courier New', monospace; line-height: 1.6;"
                        >
                        <span class="font-weight-bold text-indigo-darken-2">{{ key }}:</span> 
                        <span class="text-grey-darken-4 font-weight-medium ml-2">{{ value }}</span>
                        </div>
                    </v-sheet>
                    <v-alert v-else type="info" variant="outlined">No headers available</v-alert>
                  </v-col>
                  
                  <v-col cols="12" md="6">
                    <h4 class="mb-4">Response Body</h4>
                    <v-sheet 
                      v-if="log.response_body && log.response_body.trim()" 
                      color="grey-darken-4"
                      class="pa-4 rounded"
                      style="max-height: 400px; overflow-y: auto;"
                    >
                      <pre class="text-green-lighten-2 text-body-2 font-weight-medium" style="font-family: 'Courier New', monospace; white-space: pre-wrap; word-break: break-word; line-height: 1.6; margin: 0;">{{ formatResponseBody(log.response_body) }}</pre>
                    </v-sheet>
                    <v-alert v-else type="info" variant="outlined">No response body</v-alert>
                  </v-col>
                </v-row>
                
                <!-- Debug info -->
                <v-row class="mt-6">
                  <v-col cols="12">
                    <details>
                      <summary class="mb-2 text-subtitle-2 cursor-pointer">Debug Info</summary>
                      <v-sheet color="grey-lighten-4" class="pa-3 rounded">
                        <pre class="text-caption">{{ JSON.stringify(log, null, 2) }}</pre>
                      </v-sheet>
                    </details>
                  </v-col>
                </v-row>
              </v-expansion-panel-text>
            </v-expansion-panel>
          </v-expansion-panels>
          
          <v-alert v-else-if="!logsLoading" type="info" class="ma-4">
            No logs available for this endpoint
          </v-alert>
        </v-card-text>
        
        <!-- Pagination Controls -->
        <v-card-text v-if="logs.length > 0" class="pa-6 pt-4">
          <v-row align="center" justify="space-between">
            <v-col cols="auto">
              <v-select
                v-model="logsPerPage"
                :items="[10, 25, 50, 100]"
                label="Items per page"
                density="compact"
                style="width: 140px;"
                @update:model-value="resetAndFetchLogs"
              ></v-select>
            </v-col>
            
            <v-col cols="auto">
              <span class="text-body-2">
                Showing {{ logs.length }} of {{ totalLogs }} logs
              </span>
            </v-col>
            
            <v-col cols="auto">
              <v-btn
                v-if="hasMoreLogs"
                color="primary"
                @click="loadMoreLogs"
                :loading="loadingMore"
                size="large"
              >
                <v-icon left>mdi-plus</v-icon>
                Load More
              </v-btn>
              <span v-else class="text-body-2 text-grey">
                All logs loaded
              </span>
            </v-col>
          </v-row>
        </v-card-text>
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
import { endpointsAPI, proxiesAPI } from '@/services/api'
import ResponseTimeChart from '@/components/ResponseTimeChart.vue'

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
      loadingMore: false,
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
    endpointsData() {
      return Array.isArray(this.endpoints) ? this.endpoints : []
    },
    
    proxiesData() {
      return Array.isArray(this.proxies) ? this.proxies : []
    },
    
    totalEndpoints() {
      return this.endpointsData.length
    },
    
    activeEndpoints() {
      return this.endpointsData.filter(endpoint => endpoint.is_active).length
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
      return this.endpointsData.map(endpoint => ({
        id: endpoint.id,
        name: endpoint.name || endpoint.url
      }))
    },
    
    proxyOptions() {
      return this.proxiesData.filter(proxy => proxy.is_active)
    },
    
    totalPages() {
      return Math.ceil(this.totalLogs / this.logsPerPage)
    },
    
    hasMoreLogs() {
      return this.logs.length < this.totalLogs
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
    
  },
  
  beforeUnmount() {
  },
  
  methods: {
    async refreshData() {
      this.loading = true
      try {
        const response = await endpointsAPI.getAll()
        
        // Handle Axios response structure
        if (response && response.data) {
          if (response.data.data && Array.isArray(response.data.data)) {
            this.endpoints = response.data.data
          } else if (Array.isArray(response.data)) {
            this.endpoints = response.data
          } else {
            this.endpoints = []
          }
        } else {
          this.endpoints = []
        }
      } catch (error) {
        this.showSnackbar('Failed to fetch endpoints', 'error')
        this.endpoints = []
      } finally {
        this.loading = false
      }
    },
    
    async fetchProxies() {
      try {
        const response = await proxiesAPI.getAll()
        
        // Handle Axios response structure
        if (response && response.data) {
          if (response.data.data && Array.isArray(response.data.data)) {
            this.proxies = response.data.data
          } else if (Array.isArray(response.data)) {
            this.proxies = response.data
          } else {
            this.proxies = []
          }
        } else {
          this.proxies = []
        }
      } catch (error) {
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
        method: this.endpointForm.method || 'GET',
        headers: this.endpointForm.headers || {},
        body: this.endpointForm.body || ''
      }
      
      try {
        if (this.isEditMode) {
          await endpointsAPI.update(this.endpointForm.id, payload)
          this.showSnackbar('Endpoint updated successfully', 'success')
        } else {
          await endpointsAPI.create(payload)
          this.showSnackbar('Endpoint created successfully', 'success')
        }
        
        this.dialog = false
        this.refreshData()
      } catch (error) {
        this.showSnackbar('Failed to save endpoint', 'error')
      } finally {
        this.saving = false
      }
    },
    
    async toggleEndpoint(endpoint) {
      try {
        const response = await endpointsAPI.toggle(endpoint.id)
        
        // Use response data to update state
        if (response && response.data && response.data.data) {
          const newState = response.data.data.is_active
          endpoint.is_active = newState
          this.showSnackbar(`Endpoint ${newState ? 'activated' : 'deactivated'}`, 'success')
        } else {
          // Fallback: refresh data to get current state
          this.refreshData()
          this.showSnackbar('Endpoint status updated', 'success')
        }
      } catch (error) {
        // Revert switch state on error
        endpoint.is_active = !endpoint.is_active
        this.showSnackbar('Failed to toggle endpoint', 'error')
      }
    },
    
    deleteEndpoint(endpoint) {
      this.deletingEndpoint = endpoint
      this.deleteDialog = true
    },
    
    async confirmDelete() {
      this.deleting = true
      
      try {
        await endpointsAPI.delete(this.deletingEndpoint.id)
        this.showSnackbar('Endpoint deleted successfully', 'success')
        this.deleteDialog = false
        this.refreshData()
      } catch (error) {
        this.showSnackbar('Failed to delete endpoint', 'error')
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
    
    async fetchLogs(append = false) {
      console.log('fetchLogs called with append:', append, 'selectedEndpoint:', this.selectedEndpoint?.id)
      if (!this.selectedEndpoint) return
      
      if (!append) {
        this.logsLoading = true
      } else {
        this.loadingMore = true
      }
      
      try {
        const offset = append ? this.logs.length : 0
        const params = {
          limit: this.logsPerPage,
          offset: offset
        }
        
        // Add filters if provided
        if (this.filters.startDate) {
          params.start_date = this.filters.startDate
        }
        if (this.filters.endDate) {
          params.end_date = this.filters.endDate
        }
        if (this.filters.minResponseTime) {
          params.min_response_time = this.filters.minResponseTime
        }
        if (this.filters.statusCode) {
          params.status_code = this.filters.statusCode
        }
        
        const response = await endpointsAPI.getLogs(this.selectedEndpoint.id, params)
        
        if (response.data) {
          const newLogs = response.data.logs || []
          
          if (append) {
            this.logs = [...this.logs, ...newLogs]
          } else {
            this.logs = newLogs
          }
          
          this.totalLogs = response.data.total || 0
        } else {
          if (!append) {
            this.logs = []
            this.totalLogs = 0
          }
        }
      } catch (error) {
        this.showSnackbar('Failed to fetch logs: ' + (error.response?.data?.message || error.message), 'error')
        if (!append) {
          this.logs = []
          this.totalLogs = 0
        }
      } finally {
        this.logsLoading = false
        this.loadingMore = false
      }
    },
    
    async loadMoreLogs() {
      await this.fetchLogs(true)
    },
    
    async resetAndFetchLogs() {
      this.logs = []
      await this.fetchLogs(false)
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
        this.resetAndFetchLogs()
      }
    },
    
    parseHeaders(headersStr) {
      try {
        if (!headersStr || headersStr.trim() === '') return {}
        return JSON.parse(headersStr)
      } catch (e) {
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
    }
  }
}
</script>

<style scoped>
/* Only keep essential overrides */
.log-expansion-panel >>> .v-expansion-panel-text__wrapper {
  padding: 16px;
}

/* Logs dialog improvements */
.logs-dialog-card {
  border-radius: 12px !important;
  position: relative;
  overflow: visible !important;
}

/* Close button positioning */
.close-btn {
  position: absolute !important;
  top: 12px;
  right: 12px;
  z-index: 1000;
  border: 3px solid #FFB300 !important;
  background: linear-gradient(135deg, #FFD54F, #FF8F00) !important;
  box-shadow: 0 4px 12px rgba(255, 179, 0, 0.4) !important;
  border-radius: 50% !important;
  width: 48px !important;
  height: 48px !important;
}

.close-btn:hover {
  background: linear-gradient(135deg, #FFE082, #FFA000) !important;
  transform: scale(1.1);
  transition: all 0.2s ease;
  box-shadow: 0 6px 16px rgba(255, 179, 0, 0.6) !important;
}

.logs-expansion-panels .v-expansion-panel {
  border: 1px solid rgba(0, 0, 0, 0.12);
  border-radius: 8px !important;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.logs-expansion-panels .v-expansion-panel:not(:last-child) {
  margin-bottom: 16px;
}

.logs-expansion-panels .v-expansion-panel-title {
  border-radius: 8px 8px 0 0 !important;
}

.logs-expansion-panels .v-expansion-panel--active .v-expansion-panel-title {
  border-radius: 8px 8px 0 0 !important;
}

.cursor-pointer {
  cursor: pointer;
}

/* Responsive improvements */
@media (max-width: 960px) {
  .logs-dialog-card .v-card-text {
    max-height: 60vh !important;
  }
}
</style>
