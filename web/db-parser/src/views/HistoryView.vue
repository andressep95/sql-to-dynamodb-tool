<script setup lang="ts">
import { ref, onMounted } from 'vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import Tag from 'primevue/tag'
import { useToast } from 'primevue/usetoast'
import { getSchemas } from '@/services/schemasApi'
import type { Conversion } from '@/models/schemas'

const toast = useToast()
const conversions = ref<Conversion[]>([])
const loading = ref(false)
const selectedConversion = ref<Conversion | null>(null)
const modalVisible = ref(false)

onMounted(() => {
  fetchConversions()
})

async function fetchConversions() {
  loading.value = true
  try {
    const data = await getSchemas()
    conversions.value = data.conversions
  } catch {
    toast.add({ severity: 'error', summary: 'Error', detail: 'No se pudo cargar el historial', life: 3000 })
  } finally {
    loading.value = false
  }
}

function openDetail(conversion: Conversion) {
  selectedConversion.value = conversion
  modalVisible.value = true
}

function formatDate(dateStr: string) {
  if (!dateStr) return '—'
  const d = new Date(dateStr)
  return d.toLocaleDateString('es-CL', { year: 'numeric', month: 'short', day: 'numeric' })
}

function formatTime(dateStr: string) {
  if (!dateStr) return '—'
  const d = new Date(dateStr)
  return d.toLocaleTimeString('es-CL', {
    hour: '2-digit', minute: '2-digit', second: '2-digit',
  })
}

function formatDateTime(dateStr: string) {
  if (!dateStr) return '—'
  const d = new Date(dateStr)
  return d.toLocaleString('es-CL', {
    year: 'numeric', month: 'short', day: 'numeric',
    hour: '2-digit', minute: '2-digit', second: '2-digit',
  })
}

function statusSeverity(status: string): "success" | "info" | "warn" | "danger" | "secondary" | "contrast" | undefined {
  const map: Record<string, "success" | "warn" | "danger" | "info"> = {
    COMPLETED: 'success',
    PENDING: 'warn',
    FAILED: 'danger',
    PROCESSING: 'info',
  }
  return map[status] ?? 'info'
}

function formatSql(sql: string): string {
  return sql
    .replace(/\(\s*/g, '(\n  ')
    .replace(/,\s*/g, ',\n  ')
    .replace(/\)\s*;/g, '\n);')
    .replace(/\)\s*\)/g, ')\n)')
}
</script>

<template>
  <div class="history-view">
    <h2>Historial de Conversiones</h2>
    <p class="subtitle">Aquí aparecerán tus conversiones anteriores</p>

    <!-- Empty state -->
    <div v-if="!loading && conversions.length === 0" class="empty-state">
      <i class="pi pi-history empty-icon"></i>
      <p>No hay conversiones registradas</p>
    </div>

    <!-- Table -->
    <DataTable
      v-if="conversions.length > 0 || loading"
      :value="conversions"
      :loading="loading"
      stripedRows
      :rowHover="true"
      :paginator="conversions.length > 10"
      :rows="10"
      class="conversions-table"
      @row-click="(e: any) => openDetail(e.data)"
    >
      <Column field="conversionDate" header="Fecha" sortable>
        <template #body="{ data }">
          {{ formatDate(data.conversionDate) }}
        </template>
      </Column>
      <Column field="createdAt" header="Hora" sortable>
        <template #body="{ data }">
          {{ formatTime(data.createdAt) }}
        </template>
      </Column>
      <Column field="status" header="Estado" sortable>
        <template #body="{ data }">
          <Tag :value="data.status" :severity="statusSeverity(data.status)" />
        </template>
      </Column>
      <Column field="optimizationType" header="Optimización" sortable>
        <template #body="{ data }">
          <Tag :value="data.optimizationType" severity="info" />
        </template>
      </Column>
    </DataTable>

    <!-- Detail Modal -->
    <Dialog
      v-model:visible="modalVisible"
      header="Detalle de Conversión"
      modal
      dismissableMask
      :closable="true"
      :style="{ width: '95vw', maxWidth: '1200px', maxHeight: '90vh' }"
      :breakpoints="{ '768px': '98vw' }"
      :contentStyle="{ overflow: 'auto' }"
    >
      <template v-if="selectedConversion">
        <!-- Meta info bar -->
        <div class="detail-meta-bar">
          <div class="meta-item">
            <span class="meta-label">Fecha</span>
            <span class="meta-value">{{ formatDate(selectedConversion.conversionDate) }}</span>
          </div>
          <div class="meta-item">
            <span class="meta-label">Hora</span>
            <span class="meta-value">{{ formatTime(selectedConversion.createdAt) }}</span>
          </div>
          <div class="meta-item">
            <span class="meta-label">Expira</span>
            <span class="meta-value">{{ formatDateTime(new Date(Number(selectedConversion.expiresAt) * 1000).toISOString()) }}</span>
          </div>
          <div class="meta-item">
            <Tag :value="selectedConversion.status" :severity="statusSeverity(selectedConversion.status)" />
          </div>
          <div class="meta-item">
            <Tag :value="selectedConversion.optimizationType" severity="info" />
          </div>
          <div class="meta-item">
            <span class="meta-label">Tablas Extraidas:</span>
            <span class="meta-value">{{ selectedConversion.tablesExtracted }}</span>
          </div>
          <template v-for="table in selectedConversion.noSqlSchema.tables" :key="'sk-' + table.tableName">
            <div v-if="table.sortKey" class="meta-item">
              <span class="meta-label">Sort Key ({{ table.tableName }})</span>
              <span class="meta-value">{{ table.sortKey.name }} ({{ table.sortKey.type }})</span>
            </div>
          </template>
        </div>

        <!-- Split pane -->
        <div class="split-pane">
          <!-- LEFT: SQL -->
          <div class="pane pane-left">
            <div class="pane-header">
              <i class="pi pi-code"></i>
              <span>SQL Original</span>
            </div>
            <pre class="code-block">{{ formatSql(selectedConversion.sqlContent) }}</pre>
          </div>

          <!-- RIGHT: NoSQL Schema -->
          <div class="pane pane-right">
            <div class="pane-header">
              <i class="pi pi-database"></i>
              <span>DynamoDB</span>
            </div>
            <div class="nosql-content">
              <div v-for="table in selectedConversion.noSqlSchema.tables" :key="table.tableName" class="table-card">
                <h5 class="table-name">
                  <i class="pi pi-table"></i>
                  {{ table.tableName }}
                </h5>

                <h6>Billing Mode</h6>
                <div class="billing-badge">
                  <i class="pi pi-money-bill"></i>
                  {{ table.billingMode.replace(/_/g, ' ') }}
                </div>

                <h6>Partition Key</h6>
                <div class="attributes-list">
                  <Tag :value="`${table.partitionKey.name} (${table.partitionKey.type})`" severity="secondary" />
                </div>

                <h6>Sort Key</h6>
                <div class="attributes-list">
                  <Tag
                    :value="table.sortKey ? `${table.sortKey.name} (${table.sortKey.type})` : 'NULL'"
                    severity="secondary"
                  />
                </div>

                <h6>Atributos</h6>
                <div class="attributes-list">
                  <Tag
                    v-for="attr in table.attributes"
                    :key="attr.name"
                    :value="`${attr.name} (${attr.type})`"
                    severity="secondary"
                  />
                </div>

                <template v-if="table.globalSecondaryIndexes.length > 0">
                  <h6>Global Secondary Indexes</h6>
                  <div class="gsi-list">
                    <div v-for="gsi in table.globalSecondaryIndexes" :key="gsi.indexName" class="gsi-card">
                      <div class="gsi-name">
                        <i class="pi pi-search"></i>
                        {{ gsi.indexName }}
                      </div>
                      <div class="gsi-details">
                        <span class="key-badge pk">Partition Key</span>
                        <span>{{ gsi.partitionKey.name }} ({{ gsi.partitionKey.type }})</span>
                        <Tag :value="gsi.projectionType" severity="info" size="small" />
                      </div>
                    </div>
                  </div>
                </template>
              </div>
            </div>
          </div>
        </div>
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.history-view {
  max-width: 960px;
  margin: 0 auto;
  padding: 2rem;
}

