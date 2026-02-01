sidebar:
<template>

  <aside :class="['sidebar', { expanded: isExpanded, 'mobile-open': isMobileOpen }]" @mouseenter="expand" @mouseleave="collapse">
    <!-- Logo / Header -->
    <div class="sidebar-header">
      <i class="pi pi-shield logo-icon"></i>
      <span v-show="isExpanded || isMobileOpen" class="logo-text">CV Processor</span>
    </div>

    <!-- Navigation Menu -->
    <nav class="sidebar-nav">
      <router-link
        v-for="item in menuItems"
        :key="item.path"
        :to="item.path"
        class="sidebar-item"
        :class="{ active: isActive(item.path) }"
        :title="!isExpanded ? item.label : ''"
      >
        <i :class="['menu-icon', item.icon]"></i>
        <span v-show="isExpanded || isMobileOpen" class="menu-label">{{ item.label }}</span>
      </router-link>
    </nav>

    <!-- User Section -->
    <div class="sidebar-footer">
      <div class="user-info">
        <Avatar :label="userInitials" shape="circle" size="normal" />
        <div v-show="isExpanded || isMobileOpen" class="user-details">
          <span class="user-name">{{ username }}</span>
          <span class="user-email">{{ email }}</span>
        </div>
      </div>
      <Button
        icon="pi pi-sign-out"
        severity="danger"
        :label="(isExpanded || isMobileOpen) ? 'Salir' : ''"
        @click="handleLogout"
        class="logout-btn"
        :title="!(isExpanded || isMobileOpen) ? 'Salir' : ''"
      />
    </div>

  </aside>
</template>

<script setup lang="ts">
import { ref, computed, watch, toRefs } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import Button from 'primevue/button'
import Avatar from 'primevue/avatar'
import { decodeJWT } from '../utils/auth'

const router = useRouter()
const route = useRoute()

const props = defineProps<{
  isMobileOpen?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:expanded', value: boolean): void
}>()

const isExpanded = ref(false)
const { isMobileOpen = false } = toRefs(props)

// Emitir cuando cambia el estado de expansión
watch(isExpanded, (newValue) => {
  emit('update:expanded', newValue)
})

const menuItems = [
  {
    label: 'Procesar CV',
    icon: 'pi pi-home',
    path: '/'
  },
  {
    label: 'Mis CVs',
    icon: 'pi pi-database',
    path: '/my-resumes'
  }
]

const username = computed(() => {
  const token = localStorage.getItem('authToken')
  if (token) {
    const decoded = decodeJWT(token)
    return decoded?.email?.split('@')[0] || 'Usuario'
  }
  return 'Usuario'
})

const email = computed(() => {
  const token = localStorage.getItem('authToken')
  if (token) {
    const decoded = decodeJWT(token)
    return decoded?.email || 'usuario@ejemplo.com'
  }
  return 'usuario@ejemplo.com'
})

const userInitials = computed(() => {
  const name = username.value || 'U'
  return name
    .split(' ')
    .map((n: string) => n[0])
    .join('')
    .toUpperCase()
    .substring(0, 2)
})

const isActive = (path: string) => {
  return route.path === path
}

const expand = () => {
  isExpanded.value = true
}

const collapse = () => {
  isExpanded.value = false
}

const handleLogout = () => {
  localStorage.removeItem('authToken')
  localStorage.removeItem('refreshToken')
  router.push('/login')
}
</script>

<style scoped>
/* TIPOGRAFÍA GLOBAL */
.sidebar,
.logo-text,
.menu-label,
.user-name,
.user-email {
  font-family: 'Inter', system-ui, -apple-system, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

/* SIDEBAR PRINCIPAL */
.sidebar {
  position: fixed;
  left: 0;
  top: 0;
  height: 100vh;
  background-color: #ffffff;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -2px rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
  width: 70px;
  transition: width 0.3s ease;
  z-index: 1000;
  overflow: hidden;
  border-top-right-radius: 8px;
  border-bottom-right-radius: 8px;
  border: 1px solid #e5e7eb;
  border-left: none;
}

.sidebar.expanded {
  width: 220px;
}

/* HEADER */
.sidebar-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1.5rem 1.1rem;
  border-bottom: 1px solid #e5e7eb;
  min-height: 70px;
}

