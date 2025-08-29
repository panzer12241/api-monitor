<template>
  <v-container fluid>
    <v-row>
      <v-col cols="12">
        <v-card>
          <v-card-title>
            <v-icon left>mdi-chart-line</v-icon>
            Grafana Dashboard
            
            <v-spacer></v-spacer>
            
            <v-btn
              color="primary"
              :href="grafanaUrl"
              target="_blank"
              rel="noopener noreferrer"
            >
              <v-icon left>mdi-open-in-new</v-icon>
              Open in New Tab
            </v-btn>
          </v-card-title>
          
          <v-card-text class="pa-0">
            <iframe
              :src="grafanaEmbedUrl"
              width="100%"
              height="800"
              frameborder="0"
              style="border: none;"
            ></iframe>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
export default {
  name: 'GrafanaView',
  
  data() {
    return {
      grafanaUrl: 'http://localhost:3000',
      grafanaEmbedUrl: 'http://localhost:3000/d-solo/api-monitor/api-monitor-dashboard?orgId=1&refresh=30s&theme=dark'
    }
  }
}
</script>

<style scoped>
iframe {
  border-radius: 4px;
}
</style>
