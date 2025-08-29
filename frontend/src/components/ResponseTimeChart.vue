<template>
  <div style="position: relative;">
    <!-- Period selector -->
    <div class="d-flex justify-end mb-2">
      <v-select
        v-model="selectedPeriod"
        :items="periodOptions"
        item-title="text"
        item-value="value"
        label="Select period"
        variant="outlined"
        density="compact"
        style="max-width: 150px;"
        hide-details
      />
    </div>
    
    <div style="height: 300px; position: relative;">
      <div v-if="loading" class="d-flex align-center justify-center" style="height: 100%;">
        <v-progress-circular indeterminate color="primary"></v-progress-circular>
      </div>
      <Line
        v-else-if="chartData.datasets.length > 0 && logs.length > 0"
        :data="chartData"
        :options="chartOptions"
      />
      <div v-else class="d-flex align-center justify-center" style="height: 100%;">
        <v-alert type="info" variant="outlined">
          Select an endpoint to view response time trends
        </v-alert>
      </div>
    </div>
  </div>
</template>

<script>
import {
  Chart as ChartJS,
  Title,
  Tooltip,
  Legend,
  LineElement,
  LinearScale,
  CategoryScale,
  PointElement
} from 'chart.js'
import { Line } from 'vue-chartjs'
import axios from 'axios'

ChartJS.register(
  Title,
  Tooltip,
  Legend,
  LineElement,
  LinearScale,
  CategoryScale,
  PointElement
)

const API_BASE = import.meta.env.DEV ? '/api/v1' : 'http://localhost:8080/api/v1'

export default {
  name: 'ResponseTimeChart',
  components: {
    Line
  },
  props: {
    endpointId: {
      type: [Number, String],
      default: null
    }
  },
  data() {
    return {
      logs: [],
      loading: false,
      selectedPeriod: '1h',
      periodOptions: [
        { text: '1 hour', value: '1h' },
        { text: '6 hours', value: '6h' },
        { text: '24 hours', value: '24h' },
        { text: '7 days', value: '7d' },
        { text: '14 days', value: '14d' }
      ]
    }
  },
  computed: {
    chartData() {
      if (!this.logs.length) {
        return { labels: [], datasets: [] }
      }
      
      // Sort logs by time
      const sortedLogs = [...this.logs]
        .sort((a, b) => new Date(a.checked_at) - new Date(b.checked_at))
      
      // Limit data points based on period for better performance
      const maxPoints = this.selectedPeriod === '1h' ? 60 : 
                      this.selectedPeriod === '6h' ? 72 : 
                      this.selectedPeriod === '24h' ? 48 : 100
      
      const displayLogs = sortedLogs.slice(-maxPoints)
      
      // Format time labels based on selected period
      const labels = displayLogs.map(log => {
        const date = new Date(log.checked_at)
        
        if (this.selectedPeriod === '1h' || this.selectedPeriod === '6h') {
          // Show time only for short periods
          return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
        } else if (this.selectedPeriod === '24h') {
          // Show time with AM/PM for 24h
          return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
        } else {
          // Show date and time for longer periods
          return date.toLocaleDateString([], { month: 'short', day: 'numeric' }) + ' ' +
                 date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
        }
      })
      
      const responseTimes = displayLogs.map(log => log.response_time_ms || 0)
      
      return {
        labels,
        datasets: [
          {
            label: 'Response Time (ms)',
            data: responseTimes,
            borderColor: '#2196F3',
            backgroundColor: 'rgba(33, 150, 243, 0.1)',
            borderWidth: 2,
            fill: true,
            tension: 0.4,
            pointBackgroundColor: '#2196F3',
            pointBorderColor: '#ffffff',
            pointBorderWidth: 2,
            pointRadius: 4
          }
        ]
      }
    },
    chartOptions() {
      return {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: {
            display: false
          },
          tooltip: {
            mode: 'index',
            intersect: false
          }
        },
        scales: {
          x: {
            display: true,
            title: {
              display: true,
              text: 'Time'
            }
          },
          y: {
            display: true,
            title: {
              display: true,
              text: 'Response Time (ms)'
            },
            beginAtZero: true
          }
        },
        interaction: {
          mode: 'nearest',
          axis: 'x',
          intersect: false
        }
      }
    }
  },
  watch: {
    endpointId: {
      handler: 'fetchLogs',
      immediate: true
    },
    selectedPeriod: {
      handler: 'fetchLogs',
      immediate: false
    }
  },
  methods: {
    async fetchLogs() {
      if (!this.endpointId) {
        this.logs = []
        return Promise.resolve()
      }
      
      this.loading = true
      try {
        // Calculate date range based on selected period
        const now = new Date()
        const startDate = new Date()
        
        switch (this.selectedPeriod) {
          case '1h':
            startDate.setHours(now.getHours() - 1)
            break
          case '6h':
            startDate.setHours(now.getHours() - 6)
            break
          case '24h':
            startDate.setDate(now.getDate() - 1)
            break
          case '7d':
            startDate.setDate(now.getDate() - 7)
            break
          case '14d':
            startDate.setDate(now.getDate() - 14)
            break
          default:
            startDate.setHours(now.getHours() - 1)
        }
        
        // Format dates for API
        const startISO = startDate.toISOString()
        const endISO = now.toISOString()
        
        const url = `${API_BASE}/endpoints/${this.endpointId}/logs?start_date=${startISO}&end_date=${endISO}&limit=200`
        const response = await axios.get(url)
        this.logs = response.data.logs || response.data || []
        return Promise.resolve()
      } catch (error) {
        console.error('Failed to fetch logs for chart:', error)
        this.logs = []
        return Promise.reject(error)
      } finally {
        this.loading = false
      }
    },
    
    updateChart() {
      // Refresh the chart data by fetching logs again
      return this.fetchLogs()
    }
  }
}
</script>
