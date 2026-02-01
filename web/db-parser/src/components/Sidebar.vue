<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import Button from 'primevue/button'

const router = useRouter()
const route = useRoute()

const isExpanded = ref(true)

defineExpose({ isExpanded })

const menuItems = [
  { label: 'Parser', icon: 'pi pi-code', path: '/' },
  { label: 'Historial', icon: 'pi pi-history', path: '/history' },
]

const sidebarWidth = computed(() => (isExpanded.value ? '260px' : '70px'))

function toggleSidebar() {
  isExpanded.value = !isExpanded.value
}

function navigate(path: string) {
  router.push(path)
}

function isActive(path: string) {
  return route.path === path
}
</script>

<template>
  <aside class="sidebar" :style="{ width: sidebarWidth }">
    <div class="sidebar-header">
      <div class="logo-section" @click="navigate('/')">
        <i class="pi pi-database logo-icon"></i>
        <span v-if="isExpanded" class="logo-text">DB Parser</span>
      </div>
      <Button
        :icon="isExpanded ? 'pi pi-angle-left' : 'pi pi-angle-right'"
        text
        rounded
        size="small"
        class="toggle-btn"
        @click="toggleSidebar"
      />
    </div>

    <nav class="sidebar-nav">
      <div
        v-for="item in menuItems"
        :key="item.path"
        class="nav-item"
        :class="{ active: isActive(item.path) }"
        @click="navigate(item.path)"
        v-tooltip.right="!isExpanded ? item.label : undefined"
      >
        <i :class="item.icon"></i>
        <span v-if="isExpanded">{{ item.label }}</span>
      </div>
    </nav>

    <div class="sidebar-footer">
      <div class="user-section">
        <div class="avatar">DB</div>
        <span v-if="isExpanded" class="user-name">SQL to NoSQL</span>
      </div>
    </div>
  </aside>
</template>

<style scoped>
.sidebar {
  position: fixed;
  top: 0;
  left: 0;
  height: 100vh;
  background: #1e1e2e;
  color: #cdd6f4;
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
  overflow: hidden;
  z-index: 1000;
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 12px;
  border-bottom: 1px solid #313244;
}

.logo-section {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
}

.logo-icon {
  font-size: 1.5rem;
  color: #89b4fa;
}

.logo-text {
  font-size: 1.2rem;
  font-weight: 700;
  white-space: nowrap;
  color: #cdd6f4;
}

.toggle-btn {
  color: #cdd6f4 !important;
}

.sidebar-nav {
  flex: 1;
  padding: 12px 8px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.2s;
  white-space: nowrap;
}

.nav-item:hover {
  background: #313244;
}

.nav-item.active {
  background: #89b4fa;
  color: #1e1e2e;
}

.nav-item i {
  font-size: 1.1rem;
  min-width: 20px;
  text-align: center;
}

.sidebar-footer {
  padding: 16px 12px;
  border-top: 1px solid #313244;
}

.user-section {
  display: flex;
  align-items: center;
  gap: 10px;
}

.avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: #89b4fa;
  color: #1e1e2e;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 0.85rem;
  flex-shrink: 0;
}

.user-name {
  font-size: 0.9rem;
  white-space: nowrap;
}
</style>
