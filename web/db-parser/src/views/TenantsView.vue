<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Dropdown from 'primevue/dropdown'
import { useToast } from 'primevue/usetoast'
import { getTenants, createTenant, getUsers, createUser, type Tenant, type CreateTenantRequest, type User, type CreateUserRequest } from '@/services/adminApi'

const toast = useToast()

const tenants = ref<Tenant[]>([])
const allUsers = ref<User[]>([])
const loading = ref(false)
const showTenantDialog = ref(false)
const showUsersDialog = ref(false)
const showCreateUserDialog = ref(false)
const selectedTenant = ref<Tenant | null>(null)

const newTenant = ref<CreateTenantRequest>({
  name: '',
  description: '',
})

const newUser = ref<CreateUserRequest>({
  email: '',
  password: '',
  role: 'USER_TENANT',
  tenantId: '',
})

const roles = [
  { label: 'Usuario', value: 'USER_TENANT' },
  { label: 'Supervisor', value: 'REALM_SUPERVISOR' },
  { label: 'Admin de Tenant', value: 'REALM_ADMIN' },
]

const tenantUsers = computed(() => {
  if (!selectedTenant.value) return []
  return allUsers.value.filter(u => u.tenantId === selectedTenant.value!.id)
})

onMounted(() => {
  loadTenants()
  loadUsers()
})

async function loadTenants() {
  loading.value = true
  try {
    const response = await getTenants()
    tenants.value = response.tenants
  } catch (error: any) {
    toast.add({ severity: 'error', summary: 'Error', detail: error.message, life: 3000 })
  } finally {
    loading.value = false
  }
}

async function loadUsers() {
  try {
    const response = await getUsers()
    allUsers.value = response.users
  } catch (error: any) {
    toast.add({ severity: 'error', summary: 'Error', detail: error.message, life: 3000 })
  }
}

function openCreateTenantDialog() {
  newTenant.value = { name: '', description: '' }
  showTenantDialog.value = true
}

async function handleCreateTenant() {
  try {
    await createTenant(newTenant.value)
    toast.add({ severity: 'success', summary: 'Tenant creado', detail: 'El tenant fue creado exitosamente', life: 3000 })
    showTenantDialog.value = false
    loadTenants()
  } catch (error: any) {
    toast.add({ severity: 'error', summary: 'Error', detail: error.message, life: 3000 })
  }
}

function openUsersDialog(tenant: Tenant) {
  selectedTenant.value = tenant
  showUsersDialog.value = true
}

function openCreateUserDialog() {
  newUser.value = {
    email: '',
    password: '',
    role: 'USER_TENANT',
    tenantId: selectedTenant.value!.id,
  }
  showCreateUserDialog.value = true
}

async function handleCreateUser() {
  try {
    await createUser(newUser.value)
    toast.add({ severity: 'success', summary: 'Usuario creado', detail: 'El usuario fue creado exitosamente', life: 3000 })
    showCreateUserDialog.value = false
    await loadUsers()
    await loadTenants()
  } catch (error: any) {
    toast.add({ severity: 'error', summary: 'Error', detail: error.message, life: 3000 })
  }
}
</script>

<template>
  <div class="tenants-view">
    <div class="header">
      <h1>Gestión de Tenants</h1>
      <Button label="Crear Tenant" icon="pi pi-plus" @click="openCreateTenantDialog" />
    </div>

    <DataTable :value="tenants" :loading="loading" paginator :rows="10" @row-click="openUsersDialog($event.data)" class="clickable-rows">
      <Column field="id" header="ID" sortable />
      <Column field="name" header="Nombre" sortable />
      <Column field="description" header="Descripción" />
      <Column field="userCount" header="Usuarios" sortable />
    </DataTable>

    <!-- Dialog Crear Tenant -->
    <Dialog v-model:visible="showTenantDialog" header="Crear Tenant" :modal="true" dismissableMask :style="{ width: '450px' }">
      <div class="dialog-content">
        <div class="field">
          <label for="name">Nombre</label>
          <InputText id="name" v-model="newTenant.name" :fluid="true" />
        </div>
        <div class="field">
          <label for="description">Descripción</label>
          <InputText id="description" v-model="newTenant.description" :fluid="true" />
        </div>
      </div>
      <template #footer>
        <Button label="Cancelar" text @click="showTenantDialog = false" />
        <Button label="Crear" @click="handleCreateTenant" />
      </template>
    </Dialog>

    <!-- Dialog Usuarios del Tenant -->
    <Dialog v-model:visible="showUsersDialog" :header="`Usuarios de ${selectedTenant?.name}`" :modal="true" dismissableMask :style="{ width: '800px' }">
      <div class="users-section">
        <div class="users-header">
          <p class="tenant-info">{{ selectedTenant?.description }}</p>
          <Button label="Agregar Usuario" icon="pi pi-user-plus" size="small" @click="openCreateUserDialog" />
        </div>
        
        <DataTable :value="tenantUsers" :rows="5" paginator>
          <Column field="email" header="Email" sortable />
          <Column field="role" header="Rol" sortable />
          <Column field="enabled" header="Estado">
            <template #body="{ data }">
              <span :class="data.enabled ? 'status-active' : 'status-inactive'">
                {{ data.enabled ? 'Activo' : 'Inactivo' }}
              </span>
            </template>
          </Column>
        </DataTable>
      </div>
    </Dialog>

    <!-- Dialog Crear Usuario -->
    <Dialog v-model:visible="showCreateUserDialog" header="Crear Usuario" :modal="true" dismissableMask :style="{ width: '450px' }">
      <div class="dialog-content">
        <div class="field">
          <label for="email">Email</label>
          <InputText id="email" v-model="newUser.email" type="email" :fluid="true" />
        </div>
        <div class="field">
          <label for="password">Contraseña</label>
          <Password id="password" v-model="newUser.password" toggleMask :fluid="true" />
        </div>
        <div class="field">
          <label for="role">Rol</label>
          <Dropdown id="role" v-model="newUser.role" :options="roles" optionLabel="label" optionValue="value" :fluid="true" />
        </div>
      </div>
      <template #footer>
        <Button label="Cancelar" text @click="showCreateUserDialog = false" />
        <Button label="Crear" @click="handleCreateUser" />
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.tenants-view {
  padding: 2rem;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.header h1 {
  margin: 0;
  font-size: 1.75rem;
  color: #111827;
}

.dialog-content {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1rem 0;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.field label {
  font-size: 0.875rem;
  font-weight: 600;
  color: #374151;
}

.clickable-rows :deep(tbody tr) {
  cursor: pointer;
}

.clickable-rows :deep(tbody tr:hover) {
  background-color: #f3f4f6;
}

.users-section {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.users-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid #e5e7eb;
}

.tenant-info {
  margin: 0;
  color: #6b7280;
  font-size: 0.875rem;
}

.status-active {
  color: #059669;
  font-weight: 500;
}

.status-inactive {
  color: #dc2626;
  font-weight: 500;
}
</style>
