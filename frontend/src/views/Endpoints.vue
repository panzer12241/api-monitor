<template>
  <v-container fluid>
    <v-row>
      <v-col cols="12">
        <v-card>
          <v-card-title>
            <v-icon left>mdi-api</v-icon>
            API Endpoints Management
            
            <v-spacer></v-spacer>
            
            <v-btn color="primary" @click="showCreateDialog">
              <v-icon left>mdi-plus</v-icon>
              Add Endpoint
            </v-btn>
          </v-card-title>
          
          <v-data-table
            :headers="headers"
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
            ></v-text-field>
            
            <v-text-field
              v-model="endpointForm.url"
              label="URL"
              :rules="urlRules"
              required
              outlined
              placeholder="https://api.example.com/health"
            ></v-text-field>
            
            <v-select
              v-model="endpointForm.method"
              :items="httpMethods"
              label="Method"
              outlined
            ></v-select>
            
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
            
            <!-- Headers Section -->
            <v-card outlined class="mb-4">
              <v-card-subtitle>
                <v-icon left small>mdi-code-tags</v-icon>
                HTTP Headers
              </v-card-subtitle>
              
              <v-card-text>
                <div v-for="(header, index) in headers" :key="index" class="mb-2">
                  <v-row>
                    <v-col cols="5">
                      <v-text-field
                        v-model="header.key"
                        label="Header name"
                        outlined
                        dense
                      ></v-text-field>
                    </v-col>
                    <v-col cols="5">
                      <v-text-field
                        v-model="header.value"
                        label="Header value"
                        outlined
                        dense
                      ></v-text-field>
                    </v-col>
                    <v-col cols="2">
                      <v-btn
                        color="error"
                        icon
                        @click="removeHeader(index)"
                      >
                        <v-icon>mdi-delete</v-icon>
                      </v-btn>
                    </v-col>
                  </v-row>
                </div>
                
                <v-btn color="primary" outlined @click="addHeader">
                  <v-icon left>mdi-plus</v-icon>
                  Add Header
                </v-btn>
              </v-card-text>
            </v-card>
            
            <v-textarea
              v-model="endpointForm.body"
              label="Request Body"
              outlined
              rows="4"
              placeholder="JSON request body (for POST/PUT requests)"
            ></v-textarea>
            
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
  name: 'EndpointsView',
  
  data() {
    return {
      endpoints: [],
      dialog: false,
      deleteDialog: false,
      isEditMode: false,
      loading: false,
      saving: false,
      deleting: false,
      valid: false,
      snackbar: false,
      snackbarText: '',
      snackbarColor: 'success',
      deletingEndpoint: null,
      
      endpointForm: {
        id: null,
        name: '',
        url: '',
        method: 'GET',
        headers: {},
        body: '',
        timeout_seconds: 30,
        check_interval_seconds: 60,
        is_active: true
      },
      
      headers: [],
      httpMethods: ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD'],
      
      nameRules: [
        v => !!v || 'Name is required',
        v => v.length >= 3 || 'Name must be at least 3 characters'
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
      
      tableHeaders: [
        { title: 'Name', key: 'name', sortable: true },
        { title: 'URL', key: 'url', sortable: false },
        { title: 'Method', key: 'method', sortable: true },
        { title: 'Status', key: 'status', sortable: false },
        { title: 'Interval', key: 'check_interval_seconds', sortable: true },
        { title: 'Timeout', key: 'timeout_seconds', sortable: true },
        { title: 'Actions', key: 'actions', sortable: false }
      ]
    }
  },
  
  computed: {
    headers() {
      return this.tableHeaders
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
    
    showCreateDialog() {
      this.resetForm()
      this.isEditMode = false
      this.dialog = true
    },
    
    editEndpoint(endpoint) {
      this.endpointForm = { ...endpoint }
      this.headers = Object.entries(endpoint.headers || {}).map(([key, value]) => ({ key, value }))
      this.isEditMode = true
      this.dialog = true
    },
    
    async saveEndpoint() {
      if (!this.valid) return
      
      this.saving = true
      
      // Convert headers array to object
      this.endpointForm.headers = {}
      this.headers.forEach(header => {
        if (header.key && header.value) {
          this.endpointForm.headers[header.key] = header.value
        }
      })
      
      try {
        if (this.isEditMode) {
          await axios.put(`${API_BASE}/endpoints/${this.endpointForm.id}`, this.endpointForm)
          this.showSnackbar('Endpoint updated successfully', 'success')
        } else {
          await axios.post(`${API_BASE}/endpoints`, this.endpointForm)
          this.showSnackbar('Endpoint created successfully', 'success')
        }
        
        this.dialog = false
        this.fetchEndpoints()
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
        // Revert the switch
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
        this.fetchEndpoints()
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
        is_active: true
      }
      this.headers = []
    },
    
    addHeader() {
      this.headers.push({ key: '', value: '' })
    },
    
    removeHeader(index) {
      this.headers.splice(index, 1)
    },
    
    showSnackbar(text, color = 'success') {
      this.snackbarText = text
      this.snackbarColor = color
      this.snackbar = true
    }
  }
}
</script>
