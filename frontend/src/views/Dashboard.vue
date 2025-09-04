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

    <!-- Modal Components -->
    <EndpointFormModal
      v-model:dialog="dialog"
      :is-edit-mode="isEditMode"
      :endpoint-data="endpointForm"
      :proxy-options="proxyOptions"
      :saving="saving"
      @save="handleSaveEndpoint"
      @close="handleCloseFormModal"
    />

    <DeleteConfirmationModal
      v-model:dialog="deleteDialog"
      :endpoint="deletingEndpoint"
      :deleting="deleting"
      @confirm="handleConfirmDelete"
      @close="handleCloseDeleteModal"
    />

    <EndpointLogsModal
      v-model:dialog="logsDialog"
      :endpoint="selectedEndpoint"
      :logs="logs"
      :logs-loading="logsLoading"
      :loading-more="loadingMore"
      :total-logs="totalLogs"
      :logs-per-page="logsPerPage"
      :filters="filters"
      @refresh="resetAndFetchLogs"
      @filter-change="handleFilterChange"
      @load-more="loadMoreLogs"
      @logs-per-page-change="handleLogsPerPageChange"
      @clear-filters="clearFilters"
      @close="handleCloseLogsModal"
    />

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
import EndpointFormModal from '@/components/EndpointFormModal.vue'
import DeleteConfirmationModal from '@/components/DeleteConfirmationModal.vue'
import EndpointLogsModal from '@/components/EndpointLogsModal.vue'

export default {
  name: 'Dashboard',
  components: {
    ResponseTimeChart,
    EndpointFormModal,
    DeleteConfirmationModal,
    EndpointLogsModal
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

    // Modal event handlers
    async handleSaveEndpoint(payload) {
      this.saving = true
      
      try {
        if (this.isEditMode) {
          await endpointsAPI.update(payload.id, payload)
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

    handleCloseFormModal() {
      this.dialog = false
      this.resetForm()
    },

    async handleConfirmDelete(endpoint) {
      this.deleting = true
      
      try {
        await endpointsAPI.delete(endpoint.id)
        this.showSnackbar('Endpoint deleted successfully', 'success')
        this.deleteDialog = false
        this.refreshData()
      } catch (error) {
        this.showSnackbar('Failed to delete endpoint', 'error')
      } finally {
        this.deleting = false
      }
    },

    handleCloseDeleteModal() {
      this.deleteDialog = false
      this.deletingEndpoint = null
    },

    handleFilterChange(filters) {
      this.filters = { ...filters }
      this.resetAndFetchLogs()
    },

    handleLogsPerPageChange(newPerPage) {
      this.logsPerPage = newPerPage
      this.resetAndFetchLogs()
    },

    handleCloseLogsModal() {
      this.logsDialog = false
      this.selectedEndpoint = null
      this.logs = []
    },

    // Legacy methods (keeping for backward compatibility)
    async saveEndpoint() {
      // This is now handled by handleSaveEndpoint
      console.warn('saveEndpoint is deprecated, use handleSaveEndpoint instead')
    },

    async confirmDelete() {
      // This is now handled by handleConfirmDelete  
      console.warn('confirmDelete is deprecated, use handleConfirmDelete instead')
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
/* Dashboard specific styles */
</style>
