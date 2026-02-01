<script setup lang="ts">
import { ref } from 'vue'
import Textarea from 'primevue/textarea'
import Button from 'primevue/button'
import { useToast } from 'primevue/usetoast'

const toast = useToast()

const sqlInput = ref('')
const result = ref('')
const loading = ref(false)

async function handleParse() {
  if (!sqlInput.value.trim()) {
    toast.add({ severity: 'warn', summary: 'Atención', detail: 'Ingresa una consulta SQL', life: 3000 })
    return
  }

  loading.value = true
  result.value = ''

  try {
    // TODO: connect to actual parser API
    await new Promise((r) => setTimeout(r, 500))
    result.value = JSON.stringify({ message: 'Parser no conectado aún', input: sqlInput.value }, null, 2)
    toast.add({ severity: 'success', summary: 'Listo', detail: 'Consulta procesada', life: 3000 })
  } catch {
    toast.add({ severity: 'error', summary: 'Error', detail: 'Error al parsear la consulta', life: 3000 })
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
          placeholder="SELECT * FROM users WHERE age > 25"
          autoResize
          class="sql-textarea"
        />
        <Button label="Parsear" icon="pi pi-play" :loading="loading" @click="handleParse" class="parse-btn" />
      </div>

      <div v-if="result" class="result-section">
        <label>Resultado</label>
        <pre class="result-output">{{ result }}</pre>
      </div>
    </div>
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

.parse-btn {
  align-self: flex-start;
  margin-top: 0.5rem;
}

.result-section {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.result-output {
  background: #1e1e2e;
  color: #a6e3a1;
  padding: 1rem;
  border-radius: 8px;
  overflow-x: auto;
  font-size: 0.9rem;
  line-height: 1.5;
}
</style>