.subtitle {
  color: #6b7280;
  margin-bottom: 2rem;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  color: #9ca3af;
  gap: 1rem;
}

.empty-icon {
  font-size: 3rem;
}

.conversions-table :deep(tr) {
  cursor: pointer;
}

/* Meta bar */
.detail-meta-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 1.25rem;
  padding: 0.75rem 1rem;
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  margin-bottom: 1rem;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.8rem;
}

.meta-label {
  font-weight: 600;
  color: #6b7280;
}

.meta-value {
  color: #111827;
}

.meta-value.mono {
  font-family: monospace;
  font-size: 0.75rem;
  color: #6b7280;
}

/* Split pane layout */
.split-pane {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
  min-height: 400px;
}

.pane {
  display: flex;
  flex-direction: column;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  overflow: hidden;
}

.pane-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.65rem 1rem;
  background: #f3f4f6;
  border-bottom: 1px solid #e5e7eb;
  font-weight: 600;
  font-size: 0.85rem;
  color: #374151;
}

.pane-header i {
  color: #3b82f6;
}

/* Left pane: SQL */
.pane-left .code-block {
  flex: 1;
  margin: 0;
  border-radius: 0;
  background: #1e1e2e;
  color: #a6e3a1;
  padding: 1rem;
  overflow: auto;
  font-size: 0.85rem;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
}

/* Right pane: NoSQL */
.nosql-content {
  flex: 1;
  overflow: auto;
  padding: 1rem;
}

.table-card {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 1rem;
  margin-bottom: 1rem;
}

.table-card:last-child {
  margin-bottom: 0;
}

.table-name {
  margin: 0 0 0.25rem;
  color: #3b82f6;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 1rem;
}

.table-name i {
  font-size: 0.9rem;
}

.key-section {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  margin-bottom: 0.75rem;
}

.key-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.85rem;
  color: #374151;
}

.key-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 28px;
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.03em;
}

.key-badge.pk {
  background: #dbeafe;
  color: #1d4ed8;
}

.key-badge.sk {
  background: #fef3c7;
  color: #92400e;
}

.billing-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.8rem;
  color: #6b7280;
}

h6 {
  margin: 0.75rem 0 0.4rem;
  font-size: 0.75rem;
  color: #6b7280;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.attributes-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}

.gsi-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.gsi-card {
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  padding: 0.6rem 0.75rem;
}

.gsi-name {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-weight: 600;
  font-size: 0.85rem;
  color: #374151;
  margin-bottom: 0.3rem;
}

.gsi-name i {
  font-size: 0.8rem;
  color: #3b82f6;
}

.gsi-details {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8rem;
  color: #6b7280;
}

/* Mobile: stack vertically */
@media (max-width: 768px) {
  .split-pane {
    grid-template-columns: 1fr;
  }

  .detail-meta-bar {
    gap: 0.75rem;
  }
}
</style>
