<template>
  <div style="height: 300px; position: relative;">
    <Line
      v-if="chartData.datasets.length > 0"
      :data="chartData"
      :options="chartOptions"
    />
    <div v-else class="d-flex align-center justify-center" style="height: 100%;">
      <v-alert type="info" variant="outlined">
        Select an endpoint to view response time trends
      </v-alert>
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
      loading: false
    }
  },
  computed: {
    chartData() {
      if (!this.logs.length) {
        return { labels: [], datasets: [] }
      }
      
      // Get last 20 logs, sorted by time
      const sortedLogs = [...this.logs]
        .sort((a, b) => new Date(a.checked_at) - new Date(b.checked_at))
        .slice(-20)
      
      const labels = sortedLogs.map(log => 
        new Date(log.checked_at).toLocaleTimeString()
      )
      
      const responseTimes = sortedLogs.map(log => log.response_time_ms || 0)
      
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
      immediate: false
    }
  },
  methods: {
    async fetchLogs() {
      if (!this.endpointId) {
        this.logs = []
        return
      }
      
      this.loading = true
      try {
        const response = await axios.get(`${API_BASE}/endpoints/${this.endpointId}/logs?limit=50`)
        this.logs = response.data || []
      } catch (error) {
        console.error('Failed to fetch logs for chart:', error)
        this.logs = []
      } finally {
        this.loading = false
      }
    },
    
    updateChart() {
      // Refresh the chart data by fetching logs again
      this.fetchLogs()
    }
  }
}
</script>
