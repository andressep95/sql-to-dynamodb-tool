<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Dropdown from 'primevue/dropdown'
import Password from 'primevue/password'
import { useToast } from 'primevue/usetoast'
import { getUsers, createUser, type User, type CreateUserRequest } from '@/services/adminApi'

const auth = useAuthStore()
const toast = useToast()

const users = ref<User[]>([])
const loading = ref(false)
const showDialog = ref(false)

const newUser = ref<CreateUserRequest>({
  email: '',
  password: '',
  role: 'USER_TENANT',
  tenantId: auth.tenantId,
})

const roles = [
  { label: 'Usuario', value: 'USER_TENANT' },
  { label: 'Supervisor', value: 'REALM_SUPERVISOR' },
]

if (auth.isSuperAdmin) {
  roles.push({ label: 'Admin de Tenant', value: 'REALM_ADMIN' })
}

onMounted(() => {
  loadUsers()
})

async function loadUsers() {
  loading.value = true
  try {
    const response = await getUsers()
    users.value = response.users
  } catch (error: any) {
    toast.add({ severity: 'error', summary: 'Error', detail: error.message, life: 3000 })
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  newUser.value = {
    email: '',
    password: '',
    role: 'USER_TENANT',
    tenantId: auth.tenantId,
  }
  showDialog.value = true
}

async function handleCreateUser() {
  try {
    await createUser(newUser.value)
    toast.add({ severity: 'success', summary: 'Usuario creado', detail: 'El usuario fue creado exitosamente', life: 3000 })
    showDialog.value = false
    loadUsers()
  } catch (error: any) {
    toast.add({ severity: 'error', summary: 'Error', detail: error.message, life: 3000 })
  }
}
</script>

<template>
  <div class="users-view">
    <div class="header">
      <h1>Gestión de Usuarios</h1>
      <Button label="Crear Usuario" icon="pi pi-plus" @click="openCreateDialog" />
    </div>

    <DataTable :value="users" :loading="loading" paginator :rows="10">
      <Column field="email" header="Email" sortable />
      <Column field="role" header="Rol" sortable />
      <Column field="tenantId" header="Tenant" sortable v-if="auth.isSuperAdmin" />
      <Column header="Acciones">
        <template #body>
          <Button icon="pi pi-pencil" text rounded />
          <Button icon="pi pi-trash" text rounded />
        </template>
      </Column>
    </DataTable>

    <Dialog v-model:visible="showDialog" header="Crear Usuario" :modal="true" dismissableMask :style="{ width: '450px' }">
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
        <div class="field" v-if="auth.isSuperAdmin">
          <label for="tenantId">Tenant ID</label>
          <InputText id="tenantId" v-model="newUser.tenantId" :fluid="true" />
        </div>
      </div>
      <template #footer>
        <Button label="Cancelar" text @click="showDialog = false" />
        <Button label="Crear" @click="handleCreateUser" />
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.users-view {
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
</style>
