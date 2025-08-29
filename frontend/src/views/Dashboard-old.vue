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
            
            <template v-slot:item.actions="{ item }">
              <v-btn
                size="small"
                color="primary"
                @click="manualCheck(item.id)"
                :loading="checkingEndpoints.includes(item.id)"
                class="mr-2"
              >
                <v-icon left small>mdi-play</v-icon>
                Check Now
              </v-btn>
              
              <v-btn
                size="small"
                color="info"
                @click="viewLogs(item)"
              >
                <v-icon left small>mdi-file-document</v-icon>
                View Logs
              </v-btn>
            </template>
          </v-data-table>
        </v-card>
      </v-col>
    </v-row>

    <!-- Logs Dialog -->
    <v-dialog v-model="logsDialog" max-width="1200px" scrollable>
      <v-card>
        <v-card-title>
          <v-icon left>mdi-file-document</v-icon>
          Logs for {{ selectedEndpoint?.name }}
          
          <v-spacer></v-spacer>
          
          <v-btn icon @click="logsDialog = false">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-card-title>
        
        <v-data-table
          :headers="logHeaders"
          :items="logs"
          :loading="logsLoading"
          class="elevation-1"
          density="compact"
        >
          <template v-slot:item.checked_at="{ item }">
            {{ formatDate(item.checked_at) }}
          </template>
          
          <template v-slot:item.status_code="{ item }">
            <v-chip
              v-if="item.status_code"
              :color="getStatusColor(item.status_code)"
              text-color="white"
              small
            >
              {{ item.status_code }}
            </v-chip>
            <v-chip v-else color="error" text-color="white" small>
              Error
            </v-chip>
          </template>
          
          <template v-slot:item.response_time_ms="{ item }">
            {{ item.response_time_ms }}ms
          </template>
        </v-data-table>
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

const API_BASE = import.meta.env.DEV ? '/api/v1' : 'http://localhost:8080/api/v1'

export default {
  name: 'DashboardView',
  
  data() {
    return {
      endpoints: [],
      logs: [],
      selectedEndpoint: null,
      logsDialog: false,
      checkingEndpoints: [],
      loading: false,
      logsLoading: false,
      snackbar: false,
      snackbarText: '',
      snackbarColor: 'success',
      
      headers: [
        { title: 'Name', key: 'name', sortable: true },
        { title: 'URL', key: 'url', sortable: false },
        { title: 'Method', key: 'method', sortable: true },
        { title: 'Status', key: 'status', sortable: false },
        { title: 'Interval', key: 'check_interval_seconds', sortable: true },
        { title: 'Actions', key: 'actions', sortable: false }
      ],
      
      logHeaders: [
        { title: 'Time', key: 'checked_at', sortable: true },
        { title: 'Status Code', key: 'status_code', sortable: true },
        { title: 'Response Time', key: 'response_time_ms', sortable: true },
        { title: 'Error', key: 'error_message', sortable: false }
      ]
    }
  },
  
  computed: {
    totalEndpoints() {
      return this.endpoints.length
    },
    activeEndpoints() {
      return this.endpoints.filter(e => e.is_active).length
    },
    healthyEndpoints() {
      // This would need real-time status data from Prometheus
      return this.activeEndpoints
    },
    unhealthyEndpoints() {
      return this.activeEndpoints - this.healthyEndpoints
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
    
    async manualCheck(endpointId) {
      this.checkingEndpoints.push(endpointId)
      try {
        await axios.post(`${API_BASE}/endpoints/${endpointId}/check`)
        this.showSnackbar('Manual check initiated', 'success')
      } catch (error) {
        this.showSnackbar('Failed to initiate check', 'error')
        console.error(error)
      } finally {
        this.checkingEndpoints = this.checkingEndpoints.filter(id => id !== endpointId)
      }
    },
    
    async viewLogs(endpoint) {
      this.selectedEndpoint = endpoint
      this.logsLoading = true
      this.logsDialog = true
      
      try {
        const response = await axios.get(`${API_BASE}/endpoints/${endpoint.id}/logs?limit=50`)
        this.logs = response.data || []
      } catch (error) {
        this.showSnackbar('Failed to fetch logs', 'error')
        console.error(error)
      } finally {
        this.logsLoading = false
      }
    },
    
    refreshData() {
      this.fetchEndpoints()
    },
    
    formatDate(dateString) {
      return new Date(dateString).toLocaleString()
    },
    
    getStatusColor(statusCode) {
      if (!statusCode) return 'error'
      if (statusCode >= 200 && statusCode < 300) return 'success'
      if (statusCode >= 300 && statusCode < 400) return 'warning'
      return 'error'
    },
    
    showSnackbar(text, color = 'success') {
      this.snackbarText = text
      this.snackbarColor = color
      this.snackbar = true
    }
  }
}
</script>
