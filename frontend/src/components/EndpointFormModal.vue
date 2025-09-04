<template>
  <v-dialog v-model="localDialog" max-width="800px" persistent>
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
            v-model="formData.name"
            label="Name"
            :rules="nameRules"
            required
            outlined
            placeholder="My API Endpoint"
          ></v-text-field>
          
          <v-text-field
            v-model="formData.url"
            label="URL"
            :rules="urlRules"
            required
            outlined
            placeholder="https://api.example.com/health"
          ></v-text-field>
          
          <v-row>
            <v-col cols="6">
              <v-text-field
                v-model.number="formData.check_interval_seconds"
                label="Check Interval (seconds)"
                type="number"
                :min="10"
                :max="3600"
                outlined
              ></v-text-field>
            </v-col>
            
            <v-col cols="6">
              <v-text-field
                v-model.number="formData.timeout_seconds"
                label="Timeout (seconds)"
                type="number"
                :min="5"
                :max="300"
                outlined
              ></v-text-field>
            </v-col>
          </v-row>
          
          <v-select
            v-model="formData.proxy_id"
            :items="proxyOptions"
            item-title="name"
            item-value="id"
            label="Proxy (Optional)"
            clearable
            outlined
            prepend-inner-icon="mdi-server-network"
          >
            <template v-slot:item="{ props, item }">
              <v-list-item v-bind="props" :disabled="!item.raw.is_active">
                <v-list-item-title>{{ item.raw.name }}</v-list-item-title>
                <v-list-item-subtitle>
                  {{ item.raw.host }}:{{ item.raw.port }}
                  <v-chip
                    v-if="!item.raw.is_active"
                    size="x-small"
                    color="error"
                    class="ml-2"
                  >
                    Inactive
                  </v-chip>
                </v-list-item-subtitle>
              </v-list-item>
            </template>
          </v-select>
          
          <v-switch
            v-model="formData.is_active"
            label="Active"
            color="success"
          ></v-switch>
        </v-form>
      </v-card-text>
      
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn text @click="closeDialog">Cancel</v-btn>
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
</template>

<script>
export default {
  name: 'EndpointFormModal',
  
  props: {
    dialog: {
      type: Boolean,
      default: false
    },
    isEditMode: {
      type: Boolean,
      default: false
    },
    endpointData: {
      type: Object,
      default: () => ({
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
      })
    },
    proxyOptions: {
      type: Array,
      default: () => []
    },
    saving: {
      type: Boolean,
      default: false
    }
  },
  
  emits: ['update:dialog', 'save', 'close'],
  
  data() {
    return {
      valid: false,
      formData: { ...this.endpointData },
      
      // Validation rules
      nameRules: [
        v => !!v || 'Name is required',
        v => (v && v.length >= 3) || 'Name must be at least 3 characters',
        v => (v && v.length <= 100) || 'Name must be less than 100 characters'
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
    }
  },
  
  watch: {
    endpointData: {
      handler(newData) {
        this.formData = { ...newData }
      },
      deep: true,
      immediate: true
    },
    
    dialog(newVal) {
      if (!newVal) {
        // Reset form when dialog closes
        this.resetForm()
      }
    }
  },
  
  methods: {
    saveEndpoint() {
      if (!this.valid) return
      
      const payload = {
        ...this.formData,
        method: this.formData.method || 'GET',
        headers: this.formData.headers || {},
        body: this.formData.body || ''
      }
      
      this.$emit('save', payload)
    },
    
    closeDialog() {
      this.$emit('close')
      this.$emit('update:dialog', false)
    },
    
    resetForm() {
      this.formData = {
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
    }
  }
}
</script>