.logo-icon {
  font-size: 1.75rem;
  color: #10b981;
  min-width: 1.75rem;
}

.logo-text {
  font-size: 1.2rem;
  font-weight: 700;
  color: #111827;
  white-space: nowrap;
  overflow: hidden;
}

/* NAVIGATION */
.sidebar-nav {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 1rem 0.5rem;
  overflow-y: auto;
}

.sidebar-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.875rem 1rem;
  border-radius: 8px;
  color: #6b7280;
  text-decoration: none;
  transition: all 0.2s ease;
  cursor: pointer;
  position: relative;
}

.sidebar-item:hover {
  background-color: #f3f4f6;
  color: #374151;
}

.sidebar-item.active {
  background-color: #10b981;
  color: #ffffff;
  box-shadow: 0 4px 6px -1px rgba(16, 185, 129, 0.3), 0 2px 4px -2px rgba(16, 185, 129, 0.3);
}

.sidebar-item.active .menu-icon {
  color: #ffffff;
}

.sidebar-item.active .menu-label {
  color: #ffffff;
}

.menu-icon {
  font-size: 1.25rem;
  min-width: 1.25rem;
  color: #6b7280;
  transition: color 0.2s ease;
}

.sidebar-item:hover .menu-icon {
  color: #374151;
}

.sidebar-item.active:hover .menu-icon {
  color: #ffffff;
}

.menu-label {
  font-size: 0.95rem;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  color: #6b7280;
}

.sidebar-item:hover .menu-label {
  color: #374151;
}

/* FOOTER */
.sidebar-footer {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 1rem;
  border-top: 1px solid #e5e7eb;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0;
}

.user-info :deep(.p-avatar) {
  flex-shrink: 0;
  width: 2.25rem !important;
  height: 2.25rem !important;
  background-color: #10b981 !important;
  color: #ffffff !important;
}

.user-details {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  overflow: hidden;
}

.user-name {
  font-size: 0.875rem;
  font-weight: 600;
  color: #111827;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-email {
  font-size: 0.75rem;
  color: #6b7280;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.logout-btn {
  width: 100%;
  justify-content: center;
}

/* SCROLLBAR */
.sidebar-nav::-webkit-scrollbar {
  width: 4px;
}

.sidebar-nav::-webkit-scrollbar-track {
  background: transparent;
}

.sidebar-nav::-webkit-scrollbar-thumb {
  background: #d1d5db;
  border-radius: 4px;
}

.sidebar-nav::-webkit-scrollbar-thumb:hover {
  background: #9ca3af;
}

/* Mobile Styles */
@media (max-width: 768px) {
  .sidebar {
    transform: translateX(-100%);
    transition: transform 0.3s ease;
    width: 280px;
    z-index: 9999;
  }
  
  .sidebar.mobile-open {
    transform: translateX(0);
  }
  
  .sidebar:hover {
    width: 280px;
  }
  
  .sidebar-header {
    padding: 1rem;
    min-height: 50px;
  }
  
  .logo-icon {
    font-size: 1.5rem;
  }
  
  .logo-text {
    font-size: 1rem;
  }
  
  .sidebar-nav {
    padding: 0.5rem;
    gap: 0.3rem;
  }
  
  .sidebar-item {
    padding: 0.6rem 0.8rem;
  }
  
  .sidebar-footer {
    padding: 0.8rem;
    gap: 0.4rem;
  }
  
  .user-info {
    gap: 0.5rem;
  }
  
  .user-info :deep(.p-avatar) {
    width: 1.8rem !important;
    height: 1.8rem !important;
  }
  
  .user-name {
    font-size: 0.8rem;
  }
  
  .user-email {
    font-size: 0.7rem;
  }
}
</style>

---

App.vue
<template>

  <div class="app-container">
    <!-- Mobile Overlay -->
    <div 
      v-if="isAuthenticated && isMobile && isMobileSidebarOpen" 
      class="mobile-overlay"
      @click="closeMobileSidebar"
    ></div>

    <!-- Mobile Hamburger Button -->
    <Button
      v-if="isAuthenticated && isMobile"
      icon="pi pi-bars"
      class="mobile-menu-btn"
      @click="toggleMobileSidebar"
      severity="secondary"
    />

    <!-- Sidebar -->
    <Sidebar
      v-if="isAuthenticated"
      :is-mobile-open="isMobileSidebarOpen"
      @update:expanded="handleSidebarExpanded"
    />

    <!-- Contenido principal -->
    <main
      :class="['main-content', { 'with-sidebar': isAuthenticated && !isMobile }]"
      :style="{ marginLeft: sidebarMargin }"
    >
      <RouterView />
    </main>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import Button from 'primevue/button'
