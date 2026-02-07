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
import { getTenants, createTenant, getUsers, createUser, createInvitation, type Tenant, type CreateTenantRequest, type User, type CreateUserRequest, type InvitationResponse } from '@/services/adminApi'

const toast = useToast()

const tenants = ref<Tenant[]>([])
const allUsers = ref<User[]>([])
const loading = ref(false)
const showTenantDialog = ref(false)
const showUsersDialog = ref(false)
const showCreateUserDialog = ref(false)
const showInvitationDialog = ref(false)
const selectedTenant = ref<Tenant | null>(null)
const invitation = ref<InvitationResponse | null>(null)
const selectedRole = ref('USER_TENANT')
const invitationEmail = ref('')

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

const invitationLink = computed(() => {
  if (!invitation.value) return ''
  return `${window.location.origin}/register?code=${invitation.value.code}`
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

async function openInvitationDialog() {
  selectedRole.value = 'USER_TENANT'
  invitationEmail.value = ''
  invitation.value = null
  showInvitationDialog.value = true
}

async function generateInvitation() {
  loading.value = true
  try {
    invitation.value = await createInvitation({
      tenantId: selectedTenant.value!.id,
      role: selectedRole.value,
      email: invitationEmail.value || undefined,
    })
    const emailMsg = invitationEmail.value ? ` y enviado a ${invitationEmail.value}` : ''
    toast.add({ severity: 'success', summary: 'Invitación creada', detail: `Código generado exitosamente${emailMsg}`, life: 3000 })
  } catch (error: any) {
    toast.add({ severity: 'error', summary: 'Error', detail: error.message, life: 3000 })
  } finally {
    loading.value = false
  }
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text)
  toast.add({ severity: 'success', summary: 'Copiado', detail: 'Código copiado al portapapeles', life: 2000 })
}

function copyLinkToClipboard() {
  navigator.clipboard.writeText(invitationLink.value)
  toast.add({ severity: 'success', summary: 'Copiado', detail: 'Enlace copiado al portapapeles', life: 2000 })
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
          <div class="header-actions">
            <Button label="Invitar" icon="pi pi-send" size="small" @click="openInvitationDialog" />
            <Button label="Agregar Usuario" icon="pi pi-user-plus" size="small" @click="openCreateUserDialog" />
          </div>
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

    <!-- Dialog Invitación -->
    <Dialog v-model:visible="showInvitationDialog" header="Generar Invitación" :modal="true" dismissableMask :style="{ width: '500px' }">
      <div class="dialog-content">
        <p class="invitation-description">
          Genera un código de invitación para que nuevos usuarios se registren en este tenant.
          El código expira en 7 días.
        </p>

        <div class="field">
          <label for="invitationRole">Rol del invitado</label>
          <Dropdown id="invitationRole" v-model="selectedRole" :options="roles" optionLabel="label" optionValue="value" :fluid="true" />
        </div>

        <div class="field">
          <label for="invitationEmail">Email (opcional)</label>
          <InputText 
            id="invitationEmail" 
            v-model="invitationEmail" 
            type="email"
            placeholder="usuario@ejemplo.com"
            :fluid="true"
          />
          <small class="field-hint">Si proporcionas un email, se enviará automáticamente el enlace de invitación</small>
        </div>

        <Button 
          v-if="!invitation" 
          label="Generar Código" 
          icon="pi pi-ticket" 
          :loading="loading" 
          @click="generateInvitation" 
          :fluid="true"
        />

        <div v-if="invitation" class="invitation-result">

          <div class="link-display">
            <span class="code-label">Enlace de Registro:</span>
            <div class="link-value">
              <InputText :value="invitationLink" readonly :fluid="true" />
              <Button icon="pi pi-copy" @click="copyLinkToClipboard" />
            </div>
          </div>

          <div class="invitation-details">
            <p><strong>Rol:</strong> {{ invitation.role }}</p>
            <p><strong>Expira:</strong> {{ new Date(invitation.expiresAt * 1000).toLocaleString() }}</p>
          </div>
        </div>
      </div>
      <template #footer>
        <Button label="Cerrar" @click="showInvitationDialog = false" />
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

.header-actions {
  display: flex;
  gap: 0.5rem;
}

.invitation-description {
  color: #6b7280;
  font-size: 0.875rem;
  margin: 0 0 1rem 0;
}

.invitation-result {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  margin-top: 1rem;
  padding: 1rem;
  background: #f9fafb;
  border-radius: 8px;
}

.code-display,
.link-display {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.code-label {
  font-size: 0.875rem;
  font-weight: 600;
  color: #374151;
}

.code-value {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: white;
  padding: 1rem;
  border-radius: 8px;
  border: 2px solid #3b82f6;
}

.code-text {
  font-size: 2rem;
  font-weight: bold;
  color: #3b82f6;
  letter-spacing: 0.5rem;
  flex: 1;
  text-align: center;
}

.link-value {
  display: flex;
  gap: 0.5rem;
}

.invitation-details {
  padding-top: 1rem;
  border-top: 1px solid #e5e7eb;
}

.invitation-details p {
  margin: 0.25rem 0;
  font-size: 0.875rem;
  color: #6b7280;
}

.field-hint {
  display: block;
  margin-top: 0.25rem;
  font-size: 0.75rem;
  color: #6b7280;
  font-style: italic;
}
</style>
