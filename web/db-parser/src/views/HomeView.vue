<script setup lang="ts">
import { ref } from 'vue'
import Textarea from 'primevue/textarea'
import Button from 'primevue/button'
import Select from 'primevue/select'
import Dialog from 'primevue/dialog'
import { useToast } from 'primevue/usetoast'
import { parseSql } from '@/services/schemasApi'
import type { ParseSqlRequest } from '@/models/schemas'

const toast = useToast()

const sqlInput = ref('')
const loading = ref(false)
const successModalVisible = ref(false)

const optimizationOptions = [
  { label: 'Balanced', value: 'balanced' },
  { label: 'Read Heavy', value: 'read-heavy' },
  { label: 'Write Heavy', value: 'write-heavy' },
  { label: 'Cost Optimized', value: 'cost-optimized' },
]

const selectedOptimization = ref(optimizationOptions[0])

async function handleParse() {
  if (!sqlInput.value.trim()) {
    toast.add({ severity: 'warn', summary: 'Atención', detail: 'Ingresa una consulta SQL', life: 3000 })
    return
  }

  loading.value = true

  try {
    await parseSql({
      sqlContent: sqlInput.value,
      optimizationType: selectedOptimization.value?.value as ParseSqlRequest['optimizationType'],
    })
    successModalVisible.value = true
    sqlInput.value = ''
  } catch {
    toast.add({ severity: 'error', summary: 'Error', detail: 'Error al enviar la consulta', life: 3000 })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="home-view">
    <h2>SQL to NoSQL Parser</h2>
    <p class="subtitle">Ingresa una consulta SQL y obtén su equivalente NoSQL</p>

    <div class="parser-container">
      <div class="input-section">
        <label for="sql-input">Consulta SQL</label>
        <Textarea
          id="sql-input"
          v-model="sqlInput"
          rows="8"
          placeholder="CREATE TABLE users (id BIGSERIAL PRIMARY KEY, name VARCHAR(100) NOT NULL, ...);"
          autoResize
          class="sql-textarea"
        />

        <div class="actions-row">
          <Select
            v-model="selectedOptimization"
            :options="optimizationOptions"
            optionLabel="label"
            placeholder="Tipo de optimización"
            class="optimization-select"
          />
          <Button label="Parsear" icon="pi pi-play" :loading="loading" @click="handleParse" />
        </div>
      </div>
    </div>

    <!-- Success Modal -->
    <Dialog
      v-model:visible="successModalVisible"
      header="Solicitud Enviada"
      modal
      dismissableMask
      :closable="true"
      :style="{ width: '450px' }"
      :breakpoints="{ '768px': '90vw' }"
    >
      <div class="success-content">
        <i class="pi pi-check-circle success-icon"></i>
        <p>Su solicitud ha sido enviada exitosamente.</p>
        <p class="success-hint">Por favor manténgase atento a la sección de <strong>Historial</strong> para ver el resultado.</p>
      </div>
      <template #footer>
        <Button label="Entendido" icon="pi pi-check" @click="successModalVisible = false" />
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.home-view {
  max-width: 900px;
  margin: 0 auto;
  padding: 2rem;
}

h2 {
  margin-bottom: 0.25rem;
}

.subtitle {
  color: #6b7280;
  margin-bottom: 2rem;
}

.parser-container {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.input-section {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.sql-textarea {
  width: 100%;
  font-family: monospace;
}

.actions-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-top: 0.5rem;
}

.optimization-select {
  min-width: 200px;
}

/* Success modal */
.success-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: 0.5rem;
  padding: 1rem 0;
}

.success-icon {
  font-size: 3rem;
  color: #22c55e;
}

.success-content p {
  margin: 0;
  color: #111827;
  font-size: 0.95rem;
}

.success-hint {
  color: #6b7280 !important;
  font-size: 0.85rem !important;
}
</style>
