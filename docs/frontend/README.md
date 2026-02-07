# Frontend - Vue 3 SPA

Documentación del frontend desarrollado con Vue 3, TypeScript y PrimeVue.

## Tabla de Contenidos

1. [Stack Tecnológico](#stack-tecnológico)
2. [Estructura del Proyecto](#estructura-del-proyecto)
3. [Vistas](#vistas)
4. [Componentes](#componentes)
5. [Servicios API](#servicios-api)
6. [Modelos TypeScript](#modelos-typescript)
7. [Routing](#routing)
8. [Estilos](#estilos)
9. [Build y Deploy](#build-y-deploy)

---

## Stack Tecnológico

| Tecnología | Versión | Propósito |
|------------|---------|-----------|
| Vue | 3.5.27 | Framework reactivo |
| Vue Router | 4.6.4 | Routing SPA |
| Pinia | 3.0.4 | State management |
| PrimeVue | 4.5.4 | Componentes UI |
| PrimeIcons | 7.0.0 | Iconografía |
| TypeScript | ~5.9.3 | Tipado estático |
| Vite | 7.3.1 | Build tool |
| ESLint + OxLint | Latest | Linting |
| Prettier | Latest | Formateo |

---

## Estructura del Proyecto

```
web/db-parser/
├── public/                    # Assets estáticos
│   └── favicon.ico
│
├── src/
│   ├── App.vue                # Componente raíz
│   ├── main.ts                # Bootstrap de la app
│   │
│   ├── views/                 # Páginas/Vistas
│   │   ├── HomeView.vue       # Parser SQL
│   │   └── HistoryView.vue    # Historial de conversiones
│   │
│   ├── components/            # Componentes reutilizables
│   │   └── Sidebar.vue        # Navegación lateral
│   │
│   ├── services/              # Clientes API
│   │   └── schemasApi.ts      # Endpoints de schemas
│   │
│   ├── models/                # Interfaces TypeScript
│   │   └── schemas.ts         # Tipos de datos
│   │
│   ├── router/                # Configuración de rutas
│   │   └── index.ts
│   │
│   ├── stores/                # Pinia stores
│   │   └── counter.ts         # (No usado actualmente)
│   │
│   └── assets/                # Assets procesados
│       └── styles/
│
├── index.html                 # HTML principal
├── vite.config.ts             # Configuración Vite
├── tsconfig.json              # Configuración TypeScript
├── package.json               # Dependencias
└── .env                       # Variables de entorno
```

---

## Vistas

### HomeView - Parser SQL

**Archivo:** `src/views/HomeView.vue`
**Ruta:** `/`

#### Interfaz

```
┌────────────────────────────────────────────────────────────┐
│                     SQL to DynamoDB Parser                  │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  ┌──────────────────────────────────────────────────────┐  │
│  │                                                      │  │
│  │  Textarea para pegar SQL                             │  │
│  │  (CREATE TABLE statements)                           │  │
│  │                                                      │  │
│  │                                                      │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                            │
│  Tipo de Optimización:  [Dropdown ▼]                       │
│  - Balanced (Recommended)                                  │
│  - Read-Heavy                                              │
│  - Write-Heavy                                             │
│  - Cost-Optimized                                          │
│                                                            │
│                           [Parsear]                        │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

#### Código Clave

```vue
<script setup lang="ts">
import { ref } from 'vue'
import Textarea from 'primevue/textarea'
import Button from 'primevue/button'
import Select from 'primevue/select'
import Dialog from 'primevue/dialog'
import { parseSql, type ParseSqlRequest } from '@/services/schemasApi'

// State
const sqlInput = ref('')
const loading = ref(false)
const successModalVisible = ref(false)
const errorModalVisible = ref(false)
const errorMessage = ref('')
const errorDetails = ref<ValidationDetail[]>([])
const selectedOptimization = ref({
  label: 'Balanced (Recommended)',
  value: 'balanced'
})

const optimizationOptions = [
  { label: 'Balanced (Recommended)', value: 'balanced' },
  { label: 'Read-Heavy', value: 'read_heavy' },
  { label: 'Write-Heavy', value: 'write_heavy' },
  { label: 'Cost-Optimized', value: 'cost_optimized' }
]

// Handlers
const handleParse = async () => {
  if (!sqlInput.value.trim()) {
    errorMessage.value = 'Por favor, ingresa SQL para parsear'
    errorModalVisible.value = true
    return
  }

  loading.value = true

  try {
    const request: ParseSqlRequest = {
      sqlContent: sqlInput.value,
      optimizationType: selectedOptimization.value.value
    }

    const response = await parseSql(request)

    successModalVisible.value = true
    sqlInput.value = ''

  } catch (error: any) {
    if (error.details) {
      errorDetails.value = error.details
    }
    errorMessage.value = error.message || 'Error al procesar SQL'
    errorModalVisible.value = true
  } finally {
    loading.value = false
  }
}
</script>
```

#### Validaciones Frontend

- SQL no vacío
- Tipo de optimización válido
- Manejo de errores de API
- Feedback visual durante carga

---

### HistoryView - Historial

**Archivo:** `src/views/HistoryView.vue`
**Ruta:** `/history`

#### Interfaz

```
┌────────────────────────────────────────────────────────────────────────┐
│                         Historial de Conversiones                       │
├────────────────────────────────────────────────────────────────────────┤
│                                                                        │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ Fecha     │ Hora    │ Estado     │ Optimización │ Tablas         │  │
│  ├───────────┼─────────┼────────────┼──────────────┼────────────────┤  │
│  │ 2024-01-15│ 14:32   │ COMPLETED  │ balanced     │ 3              │  │
│  │ 2024-01-15│ 14:28   │ PROCESSING │ read_heavy   │ 2              │  │
│  │ 2024-01-15│ 14:25   │ FAILED     │ balanced     │ 1              │  │
│  └───────────┴─────────┴────────────┴──────────────┴────────────────┘  │
│                                                                        │
│  Página: [1] [2] [3]                                      10 por página│
│                                                                        │
└────────────────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────────────────┐
│                         Modal: Detalle de Conversión                   │
├───────────────────────────────┬────────────────────────────────────────┤
│                               │                                        │
│  SQL Original:                │  [Estructura] [Access Patterns] [Data] │
│  ┌─────────────────────────┐  │  ┌──────────────────────────────────┐  │
│  │ CREATE TABLE users (    │  │  │ Table: MyAppTable                │  │
│  │   id SERIAL PRIMARY..   │  │  │ Billing: PAY_PER_REQUEST         │  │
│  │   email VARCHAR(255)..  │  │  │                                  │  │
│  │   ...                   │  │  │ Primary Key:                     │  │
│  │ );                      │  │  │   PK: pk (S)                     │  │
│  │                         │  │  │   SK: sk (S)                     │  │
│  │ CREATE TABLE orders (   │  │  │                                  │  │
│  │   ...                   │  │  │ GSI:                             │  │
│  │ );                      │  │  │ - GSI1 (gsi1pk, gsi1sk)          │  │
│  └─────────────────────────┘  │  │ - GSI2 (gsi2pk, gsi2sk)          │  │
│                               │  └──────────────────────────────────┘  │
└───────────────────────────────┴────────────────────────────────────────┘
```

#### Código Clave

```vue
<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import Tabs from 'primevue/tabs'
import TabList from 'primevue/tablist'
import Tab from 'primevue/tab'
import TabPanels from 'primevue/tabpanels'
import TabPanel from 'primevue/tabpanel'
import Tag from 'primevue/tag'
import { getSchemas, getSchemaById, type Conversion } from '@/services/schemasApi'

// State
const conversions = ref<Conversion[]>([])
const loading = ref(true)
const selectedConversion = ref<Conversion | null>(null)
const detailModalVisible = ref(false)
const activeTab = ref('structure')

// Lifecycle
onMounted(async () => {
  await loadConversions()
})

// Methods
const loadConversions = async () => {
  loading.value = true
  try {
    const response = await getSchemas()
    conversions.value = response.conversions
  } catch (error) {
    console.error('Failed to load conversions:', error)
  } finally {
    loading.value = false
  }
}

const showDetails = async (conversion: Conversion) => {
  try {
    const fullData = await getSchemaById(conversion.conversionId)
    selectedConversion.value = fullData
    detailModalVisible.value = true
  } catch (error) {
    console.error('Failed to load details:', error)
  }
}

// Computed
const statusSeverity = computed(() => (status: string) => {
  const map: Record<string, string> = {
    'PENDING': 'warning',
    'PROCESSING': 'info',
    'DESIGN_COMPLETED': 'info',
    'PROCESSING_PATTERNS': 'info',
    'COMPLETED': 'success',
    'FAILED': 'danger'
  }
  return map[status] || 'secondary'
})
</script>

<template>
  <DataTable
    :value="conversions"
    :loading="loading"
    paginator
    :rows="10"
    @row-click="showDetails($event.data)"
  >
    <Column field="conversionDate" header="Fecha" sortable />
    <Column field="createdAt" header="Hora">
      <template #body="{ data }">
        {{ formatTime(data.createdAt) }}
      </template>
    </Column>
    <Column field="status" header="Estado">
      <template #body="{ data }">
        <Tag :severity="statusSeverity(data.status)" :value="data.status" />
      </template>
    </Column>
    <Column field="optimizationType" header="Optimización" />
    <Column field="tablesExtracted" header="Tablas" />
  </DataTable>

  <Dialog v-model:visible="detailModalVisible" modal maximizable>
    <template #header>
      Conversión: {{ selectedConversion?.conversionId }}
    </template>

    <div class="detail-container">
      <!-- SQL Original -->
      <div class="sql-panel">
        <h4>SQL Original</h4>
        <pre>{{ selectedConversion?.sqlContent }}</pre>
      </div>

      <!-- DynamoDB Design -->
      <div class="design-panel">
        <Tabs v-model:value="activeTab">
          <TabList>
            <Tab value="structure">Estructura</Tab>
            <Tab value="patterns">Access Patterns</Tab>
            <Tab value="data">Sample Data</Tab>
          </TabList>
          <TabPanels>
            <TabPanel value="structure">
              <StructureTab :design="selectedConversion?.noSqlSchema?.design" />
            </TabPanel>
            <TabPanel value="patterns">
              <PatternsTab :patterns="selectedConversion?.noSqlSchema?.accessPatternImplementation" />
            </TabPanel>
            <TabPanel value="data">
              <DataTab :samples="selectedConversion?.noSqlSchema?.sampleData" />
            </TabPanel>
          </TabPanels>
        </Tabs>
      </div>
    </div>
  </Dialog>
</template>
```

#### Tabs de Detalle

**Tab 1: Estructura**
- Nombre de tabla DynamoDB
- Billing mode
- Primary Key (PK y SK)
- Global Secondary Indexes
- Entity Schemas
- Edge Items

**Tab 2: Access Patterns**
- Lista de patrones
- ID del patrón
- Descripción
- Operación (GET, QUERY, SCAN)
- Índice utilizado
- Key condition expression

**Tab 3: Sample Data**
- Items de ejemplo en formato JSON
- Visualización de cómo se verían los datos

---

## Componentes

### Sidebar

**Archivo:** `src/components/Sidebar.vue`

#### Comportamiento

- Colapsable (70px → 220px)
- Hover para expandir
- Responsive (drawer en mobile)
- Indicador de ruta activa

#### Código

```vue
<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const isExpanded = ref(false)

const menuItems = [
  { label: 'Parser', icon: 'pi pi-code', path: '/' },
  { label: 'Historial', icon: 'pi pi-history', path: '/history' }
]

const isActive = (path: string) => route.path === path

const navigate = (path: string) => {
  router.push(path)
}
</script>

<template>
  <aside
    class="sidebar"
    :class="{ expanded: isExpanded }"
    @mouseenter="isExpanded = true"
    @mouseleave="isExpanded = false"
  >
    <div class="logo">
      <i class="pi pi-database"></i>
      <span v-show="isExpanded">DB Parser</span>
    </div>

    <nav class="menu">
      <div
        v-for="item in menuItems"
        :key="item.path"
        class="menu-item"
        :class="{ active: isActive(item.path) }"
        @click="navigate(item.path)"
      >
        <i :class="item.icon"></i>
        <span v-show="isExpanded">{{ item.label }}</span>
      </div>
    </nav>
  </aside>
</template>

<style scoped>
.sidebar {
  position: fixed;
  left: 0;
  top: 0;
  height: 100vh;
  width: 70px;
  background: var(--surface-card);
  transition: width 0.3s ease;
  z-index: 100;
}

.sidebar.expanded {
  width: 220px;
}

.menu-item {
  padding: 1rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 1rem;
}

.menu-item.active {
  background: var(--primary-color);
  color: white;
}

.menu-item:hover:not(.active) {
  background: var(--surface-hover);
}
</style>
```

---

## Servicios API

**Archivo:** `src/services/schemasApi.ts`

### Configuración Base

```typescript
const API_BASE_URL = import.meta.env.VITE_BASE_PATH_URL || ''
const ENDPOINT_URL = import.meta.env.VITE_ENDPOINT_URL || 'prod/api/v1/schemas'

const buildUrl = (path: string = ''): string => {
  const base = `${API_BASE_URL}/${ENDPOINT_URL}`.replace(/\/+/g, '/')
  return path ? `${base}/${path}` : base
}
```

### Endpoints

```typescript
// Listar todas las conversiones
export const getSchemas = async (): Promise<GetSchemasResponse> => {
  const response = await fetch(buildUrl(), {
    method: 'GET',
    headers: { 'Content-Type': 'application/json' }
  })

  if (!response.ok) {
    throw new Error('Failed to fetch schemas')
  }

  return response.json()
}

// Iniciar una conversión
export const parseSql = async (request: ParseSqlRequest): Promise<Conversion> => {
  const response = await fetch(buildUrl(), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(request)
  })

  if (!response.ok) {
    const error = await response.json()
    throw error
  }

  return response.json()
}

// Obtener conversión por ID
export const getSchemaById = async (conversionId: string): Promise<Conversion> => {
  const response = await fetch(buildUrl(conversionId), {
    method: 'GET',
    headers: { 'Content-Type': 'application/json' }
  })

  if (!response.ok) {
    throw new Error('Failed to fetch schema')
  }

  return response.json()
}
```

---

## Modelos TypeScript

**Archivo:** `src/models/schemas.ts`

```typescript
// Request para parsear SQL
export interface ParseSqlRequest {
  sqlContent: string
  optimizationType?: 'balanced' | 'read_heavy' | 'write_heavy' | 'cost_optimized'
}

// Respuesta de lista
export interface GetSchemasResponse {
  conversions: Conversion[]
  count: number
}

// Conversión individual
export interface Conversion {
  conversionId: string
  conversionDate: string
  createdAt: string
  expiresAt: string
  sqlContent: string
  noSqlSchema: DynamoDBDesign | null
  optimizationType: string
  status: ConversionStatus
  tablesExtracted: string
}

// Estados posibles
export type ConversionStatus =
  | 'PENDING'
  | 'PROCESSING'
  | 'DESIGN_COMPLETED'
  | 'PROCESSING_PATTERNS'
  | 'COMPLETED'
  | 'FAILED'

// Diseño DynamoDB
export interface DynamoDBDesign {
  analysis: Analysis
  design: Design
  sampleData: SampleDataItem[]
  accessPatternImplementation: AccessPatternImpl[]
}

export interface Analysis {
  entities: Entity[]
  accessPatterns: AccessPattern[]
}

export interface Entity {
  name: string
  type: string
  attributes: string[]
}

export interface AccessPattern {
  id: string
  description: string
  frequency: string
}

export interface Design {
  tableName: string
  billingMode: string
  primaryKey: PrimaryKey
  globalSecondaryIndexes: GlobalSecondaryIndex[]
  entitySchemas: EntitySchema[]
  edgeItems?: EdgeItemSchema[]
}

export interface PrimaryKey {
  partitionKey: KeyDefinition
  sortKey?: KeyDefinition
}

export interface KeyDefinition {
  name: string
  type: 'S' | 'N' | 'B'
}

export interface GlobalSecondaryIndex {
  indexName: string
  partitionKey: KeyDefinition
  sortKey?: KeyDefinition
  projection: 'ALL' | 'KEYS_ONLY' | string[]
  purpose: string
}

export interface EntitySchema {
  entityType: string
  pkPattern: string
  skPattern: string
  attributes: AttributeDefinition[]
}

export interface AttributeDefinition {
  name: string
  type: string
  required: boolean
  source?: string
}

export interface EdgeItemSchema {
  name: string
  pkPattern: string
  skPattern: string
  attributes: AttributeDefinition[]
}

export interface SampleDataItem {
  pk: string
  sk: string
  [key: string]: any
}

export interface AccessPatternImpl {
  patternId: string
  description: string
  operation: 'GetItem' | 'Query' | 'Scan'
  index: 'PRIMARY' | string
  keyCondition: string
  implementation?: string
}
```

---

## Routing

**Archivo:** `src/router/index.ts`

```typescript
import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '@/views/HomeView.vue'
import HistoryView from '@/views/HistoryView.vue'

const routes = [
  {
    path: '/',
    name: 'home',
    component: HomeView,
    meta: { title: 'Parser' }
  },
  {
    path: '/history',
    name: 'history',
    component: HistoryView,
    meta: { title: 'Historial' }
  },
  {
    // Catch-all redirect
    path: '/:pathMatch(.*)*',
    redirect: '/'
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

// Update document title
router.afterEach((to) => {
  document.title = `${to.meta.title} | DB Parser`
})

export default router
```

---

## Estilos

### Tema PrimeVue

```typescript
// main.ts
import { createApp } from 'vue'
import PrimeVue from 'primevue/config'
import Aura from '@primevue/themes/aura'

const app = createApp(App)

app.use(PrimeVue, {
  theme: {
    preset: Aura,
    options: {
      darkModeSelector: '.dark-mode',
      cssLayer: {
        name: 'primevue',
        order: 'tailwind-base, primevue, tailwind-utilities'
      }
    }
  }
})
```

### Variables CSS

```css
/* src/assets/styles/main.css */
:root {
  --sidebar-width: 70px;
  --sidebar-expanded: 220px;
  --header-height: 60px;
}

.dark-mode {
  --surface-card: #1e1e1e;
  --surface-hover: #2d2d2d;
  --text-color: #ffffff;
}
```

### Layout Principal

```vue
<!-- App.vue -->
<template>
  <div class="app-container">
    <Sidebar />
    <main class="main-content">
      <RouterView />
    </main>
  </div>
</template>

<style>
.app-container {
  display: flex;
  min-height: 100vh;
}

.main-content {
  flex: 1;
  margin-left: var(--sidebar-width);
  padding: 2rem;
  transition: margin-left 0.3s ease;
}
</style>
```

---

## Build y Deploy

### Scripts NPM

```json
{
  "scripts": {
    "dev": "vite",
    "build": "vue-tsc && vite build",
    "preview": "vite preview",
    "lint": "eslint . --ext .vue,.ts --fix && oxlint",
    "format": "prettier --write src/"
  }
}
```

### Configuración Vite

```typescript
// vite.config.ts
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
    minify: 'terser',
    rollupOptions: {
      output: {
        manualChunks: {
          'primevue': ['primevue'],
          'vendor': ['vue', 'vue-router', 'pinia']
        }
      }
    }
  }
})
```

### Variables de Entorno

```bash
# .env.production
VITE_BASE_PATH_URL=https://app-sql.cloudcentinel.com
VITE_ENDPOINT_URL=prod/api/v1/schemas

# .env.development
VITE_BASE_PATH_URL=http://localhost:4566
VITE_ENDPOINT_URL=restapis/xxx/dev/_user_request_/api/v1/schemas
```

### Deploy

```bash
# Build
cd web/db-parser
npm install
npm run build

# Deploy a S3
aws s3 sync dist/ s3://sql-to-nosql-frontend/ --delete

# Invalidar cache CloudFront
aws cloudfront create-invalidation \
  --distribution-id EXXXXX \
  --paths "/*"
```

### Makefile Targets

```makefile
frontend:
	cd web/db-parser && npm install && npm run build

deploy-frontend: frontend
	aws s3 sync web/db-parser/dist/ s3://$(FRONTEND_BUCKET)/ --delete
	aws cloudfront create-invalidation --distribution-id $(CF_DISTRIBUTION_ID) --paths "/*"
```