import Sidebar from './components/Sidebar.vue'

const router = useRouter()
const route = useRoute()
const isSidebarExpanded = ref(false)
const isAuthenticated = ref(false)
const isMobileSidebarOpen = ref(false)
const isMobile = ref(false)

const sidebarMargin = computed(() => {
  if (!isAuthenticated.value || isMobile.value) return '0'
  return isSidebarExpanded.value ? '220px' : '70px'
})

const handleSidebarExpanded = (expanded: boolean) => {
  isSidebarExpanded.value = expanded
}

const checkMobile = () => {
  isMobile.value = window.innerWidth <= 768
}

const toggleMobileSidebar = () => {
  isMobileSidebarOpen.value = !isMobileSidebarOpen.value
}

const closeMobileSidebar = () => {
  isMobileSidebarOpen.value = false
}

const checkAuth = () => {
  const token = localStorage.getItem('authToken')
  isAuthenticated.value = !!token
}

// Watcher para detectar cambios de ruta
watch(() => route.path, () => {
  checkAuth()
})

onMounted(() => {
  checkAuth()
  checkMobile()
  window.addEventListener('resize', checkMobile)
})
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: var(--font-family);
  overflow-x: hidden;
  background-color: #f9fafb;
}

#app {
  min-height: 100vh;
}
</style>

<style scoped>
.app-container {
  min-height: 100vh;
  background-color: #f9fafb;
}

.main-content {
  min-height: 100vh;
  padding: 2rem;
  transition: margin-left 0.3s ease;
  overflow-y: auto;
}

.main-content:not(.with-sidebar) {
  padding: 0;
}

/* Mobile Styles */
.mobile-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(0, 0, 0, 0.5);
  z-index: 9998;
}

.mobile-menu-btn {
  position: fixed;
  bottom: 2rem;
  right: 2rem;
  z-index: 10000;
  border-radius: 50%;
  width: 3.5rem;
  height: 3.5rem;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

@media (max-width: 768px) {
  .main-content {
    padding: 1rem;
    margin-left: 0 !important;
    padding-bottom: 5rem;
  }
  
  .main-content:not(.with-sidebar) {
    padding: 0;
    padding-bottom: 5rem;
  }
}
</style>

---

homeview:

<template>
  <div class="home-view">
    <ResumeUploadForm @upload="handleUpload" />
    
    <div v-if="uploadResponse" class="upload-response">
      <h3>Resultado:</h3>
      <pre>{{ JSON.stringify(uploadResponse, null, 2) }}</pre>
    </div>
    
    <div v-if="errorMessage" class="error-message">
      <h3>Error:</h3>
      <p>{{ errorMessage }}</p>
    </div>
    
    <Toast />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useToast } from 'primevue/usetoast'
import ResumeUploadForm from '../components/ResumeUploadForm.vue'
import Toast from 'primevue/toast'

const router = useRouter()
const toast = useToast()
const uploadResponse = ref<any>(null)
const errorMessage = ref('')

// Verificar autenticación básica
const checkAuth = () => {
  const token = localStorage.getItem('authToken')
  if (!token) {
    router.push('/login')
    return
  }
  
  // Verificar si el token sigue siendo válido
  import('../utils/auth').then(({ isTokenValid, clearAuthAndRedirect }) => {
    if (!isTokenValid(token)) {
      clearAuthAndRedirect()
    }
  })
}

