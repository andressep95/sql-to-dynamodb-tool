<template>
  <aside :class="['sidebar', { expanded: isExpanded, 'mobile-open': isMobileOpen }]" @mouseenter="expand" @mouseleave="collapse">
    <!-- Logo / Header -->
    <div class="sidebar-header">
      <i class="pi pi-database logo-icon"></i>
      <span v-show="isExpanded || isMobileOpen" class="logo-text">DB Parser</span>
    </div>

    <!-- Navigation Menu -->
    <nav class="sidebar-nav">
      <router-link
        v-for="item in visibleMenuItems"
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
          <span class="user-name">{{ auth.displayName || 'Usuario' }}</span>
          <span class="user-email">{{ auth.userEmail }}</span>
        </div>
      </div>
      <button
        v-show="isExpanded || isMobileOpen"
        class="logout-btn"
        @click="handleLogout"
        title="Cerrar sesión"
      >
        <i class="pi pi-sign-out"></i>
        <span>Cerrar sesión</span>
      </button>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { ref, watch, toRefs, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Avatar from 'primevue/avatar'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const props = defineProps<{
  isMobileOpen?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:expanded', value: boolean): void
}>()

const isExpanded = ref(false)
const { isMobileOpen } = toRefs(props)

watch(isExpanded, (newValue) => {
  emit('update:expanded', newValue)
})

const menuItems = [
  { label: 'Parser', icon: 'pi pi-code', path: '/', roles: ['SUPER_ADMIN', 'REALM_ADMIN', 'REALM_SUPERVISOR', 'USER_TENANT'] },
  { label: 'Historial', icon: 'pi pi-history', path: '/history', roles: ['SUPER_ADMIN', 'REALM_ADMIN', 'REALM_SUPERVISOR', 'USER_TENANT'] },
  { label: 'Usuarios', icon: 'pi pi-users', path: '/tenants', roles: ['SUPER_ADMIN', 'REALM_ADMIN'] },
]

const visibleMenuItems = computed(() => {
  const userRole = auth.userRole || 'USER_TENANT'
  return menuItems.filter(item => item.roles.includes(userRole))
})

const userInitials = computed(() => {
  if (!auth.displayName) return 'U'
  return auth.displayName.substring(0, 2).toUpperCase()
})

const isActive = (path: string) => route.path === path

const expand = () => { isExpanded.value = true }
const collapse = () => { isExpanded.value = false }

const handleLogout = () => {
  auth.logout()
  router.push('/login')
}
</script>

<style scoped>
.sidebar,
.logo-text,
.menu-label,
.user-name,
.user-email {
  font-family: 'Inter', system-ui, -apple-system, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

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

.sidebar-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1.5rem 1rem;
  border-bottom: 1px solid #e5e7eb;
  min-height: 70px;
}

.logo-icon {
  font-size: 1.75rem;
  color: #3b82f6;
  min-width: 1.75rem;
}

.logo-text {
  font-size: 1.2rem;
  font-weight: 700;
  color: #111827;
  white-space: nowrap;
  overflow: hidden;
}

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
}

.sidebar-item:hover {
  background-color: #f3f4f6;
  color: #374151;
}

.sidebar-item.active {
  background-color: #3b82f6;
  color: #ffffff;
  box-shadow: 0 4px 6px -1px rgba(59, 130, 246, 0.3), 0 2px 4px -2px rgba(59, 130, 246, 0.3);
}

.sidebar-item.active .menu-icon,
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
}

.user-info :deep(.p-avatar) {
  flex-shrink: 0;
  width: 2.25rem !important;
  height: 2.25rem !important;
  background-color: #3b82f6 !important;
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
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  background: transparent;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  color: #6b7280;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.logout-btn:hover {
  background-color: #fee2e2;
  border-color: #fecaca;
  color: #dc2626;
}

.logout-btn i {
  font-size: 1rem;
}

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
