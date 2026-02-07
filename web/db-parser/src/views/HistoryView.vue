<script setup lang="ts">
import { ref, onMounted } from 'vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import Tag from 'primevue/tag'
import Tabs from 'primevue/tabs'
import TabList from 'primevue/tablist'
import Tab from 'primevue/tab'
import TabPanels from 'primevue/tabpanels'
import TabPanel from 'primevue/tabpanel'
import { useToast } from 'primevue/usetoast'
import { getSchemas, getSchemaById } from '@/services/schemasApi'
import type { Conversion, ConversionSummary } from '@/models/schemas'

const toast = useToast()
const conversions = ref<ConversionSummary[]>([])
const loading = ref(false)
const loadingDetail = ref(false)
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

async function openDetail(summary: ConversionSummary) {
  modalVisible.value = true
  loadingDetail.value = true
  selectedConversion.value = null
  try {
    selectedConversion.value = await getSchemaById(summary.conversionId)
  } catch {
    toast.add({ severity: 'error', summary: 'Error', detail: 'No se pudo cargar el detalle', life: 3000 })
    modalVisible.value = false
  } finally {
    loadingDetail.value = false
  }
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
      <!-- Loading state -->
      <div v-if="loadingDetail" class="detail-loading">
        <i class="pi pi-spin pi-spinner" style="font-size: 2rem; color: #3b82f6;"></i>
        <p>Cargando detalle...</p>
      </div>

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
            <Tag :value="selectedConversion.status" :severity="statusSeverity(selectedConversion.status)" />
          </div>
          <div class="meta-item">
            <Tag :value="selectedConversion.optimizationType" severity="info" />
          </div>
          <div class="meta-item">
            <span class="meta-label">Tablas Extraidas:</span>
            <span class="meta-value">{{ selectedConversion.tablesExtracted }}</span>
          </div>
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
              <!-- New DynamoDBDesign format -->
              <template v-if="selectedConversion.noSqlSchema.design">
                <Tabs value="0">
                  <TabList>
                    <Tab value="0"><i class="pi pi-server tab-icon"></i> Estructura</Tab>
                    <Tab value="1" v-if="selectedConversion.noSqlSchema.analysis?.accessPatterns?.length"><i class="pi pi-directions tab-icon"></i> Access Patterns</Tab>
                    <Tab value="2" v-if="selectedConversion.noSqlSchema.sampleData?.length"><i class="pi pi-database tab-icon"></i> Sample Data</Tab>
                  </TabList>
                  <TabPanels>
                    <!-- Tab: Estructura -->
                    <TabPanel value="0">
                      <div class="table-card">
                        <h5 class="table-name">
                          <i class="pi pi-table"></i>
                          {{ selectedConversion.noSqlSchema.design.tableName }}
                        </h5>

                        <h6>Billing Mode</h6>
                        <div class="billing-badge">
                          <i class="pi pi-money-bill"></i>
                          {{ selectedConversion.noSqlSchema.design.billingMode.replace(/_/g, ' ') }}
                        </div>

                        <h6>Primary Key</h6>
                        <div class="key-section">
                          <div class="key-item">
                            <span class="key-badge pk">PK</span>
                            <span>{{ selectedConversion.noSqlSchema.design.primaryKey.partitionKey.name }} ({{ selectedConversion.noSqlSchema.design.primaryKey.partitionKey.type }})</span>
                          </div>
                          <div class="key-item">
                            <span class="key-badge sk">SK</span>
                            <span>{{ selectedConversion.noSqlSchema.design.primaryKey.sortKey.name }} ({{ selectedConversion.noSqlSchema.design.primaryKey.sortKey.type }})</span>
                          </div>
                        </div>

                        <template v-if="selectedConversion.noSqlSchema.design.globalSecondaryIndexes.length > 0">
                          <h6>Global Secondary Indexes</h6>
                          <div class="gsi-list">
                            <div v-for="gsi in selectedConversion.noSqlSchema.design.globalSecondaryIndexes" :key="gsi.indexName" class="gsi-card">
                              <div class="gsi-name">
                                <i class="pi pi-search"></i>
                                {{ gsi.indexName }}
                              </div>
                              <div class="gsi-details">
                                <span class="key-badge pk">{{ gsi.partitionKey.name }}</span>
                                <span v-if="gsi.sortKey" class="key-badge sk">{{ gsi.sortKey.name }}</span>
                                <Tag :value="gsi.projection" severity="info" size="small" />
                                <span class="gsi-purpose">{{ gsi.purpose }}</span>
                              </div>
                            </div>
                          </div>
                        </template>

                        <template v-if="selectedConversion.noSqlSchema.design.entitySchemas.length > 0">
                          <h6>Entity Schemas</h6>
                          <div class="entity-list">
                            <div v-for="entity in selectedConversion.noSqlSchema.design.entitySchemas" :key="entity.entityType" class="entity-card">
                              <div class="entity-name">
                                <i class="pi pi-box"></i>
                                {{ entity.entityType }}
                              </div>
                              <div class="entity-patterns">
                                <span class="pattern-badge pk">PK: {{ entity.pkPattern }}</span>
                                <span class="pattern-badge sk">SK: {{ entity.skPattern }}</span>
                              </div>
                              <div class="attributes-list">
                                <Tag
                                  v-for="attr in entity.attributes"
                                  :key="attr.name"
                                  :value="`${attr.name} (${attr.type})${attr.required ? ' *' : ''}`"
                                  :severity="attr.required ? 'warn' : 'secondary'"
                                  size="small"
                                />
                              </div>
                            </div>
                          </div>
                        </template>

                        <template v-if="selectedConversion.noSqlSchema.design.edgeItems?.length > 0">
                          <h6>Edge Items</h6>
                          <div class="entity-list">
                            <div v-for="edge in selectedConversion.noSqlSchema.design.edgeItems" :key="edge.name" class="entity-card edge-card">
                              <div class="entity-name">
                                <i class="pi pi-link"></i>
                                {{ edge.name }}
                              </div>
                              <p class="edge-description" v-if="edge.description">{{ edge.description }}</p>
                              <div class="entity-patterns">
                                <span class="pattern-badge pk">PK: {{ edge.pkPattern }}</span>
                                <span class="pattern-badge sk">SK: {{ edge.skPattern }}</span>
                              </div>
                              <div class="attributes-list" v-if="edge.attributes?.length">
                                <Tag
                                  v-for="attr in edge.attributes"
                                  :key="attr.name"
                                  :value="`${attr.name} (${attr.type})`"
                                  severity="secondary"
                                  size="small"
                                />
                              </div>
                            </div>
                          </div>
                        </template>
                      </div>
                    </TabPanel>

                    <!-- Tab: Access Patterns -->
                    <TabPanel value="1">
                      <div class="ap-section">
                        <div v-for="ap in selectedConversion.noSqlSchema.analysis?.accessPatterns" :key="ap.id" class="ap-card">
                          <div class="ap-header">
                            <span class="ap-id">{{ ap.id }}</span>
                            <span class="ap-operation">
                              <Tag :value="ap.operation" severity="info" size="small" />
                            </span>
                            <span class="ap-index">
                              <Tag :value="ap.index" severity="secondary" size="small" />
                            </span>
                          </div>
                          <p class="ap-description">{{ ap.description }}</p>
                          <code class="ap-key-condition">{{ ap.keyCondition }}</code>
                        </div>
                      </div>

                      <template v-if="selectedConversion.noSqlSchema.accessPatternImplementation?.length > 0">
                        <h6>Implementaciones</h6>
                        <div class="ap-section">
                          <div v-for="impl in selectedConversion.noSqlSchema.accessPatternImplementation" :key="impl.patternId" class="impl-card">
                            <div class="ap-header">
                              <span class="ap-id">{{ impl.patternId }}</span>
                              <span class="ap-desc-text">{{ impl.description }}</span>
                            </div>
                            <pre class="impl-json">{{ JSON.stringify(impl.implementation, null, 2) }}</pre>
                          </div>
                        </div>
                      </template>
                    </TabPanel>

                    <!-- Tab: Sample Data -->
                    <TabPanel value="2">
                      <div class="sample-section">
                        <div v-for="(sample, idx) in selectedConversion.noSqlSchema.sampleData" :key="idx" class="sample-card">
                          <div class="sample-header">
                            <i class="pi pi-file"></i>
                            <span>{{ sample.description }}</span>
                          </div>
                          <pre class="sample-json" v-if="sample.item">{{ JSON.stringify(sample.item, null, 2) }}</pre>
                          <p v-else class="sample-empty">Sin datos de ejemplo</p>
                        </div>
                      </div>
                    </TabPanel>
                  </TabPanels>
                </Tabs>
              </template>

              <!-- Legacy format fallback -->
              <template v-else>
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
                          <Tag :value="gsi.projectionType || gsi.projection" severity="info" size="small" />
                        </div>
                      </div>
                    </div>
                  </template>
                </div>
              </template>
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

/* Detail loading */
.detail-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem;
  gap: 1rem;
  color: #6b7280;
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

.gsi-purpose {
  font-style: italic;
  color: #9ca3af;
  font-size: 0.75rem;
}

/* Entity schemas */
.entity-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.entity-card {
  background: #f0f9ff;
  border: 1px solid #bae6fd;
  border-radius: 6px;
  padding: 0.6rem 0.75rem;
}

.entity-name {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-weight: 600;
  font-size: 0.85rem;
  color: #0c4a6e;
  margin-bottom: 0.3rem;
}

.entity-name i {
  font-size: 0.8rem;
  color: #0284c7;
}

.entity-patterns {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 0.4rem;
}

.pattern-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 600;
  font-family: monospace;
}