// Manejar upload de CV
const handleUpload = async (data: { file: File; language: string; instructions: string }) => {
  const token = localStorage.getItem('authToken')
  errorMessage.value = ''
  uploadResponse.value = null
  
  if (!token) {
    errorMessage.value = 'No hay token de autenticación'
    return
  }
  
  try {
    const apiUrl = import.meta.env.VITE_RESUME_API_URL || 'https://api.cloudcentinel.com/resume/api/v1/resume'
    
    const formData = new FormData()
    formData.append('file', data.file)
    formData.append('language', data.language)
    if (data.instructions) {
      formData.append('instructions', data.instructions)
    }
    
    console.log('📤 Enviando CV:', {
      fileName: data.file.name,
      fileSize: data.file.size,
      language: data.language,
      instructions: data.instructions
    })
    
    const response = await fetch(apiUrl, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`
      },
      body: formData
    })
    
    console.log('📡 Respuesta HTTP status:', response.status)
    
    if (response.ok) {
      const result = await response.json()
      uploadResponse.value = result
      
      toast.add({
        severity: 'success',
        summary: 'Éxito',
        detail: 'CV enviado para procesamiento',
        life: 5000
      })
    } else {
      const error = await response.json()
      errorMessage.value = error.message || 'Error al procesar CV'
      
      toast.add({
        severity: 'error',
        summary: 'Error',
        detail: error.message || 'Error al procesar CV',
        life: 5000
      })
    }
  } catch (error) {
    console.error('❌ Error en upload:', error)
    errorMessage.value = `Error de conexión: ${error instanceof Error ? error.message : 'Error desconocido'}`
    
    toast.add({
      severity: 'error',
      summary: 'Error',
      detail: 'Error de conexión. Intenta nuevamente.',
      life: 5000
    })
  }
}





onMounted(() => {
  checkAuth()
})
</script>

<style scoped>
.home-view {
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}



.upload-response, .error-message {
  margin: 2rem 0;
  padding: 1rem;
  border-radius: 4px;
}

.upload-response {
  border: 1px solid #28a745;
  background: #f8fff9;
}

.error-message {
  border: 1px solid #dc3545;
  background: #fff8f8;
  color: #dc3545;
}

pre {
  background: #f5f5f5;
  padding: 1rem;
  border-radius: 4px;
  overflow-x: auto;
  font-size: 0.9rem;
  max-height: 400px;
  overflow-y: auto;
}

@media (max-width: 768px) {
  .header {
    flex-direction: column;
    gap: 1rem;
    text-align: center;
  }
}
</style>

---

router:
import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import HomeView from '../views/HomeView.vue'
import MyResumesView from '../views/MyResumesView.vue'
import ResumeDetailView from '../views/ResumeDetailView.vue'
import VerifyEmailView from '../views/VerifyEmailView.vue'
import RegisterSuccessView from '../views/RegisterSuccessView.vue'
import ForgotPasswordView from '../views/ForgotPasswordView.vue'
import ResetPasswordView from '../views/ResetPasswordView.vue'

const router = createRouter({
history: createWebHistory(import.meta.env.BASE_URL),
routes: [
{
path: '/login',
name: 'login',
component: LoginView,
},
{
path: '/',
name: 'home',
component: HomeView,
meta: { requiresAuth: true },
},
{
path: '/my-resumes',
name: 'my-resumes',
component: MyResumesView,
meta: { requiresAuth: true },
},
{
path: '/resume/:id',
name: 'resume-detail',
component: ResumeDetailView,
meta: { requiresAuth: true },
},
{
path: '/verify-email',
name: 'verify-email',
component: VerifyEmailView,
},
{
path: '/register-success',
name: 'register-success',
component: RegisterSuccessView,
},
{
path: '/forgot-password',
name: 'forgot-password',
component: ForgotPasswordView,
},
{
path: '/reset-password',
name: 'reset-password',
component: ResetPasswordView,
},
{
path: '/:pathMatch(.*)*',
name: 'not-found',
redirect: '/login'
},
],
})

// Guard de navegación para rutas protegidas
router.beforeEach(async (to, from, next) => {
// Permitir acceso a rutas públicas sin verificación
const publicRoutes = ['verify-email', 'register-success', 'forgot-password', 'reset-password']
if (publicRoutes.includes(to.name as string)) {
next()
return
}

const token = localStorage.getItem('authToken')

if (to.meta.requiresAuth) {
if (!token) {
next('/login')
return
}

    // Verificar si el token es válido
    const { isTokenValid, clearAuthAndRedirect } = await import('../utils/auth')
    if (!isTokenValid(token)) {
      clearAuthAndRedirect()
      return
    }

}

if (to.name === 'login' && token) {
const { isTokenValid } = await import('../utils/auth')
if (isTokenValid(token)) {
next('/')
return
}
}

next()
})

export default router
