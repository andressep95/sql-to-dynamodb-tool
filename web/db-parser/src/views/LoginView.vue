<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useToast } from 'primevue/usetoast'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Button from 'primevue/button'
import Message from 'primevue/message'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const toast = useToast()
const auth = useAuthStore()

type FormState = 'login' | 'register' | 'confirm'
const formState = ref<FormState>('login')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const confirmCode = ref('')
const errorMessage = ref('')

async function handleLogin() {
  errorMessage.value = ''
  try {
    await auth.login(email.value, password.value)
    router.push('/')
  } catch (err: any) {
    errorMessage.value = err.message || 'Error al iniciar sesión'
  }
}

async function handleRegister() {
  errorMessage.value = ''
  if (password.value !== confirmPassword.value) {
    errorMessage.value = 'Las contraseñas no coinciden'
    return
  }
  try {
    await auth.register(email.value, password.value)
    toast.add({ severity: 'success', summary: 'Registro exitoso', detail: 'Revisa tu email para el código de verificación', life: 5000 })
    formState.value = 'confirm'
  } catch (err: any) {
    errorMessage.value = err.message || 'Error al registrarse'
  }
}

async function handleConfirm() {
  errorMessage.value = ''
  try {
    await auth.confirmRegistration(email.value, confirmCode.value)
    toast.add({ severity: 'success', summary: 'Cuenta verificada', detail: 'Ya puedes iniciar sesión', life: 3000 })
    formState.value = 'login'
  } catch (err: any) {
    errorMessage.value = err.message || 'Error al confirmar'
  }
}

function switchToRegister() {
  errorMessage.value = ''
  formState.value = 'register'
}

function switchToLogin() {
  errorMessage.value = ''
  formState.value = 'login'
}
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <div class="login-header">
        <i class="pi pi-database logo-icon"></i>
        <h1>DB Parser</h1>
        <p class="login-subtitle">SQL to NoSQL Converter</p>
      </div>

      <Message v-if="errorMessage" severity="error" :closable="false">{{ errorMessage }}</Message>

      <!-- Login Form -->
      <form v-if="formState === 'login'" @submit.prevent="handleLogin" class="login-form">
        <div class="field">
          <label for="email">Email</label>
          <InputText id="email" v-model="email" type="email" placeholder="tu@email.com" :fluid="true" />
        </div>
        <div class="field">
          <label for="password">Contraseña</label>
          <Password id="password" v-model="password" :feedback="false" toggleMask :fluid="true" />
        </div>
        <Button type="submit" label="Iniciar Sesión" :loading="auth.loading" :fluid="true" />
        <p class="switch-text">
          ¿No tienes cuenta?
          <a href="#" @click.prevent="switchToRegister">Regístrate</a>
        </p>
      </form>

      <!-- Register Form -->
      <form v-if="formState === 'register'" @submit.prevent="handleRegister" class="login-form">
        <div class="field">
          <label for="reg-email">Email</label>
          <InputText id="reg-email" v-model="email" type="email" placeholder="tu@email.com" :fluid="true" />
        </div>
        <div class="field">
          <label for="reg-password">Contraseña</label>
          <Password id="reg-password" v-model="password" toggleMask :fluid="true" />
        </div>
        <div class="field">
          <label for="reg-confirm">Confirmar Contraseña</label>
          <Password id="reg-confirm" v-model="confirmPassword" :feedback="false" toggleMask :fluid="true" />
        </div>
        <Button type="submit" label="Registrarse" :loading="auth.loading" :fluid="true" />
        <p class="switch-text">
          ¿Ya tienes cuenta?
          <a href="#" @click.prevent="switchToLogin">Inicia Sesión</a>
        </p>
      </form>

      <!-- Confirm Form -->
      <form v-if="formState === 'confirm'" @submit.prevent="handleConfirm" class="login-form">
        <p class="confirm-info">Se envió un código de verificación a <strong>{{ email }}</strong></p>
        <div class="field">
          <label for="confirm-code">Código de Verificación</label>
          <InputText id="confirm-code" v-model="confirmCode" placeholder="123456" :fluid="true" />
        </div>
        <Button type="submit" label="Verificar" :loading="auth.loading" :fluid="true" />
        <p class="switch-text">
          <a href="#" @click.prevent="switchToLogin">Volver al login</a>
        </p>
      </form>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 1rem;
}

.login-card {
  background: #ffffff;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
  padding: 2.5rem;
  width: 100%;
  max-width: 420px;
}

.login-header {
  text-align: center;
  margin-bottom: 2rem;
}

.logo-icon {
  font-size: 2.5rem;
  color: #3b82f6;
  margin-bottom: 0.5rem;
}

.login-header h1 {
  margin: 0.5rem 0 0.25rem;
  font-size: 1.5rem;
  color: #111827;
}

.login-subtitle {
  margin: 0;
  font-size: 0.875rem;
  color: #6b7280;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.field label {
  font-size: 0.85rem;
  font-weight: 600;
  color: #374151;
}

.switch-text {
  text-align: center;
  font-size: 0.85rem;
  color: #6b7280;
  margin: 0;
}

.switch-text a {
  color: #3b82f6;
  font-weight: 600;
  text-decoration: none;
}

.switch-text a:hover {
  text-decoration: underline;
}

.confirm-info {
  font-size: 0.875rem;
  color: #374151;
  text-align: center;
  margin: 0;
}
</style>
