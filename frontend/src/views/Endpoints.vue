<template>
  <div class="endpoints">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>API Endpoints Management</span>
          <el-button type="primary" @click="showCreateDialog">
            <el-icon><Plus /></el-icon>
            Add Endpoint
          </el-button>
        </div>
      </template>
      
      <el-table :data="endpoints" style="width: 100%">
        <el-table-column prop="name" label="Name" width="200" />
        <el-table-column prop="url" label="URL" show-overflow-tooltip />
        <el-table-column prop="method" label="Method" width="100" />
        <el-table-column label="Status" width="120">
          <template #default="scope">
            <el-switch
              v-model="scope.row.is_active"
              @change="toggleEndpoint(scope.row)"
              active-text="Active"
              inactive-text="Inactive"
            />
          </template>
        </el-table-column>
        <el-table-column prop="check_interval_seconds" label="Interval" width="100">
          <template #default="scope">
            {{ scope.row.check_interval_seconds }}s
          </template>
        </el-table-column>
        <el-table-column prop="timeout_seconds" label="Timeout" width="100">
          <template #default="scope">
            {{ scope.row.timeout_seconds }}s
          </template>
        </el-table-column>
        <el-table-column label="Actions" width="200">
          <template #default="scope">
            <el-button size="small" @click="editEndpoint(scope.row)">
              <el-icon><Edit /></el-icon>
              Edit
            </el-button>
            <el-button 
              size="small" 
              type="danger" 
              @click="deleteEndpoint(scope.row)"
            >
              <el-icon><Delete /></el-icon>
              Delete
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Create/Edit Dialog -->
    <el-dialog 
      v-model="dialogVisible" 
      :title="isEditMode ? 'Edit Endpoint' : 'Create Endpoint'"
      width="600px"
    >
      <el-form :model="endpointForm" label-width="140px">
        <el-form-item label="Name" required>
          <el-input v-model="endpointForm.name" placeholder="Enter endpoint name" />
        </el-form-item>
        
        <el-form-item label="URL" required>
          <el-input v-model="endpointForm.url" placeholder="https://api.example.com/health" />
        </el-form-item>
        
        <el-form-item label="Method">
          <el-select v-model="endpointForm.method">
            <el-option label="GET" value="GET" />
            <el-option label="POST" value="POST" />
            <el-option label="PUT" value="PUT" />
            <el-option label="DELETE" value="DELETE" />
            <el-option label="PATCH" value="PATCH" />
            <el-option label="HEAD" value="HEAD" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="Check Interval">
          <el-input-number 
            v-model="endpointForm.check_interval_seconds"
            :min="10"
            :max="3600"
            controls-position="right"
          />
          <span style="margin-left: 10px;">seconds</span>
        </el-form-item>
        
        <el-form-item label="Timeout">
          <el-input-number 
            v-model="endpointForm.timeout_seconds"
            :min="5"
            :max="300"
            controls-position="right"
          />
          <span style="margin-left: 10px;">seconds</span>
        </el-form-item>
        
        <el-form-item label="Headers">
          <div v-for="(header, index) in headers" :key="index" style="margin-bottom: 10px;">
            <el-row :gutter="10">
              <el-col :span="10">
                <el-input v-model="header.key" placeholder="Header name" />
              </el-col>
              <el-col :span="10">
                <el-input v-model="header.value" placeholder="Header value" />
              </el-col>
              <el-col :span="4">
                <el-button type="danger" @click="removeHeader(index)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </el-col>
            </el-row>
          </div>
          <el-button @click="addHeader">Add Header</el-button>
        </el-form-item>
        
        <el-form-item label="Request Body">
          <el-input 
            v-model="endpointForm.body"
            type="textarea"
            :rows="4"
            placeholder="JSON request body (for POST/PUT requests)"
          />
        </el-form-item>
        
        <el-form-item label="Active">
          <el-switch v-model="endpointForm.is_active" />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">Cancel</el-button>
          <el-button type="primary" @click="saveEndpoint">
            {{ isEditMode ? 'Update' : 'Create' }}
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import axios from 'axios'
import { Plus, Edit, Delete } from '@element-plus/icons-vue'

const API_BASE = 'http://localhost:8080/api/v1'

export default {
  name: 'Endpoints',
  components: {
    Plus,
    Edit,
    Delete
  },
  data() {
    return {
      endpoints: [],
      dialogVisible: false,
      isEditMode: false,
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
      headers: []
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
    
    showCreateDialog() {
      this.resetForm()
      this.isEditMode = false
      this.dialogVisible = true
    },
    
    editEndpoint(endpoint) {
      this.endpointForm = { ...endpoint }
      this.headers = Object.entries(endpoint.headers || {}).map(([key, value]) => ({ key, value }))
      this.isEditMode = true
      this.dialogVisible = true
    },
    
    async saveEndpoint() {
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
          this.$message.success('Endpoint updated successfully')
        } else {
          await axios.post(`${API_BASE}/endpoints`, this.endpointForm)
          this.$message.success('Endpoint created successfully')
        }
        
        this.dialogVisible = false
        this.fetchEndpoints()
      } catch (error) {
        this.$message.error('Failed to save endpoint')
        console.error(error)
      }
    },
    
    async toggleEndpoint(endpoint) {
      try {
        await axios.post(`${API_BASE}/endpoints/${endpoint.id}/toggle`)
        this.$message.success(`Endpoint ${endpoint.is_active ? 'activated' : 'deactivated'}`)
      } catch (error) {
        this.$message.error('Failed to toggle endpoint')
        // Revert the switch
        endpoint.is_active = !endpoint.is_active
        console.error(error)
      }
    },
    
    async deleteEndpoint(endpoint) {
      try {
        await this.$confirm('Are you sure you want to delete this endpoint?', 'Warning', {
          confirmButtonText: 'Delete',
          cancelButtonText: 'Cancel',
          type: 'warning'
        })
        
        await axios.delete(`${API_BASE}/endpoints/${endpoint.id}`)
        this.$message.success('Endpoint deleted successfully')
        this.fetchEndpoints()
      } catch (error) {
        if (error !== 'cancel') {
          this.$message.error('Failed to delete endpoint')
          console.error(error)
        }
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
    }
  }
}
</script>

<style scoped>
.endpoints {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.dialog-footer {
  text-align: right;
}
</style>
