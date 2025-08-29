<template>
  <div class="dashboard">
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card class="stats-card">
          <div class="stat-content">
            <div class="stat-number">{{ totalEndpoints }}</div>
            <div class="stat-label">Total Endpoints</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stats-card">
          <div class="stat-content">
            <div class="stat-number">{{ activeEndpoints }}</div>
            <div class="stat-label">Active Endpoints</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stats-card">
          <div class="stat-content">
            <div class="stat-number">{{ healthyEndpoints }}</div>
            <div class="stat-label">Healthy Endpoints</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stats-card">
          <div class="stat-content">
            <div class="stat-number">{{ unhealthyEndpoints }}</div>
            <div class="stat-label">Unhealthy Endpoints</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>Endpoints Status</span>
              <el-button type="primary" @click="refreshData">
                <el-icon><Refresh /></el-icon>
                Refresh
              </el-button>
            </div>
          </template>
          
          <el-table :data="endpoints" style="width: 100%">
            <el-table-column prop="name" label="Name" width="200" />
            <el-table-column prop="url" label="URL" show-overflow-tooltip />
            <el-table-column prop="method" label="Method" width="80" />
            <el-table-column label="Status" width="100">
              <template #default="scope">
                <el-tag :type="scope.row.is_active ? 'success' : 'info'">
                  {{ scope.row.is_active ? 'Active' : 'Inactive' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="check_interval_seconds" label="Interval" width="100">
              <template #default="scope">
                {{ scope.row.check_interval_seconds }}s
              </template>
            </el-table-column>
            <el-table-column label="Actions" width="200">
              <template #default="scope">
                <el-button 
                  size="small" 
                  @click="manualCheck(scope.row.id)"
                  :loading="checkingEndpoints.includes(scope.row.id)"
                >
                  Check Now
                </el-button>
                <el-button 
                  size="small" 
                  @click="viewLogs(scope.row)"
                >
                  View Logs
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <!-- Logs Dialog -->
    <el-dialog 
      v-model="logsDialogVisible" 
      :title="`Logs for ${selectedEndpoint?.name}`"
      width="80%"
    >
      <el-table :data="logs" style="width: 100%">
        <el-table-column prop="checked_at" label="Time" width="180">
          <template #default="scope">
            {{ formatDate(scope.row.checked_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="status_code" label="Status Code" width="120">
          <template #default="scope">
            <el-tag 
              :type="getStatusType(scope.row.status_code)"
              v-if="scope.row.status_code"
            >
              {{ scope.row.status_code }}
            </el-tag>
            <el-tag type="danger" v-else>Error</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="response_time_ms" label="Response Time" width="130">
          <template #default="scope">
            {{ scope.row.response_time_ms }}ms
          </template>
        </el-table-column>
        <el-table-column prop="error_message" label="Error" show-overflow-tooltip />
      </el-table>
    </el-dialog>
  </div>
</template>

<script>
import axios from 'axios'
import { Refresh } from '@element-plus/icons-vue'

const API_BASE = import.meta.env.DEV ? '/api/v1' : 'http://localhost:8080/api/v1'

export default {
  name: 'Dashboard',
  components: {
    Refresh
  },
  data() {
    return {
      endpoints: [],
      logs: [],
      selectedEndpoint: null,
      logsDialogVisible: false,
      checkingEndpoints: []
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
      try {
        const response = await axios.get(`${API_BASE}/endpoints`)
        this.endpoints = response.data || []
      } catch (error) {
        this.$message.error('Failed to fetch endpoints')
        console.error(error)
      }
    },
    async manualCheck(endpointId) {
      this.checkingEndpoints.push(endpointId)
      try {
        await axios.post(`${API_BASE}/endpoints/${endpointId}/check`)
        this.$message.success('Manual check initiated')
      } catch (error) {
        this.$message.error('Failed to initiate check')
        console.error(error)
      } finally {
        this.checkingEndpoints = this.checkingEndpoints.filter(id => id !== endpointId)
      }
    },
    async viewLogs(endpoint) {
      this.selectedEndpoint = endpoint
      try {
        const response = await axios.get(`${API_BASE}/endpoints/${endpoint.id}/logs?limit=50`)
        this.logs = response.data || []
        this.logsDialogVisible = true
      } catch (error) {
        this.$message.error('Failed to fetch logs')
        console.error(error)
      }
    },
    refreshData() {
      this.fetchEndpoints()
    },
    formatDate(dateString) {
      return new Date(dateString).toLocaleString()
    },
    getStatusType(statusCode) {
      if (!statusCode) return 'danger'
      if (statusCode >= 200 && statusCode < 300) return 'success'
      if (statusCode >= 300 && statusCode < 400) return 'warning'
      return 'danger'
    }
  }
}
</script>

<style scoped>
.dashboard {
  padding: 20px;
}

.stats-card {
  text-align: center;
}

.stat-content {
  padding: 20px;
}

.stat-number {
  font-size: 2.5em;
  font-weight: bold;
  color: #409eff;
  margin-bottom: 10px;
}

.stat-label {
  font-size: 1.1em;
  color: #666;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