.pattern-badge.pk {
  background: #dbeafe;
  color: #1e40af;
}

.pattern-badge.sk {
  background: #fef3c7;
  color: #92400e;
}

/* Recommendations */
.recommendations-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.recommendation-card {
  background: #fffbeb;
  border: 1px solid #fed7aa;
  border-radius: 6px;
  padding: 0.75rem;
}

.rec-header {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  margin-bottom: 0.4rem;
}

.rec-header i {
  color: #f59e0b;
  font-size: 0.9rem;
}

.rec-category {
  font-weight: 600;
  font-size: 0.85rem;
  color: #92400e;
}

.rec-description {
  margin: 0 0 0.3rem;
  font-size: 0.85rem;
  color: #374151;
}

.rec-rationale {
  margin: 0;
  font-size: 0.8rem;
  color: #6b7280;
  font-style: italic;
}

/* Tabs */
.tab-icon {
  margin-right: 0.3rem;
  font-size: 0.8rem;
}

/* Access Patterns */
.ap-section {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.ap-card {
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  border-radius: 6px;
  padding: 0.75rem;
}

.ap-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.3rem;
}

.ap-id {
  font-weight: 700;
  font-size: 0.8rem;
  color: #166534;
  background: #dcfce7;
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
  font-family: monospace;
}

