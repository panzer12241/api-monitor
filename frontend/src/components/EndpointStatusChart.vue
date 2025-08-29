<template>
  <div style="height: 300px; position: relative;">
    <Doughnut
      :data="chartData"
      :options="chartOptions"
    />
  </div>
</template>

<script>
import {
  Chart as ChartJS,
  Title,
  Tooltip,
  Legend,
  ArcElement
} from 'chart.js'
import { Doughnut } from 'vue-chartjs'

ChartJS.register(Title, Tooltip, Legend, ArcElement)

export default {
  name: 'EndpointStatusChart',
  components: {
    Doughnut
  },
  props: {
    endpoints: {
      type: Array,
      default: () => []
    }
  },
  computed: {
    chartData() {
      const active = this.endpoints.filter(e => e.is_active).length
      const inactive = this.endpoints.length - active
      
      return {
        labels: ['Active', 'Inactive'],
        datasets: [
          {
            data: [active, inactive],
            backgroundColor: [
              '#4CAF50', // Green for active
              '#F44336'  // Red for inactive
            ],
            borderWidth: 2,
            borderColor: '#ffffff'
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
            position: 'bottom'
          },
          tooltip: {
            callbacks: {
              label: (context) => {
                const label = context.label || ''
                const value = context.parsed
                const total = context.dataset.data.reduce((a, b) => a + b, 0)
                const percentage = total > 0 ? Math.round((value / total) * 100) : 0
                return `${label}: ${value} (${percentage}%)`
              }
            }
          }
        }
      }
    }
  },
  methods: {
    updateChart() {
      // Force reactivity by updating the chart
      // The computed chartData will automatically recalculate
      this.$forceUpdate()
    }
  }
}
</script>
