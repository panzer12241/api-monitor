<template>
  <v-dialog v-model="localDialog" max-width="400px">
    <v-card>
      <v-card-title>
        <v-icon left color="error">mdi-delete</v-icon>
        Confirm Delete
      </v-card-title>
      
      <v-card-text>
        Are you sure you want to delete the endpoint "{{ endpoint?.name }}"?
        This action cannot be undone.
      </v-card-text>
      
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn text @click="closeDialog">Cancel</v-btn>
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
</template>

<script>
export default {
  name: 'DeleteConfirmationModal',
  
  props: {
    dialog: {
      type: Boolean,
      default: false
    },
    endpoint: {
      type: Object,
      default: null
    },
    deleting: {
      type: Boolean,
      default: false
    }
  },
  
  emits: ['update:dialog', 'confirm', 'close'],
  
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
  
  methods: {
    confirmDelete() {
      this.$emit('confirm', this.endpoint)
    },
    
    closeDialog() {
      this.$emit('close')
      this.$emit('update:dialog', false)
    }
  }
}
</script>