.ap-description {
  margin: 0 0 0.3rem;
  font-size: 0.85rem;
  color: #374151;
}

.ap-key-condition {
  display: block;
  font-size: 0.78rem;
  color: #6b7280;
  background: #f9fafb;
  padding: 0.3rem 0.5rem;
  border-radius: 4px;
  border: 1px solid #e5e7eb;
  font-family: monospace;
  word-break: break-all;
}

.ap-desc-text {
  font-size: 0.8rem;
  color: #374151;
}

/* Implementations */
.impl-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  padding: 0.75rem;
}

.impl-json {
  margin: 0.4rem 0 0;
  font-size: 0.75rem;
  color: #334155;
  background: #1e1e2e;
  color: #a6e3a1;
  padding: 0.5rem;
  border-radius: 4px;
  overflow: auto;
  max-height: 200px;
}

/* Sample Data */
.sample-section {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.sample-card {
  background: #faf5ff;
  border: 1px solid #e9d5ff;
  border-radius: 6px;
  padding: 0.75rem;
}

.sample-header {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-weight: 600;
  font-size: 0.85rem;
  color: #6b21a8;
  margin-bottom: 0.4rem;
}

.sample-header i {
  font-size: 0.8rem;
  color: #9333ea;
}

.sample-json {
  margin: 0;
  font-size: 0.75rem;
  background: #1e1e2e;
  color: #cba6f7;
  padding: 0.5rem;
  border-radius: 4px;
  overflow: auto;
  max-height: 250px;
}

.sample-empty {
  margin: 0;
  font-size: 0.8rem;
  color: #9ca3af;
  font-style: italic;
}

/* Edge Items */
.edge-card {
  background: #fefce8 !important;
  border-color: #fde68a !important;
}

.edge-description {
  margin: 0 0 0.3rem;
  font-size: 0.8rem;
  color: #6b7280;
  font-style: italic;
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
