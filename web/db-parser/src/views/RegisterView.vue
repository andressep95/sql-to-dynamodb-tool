<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Button from 'primevue/button'
import Message from 'primevue/message'
import { useToast } from 'primevue/usetoast'
import { getInvitation, register, type InvitationResponse } from '@/services/adminApi'

const route = useRoute()
const router = useRouter()
const toast = useToast()

const invitationCode = ref('')
const invitation = ref<InvitationResponse | null>(null)
const loading = ref(false)
const validating = ref(false)
const errorMessage = ref('')

const email = ref('')
const password = ref('')
const confirmPassword = ref('')

onMounted(async () => {
  const code = route.query.code as string
  if (code) {
    invitationCode.value = code
    await validateInvitation()
  }
})

async function validateInvitation() {
  if (!invitationCode.value || invitationCode.value.length !== 6) {
    errorMessage.value = 'El código debe tener 6 dígitos'
    return
  }

  validating.value = true
  errorMessage.value = ''

  try {
    invitation.value = await getInvitation(invitationCode.value)
  } catch (error: any) {
    errorMessage.value = error.message || 'Código de invitación inválido o expirado'
    invitation.value = null
  } finally {
    validating.value = false
  }
}

async function handleRegister() {
  if (!invitation.value) {
    toast.add({ severity: 'error', summary: 'Error', detail: 'Primero valida el código de invitación', life: 3000 })
    return
  }

  if (password.value !== confirmPassword.value) {
    toast.add({ severity: 'error', summary: 'Error', detail: 'Las contraseñas no coinciden', life: 3000 })
    return
  }

  if (password.value.length < 8) {
    toast.add({ severity: 'error', summary: 'Error', detail: 'La contraseña debe tener al menos 8 caracteres', life: 3000 })
    return
  }

  loading.value = true

  try {
    await register({
      invitationCode: invitationCode.value,
      email: email.value,
      password: password.value,
    })
    toast.add({ severity: 'success', summary: 'Registro exitoso', detail: 'Ya puedes iniciar sesión', life: 3000 })
    setTimeout(() => router.push('/login'), 2000)
  } catch (error: any) {
    toast.add({ severity: 'error', summary: 'Error', detail: error.message, life: 5000 })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="register-view">
    <div class="register-container">
      <div class="register-header">
        <i class="pi pi-user-plus header-icon"></i>
        <h1>Registro con Invitación</h1>
        <p>Ingresa tu código de invitación para crear tu cuenta</p>
      </div>

      <div class="register-form">
        <!-- Paso 1: Validar código -->
        <div v-if="!invitation" class="invitation-step">
          <div class="field">
            <label for="code">Código de Invitación</label>
            <InputText
              id="code"
              v-model="invitationCode"
              placeholder="123456"
              maxlength="6"
              :fluid="true"
              @keyup.enter="validateInvitation"
            />
          </div>

          <Message v-if="errorMessage" severity="error" :closable="false">{{ errorMessage }}</Message>

          <Button
            label="Validar Código"
            icon="pi pi-check"
            :loading="validating"
            @click="validateInvitation"
            :fluid="true"
          />
        </div>

        <!-- Paso 2: Completar registro -->
        <div v-else class="registration-step">
          <Message severity="success" :closable="false">
            <div class="invitation-info">
              <p><strong>Invitación válida</strong></p>
              <p>Rol: {{ invitation.role }}</p>
              <p>Invitado por: {{ invitation.createdBy }}</p>
            </div>
          </Message>

          <div class="field">
            <label for="email">Email</label>
            <InputText id="email" v-model="email" type="email" :fluid="true" />
          </div>

          <div class="field">
            <label for="password">Contraseña</label>
            <Password id="password" v-model="password" toggleMask :fluid="true" />
          </div>

          <div class="field">
            <label for="confirmPassword">Confirmar Contraseña</label>
            <Password id="confirmPassword" v-model="confirmPassword" toggleMask :fluid="true" :feedback="false" />
          </div>

          <div class="button-group">
            <Button label="Volver" text @click="invitation = null" />
            <Button label="Registrarse" icon="pi pi-user-plus" :loading="loading" @click="handleRegister" />
          </div>
        </div>
      </div>

      <div class="register-footer">
        <p>¿Ya tienes cuenta? <router-link to="/login">Inicia sesión</router-link></p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.register-view {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 2rem;
}

.register-container {
  background: white;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  max-width: 450px;
  width: 100%;
  overflow: hidden;
}

.register-header {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: white;
  padding: 2rem;
  text-align: center;
}

.header-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.register-header h1 {
  margin: 0 0 0.5rem 0;
  font-size: 1.75rem;
}

.register-header p {
  margin: 0;
  opacity: 0.9;
  font-size: 0.95rem;
}

.register-form {
  padding: 2rem;
}

.invitation-step,
.registration-step {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
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

.invitation-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.invitation-info p {
  margin: 0;
  font-size: 0.875rem;
}

.button-group {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
}

.register-footer {
  background: #f9fafb;
  padding: 1.5rem;
  text-align: center;
  border-top: 1px solid #e5e7eb;
}

.register-footer p {
  margin: 0;
  color: #6b7280;
  font-size: 0.875rem;
}

.register-footer a {
  color: #3b82f6;
  text-decoration: none;
  font-weight: 600;
}

.register-footer a:hover {
  text-decoration: underline;
}
</style>
