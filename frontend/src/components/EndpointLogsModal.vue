<template>
  <v-dialog v-model="localDialog" max-width="1400px" fullscreen-mobile>
    <v-card class="logs-dialog-card">
      <v-card-title class="pa-6 pb-4 position-relative">
        <v-icon left>mdi-history</v-icon>
        Endpoint Logs: {{ endpoint?.name || endpoint?.url }}
        <v-spacer></v-spacer>
        <v-btn 
          color="primary" 
          @click="refreshLogs" 
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
          @click="closeDialog"
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
                    v-model="localFilters.startDate"
                    label="Start Date"
                    type="datetime-local"
                    outlined
                    dense
                    @update:model-value="onFilterChange"
                  ></v-text-field>
                </v-col>
                
                <v-col cols="12" md="3">
                  <v-text-field
                    v-model="localFilters.endDate"
                    label="End Date"
                    type="datetime-local"
                    outlined
                    dense
                    @update:model-value="onFilterChange"
                  ></v-text-field>
                </v-col>
                
                <v-col cols="12" md="3">
                  <v-text-field
                    v-model.number="localFilters.minResponseTime"
                    label="Min Response Time (ms)"
                    type="number"
                    outlined
                    dense
                    placeholder="e.g. 1000"
                    @update:model-value="onFilterChange"
                  ></v-text-field>
                </v-col>
                
                <v-col cols="12" md="3">
                  <v-select
                    v-model="localFilters.statusCode"
                    :items="statusCodeOptions"
                    label="Status Code"
                    outlined
                    dense
                    clearable
                    @update:model-value="onFilterChange"
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
              v-model="localLogsPerPage"
              :items="[10, 25, 50, 100]"
              label="Items per page"
              density="compact"
              style="width: 140px;"
              @update:model-value="onLogsPerPageChange"
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
              @click="loadMore"
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
</template>

<script>
export default {
  name: 'EndpointLogsModal',
  
  props: {
    dialog: {
      type: Boolean,
      default: false
    },
    endpoint: {
      type: Object,
      default: null
    },
    logs: {
      type: Array,
      default: () => []
    },
    logsLoading: {
      type: Boolean,
      default: false
    },
    loadingMore: {
      type: Boolean,
      default: false
    },
    totalLogs: {
      type: Number,
      default: 0
    },
    logsPerPage: {
      type: Number,
      default: 25
    },
    filters: {
      type: Object,
      default: () => ({
        startDate: '',
        endDate: '',
        minResponseTime: null,
        statusCode: null
      })
    }
  },
  
  emits: ['update:dialog', 'close', 'refresh', 'filter-change', 'load-more', 'logs-per-page-change', 'clear-filters'],
  
  data() {
    return {
      filtersExpanded: [],
      localFilters: { ...this.filters },
      localLogsPerPage: this.logsPerPage,
      
      statusCodeOptions: [
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
  
  computed: {
    localDialog: {
      get() {
        return this.dialog
      },
      set(value) {
        this.$emit('update:dialog', value)
      }
    },
    
    hasMoreLogs() {
      return this.logs.length < this.totalLogs
    }
  },
  
  watch: {
    filters: {
      handler(newFilters) {
        this.localFilters = { ...newFilters }
      },
      deep: true,
      immediate: true
    },
    
    logsPerPage(newValue) {
      this.localLogsPerPage = newValue
    }
  },
  
  methods: {
    formatDate(dateString) {
      if (!dateString) return 'N/A'
      
      try {
        const date = new Date(dateString)
        return date.toLocaleString('en-US', {
          year: 'numeric',
          month: '2-digit',
          day: '2-digit',
          hour: '2-digit',
          minute: '2-digit',
          second: '2-digit'
        })
      } catch {
        return dateString
      }
    },
    
    parseHeaders(headersString) {
      if (!headersString || headersString.trim() === '') {
        return {}
      }
      
      try {
        // Try to parse as JSON first
        return JSON.parse(headersString)
      } catch {
        // If that fails, try to parse as key: value pairs
        const headers = {}
        const lines = headersString.split('\n')
        
        for (const line of lines) {
          const colonIndex = line.indexOf(':')
          if (colonIndex > 0) {
            const key = line.substring(0, colonIndex).trim()
            const value = line.substring(colonIndex + 1).trim()
            if (key && value) {
              headers[key] = value
            }
          }
        }
        
        return headers
      }
    },
    
    formatResponseBody(body) {
      if (!body || body.trim() === '') {
        return 'No response body'
      }
      
      try {
        // Try to parse and format JSON
        const parsed = JSON.parse(body)
        return JSON.stringify(parsed, null, 2)
      } catch {
        // Return as-is if not JSON
        return body
      }
    },
    
    refreshLogs() {
      this.$emit('refresh')
    },
    
    onFilterChange() {
      this.$emit('filter-change', this.localFilters)
    },
    
    clearFilters() {
      this.localFilters = {
        startDate: '',
        endDate: '',
        minResponseTime: null,
        statusCode: null
      }
      this.$emit('clear-filters')
    },
    
    loadMore() {
      this.$emit('load-more')
    },
    
    onLogsPerPageChange() {
      this.$emit('logs-per-page-change', this.localLogsPerPage)
    },
    
    closeDialog() {
      this.$emit('close')
      this.$emit('update:dialog', false)
    }
  }
}
</script>

<style scoped>
.logs-dialog-card {
  height: 90vh;
}

.close-btn {
  position: absolute !important;
  top: 16px;
  right: 16px;
  z-index: 1;
}

.logs-expansion-panels {
  max-height: calc(70vh - 100px);
  overflow-y: auto;
}

.cursor-pointer {
  cursor: pointer;
}
</style>
