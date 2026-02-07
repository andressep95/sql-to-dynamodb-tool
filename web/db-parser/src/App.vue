<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import Sidebar from '@/components/Sidebar.vue'
import Toast from 'primevue/toast'
import Button from 'primevue/button'

const route = useRoute()
const sidebarExpanded = ref(false)
const mobileMenuOpen = ref(false)

const showSidebar = computed(() => route.name !== 'login')

const mainMargin = computed(() => {
  if (!showSidebar.value) return '0'
  return sidebarExpanded.value ? '220px' : '70px'
})

function toggleMobileMenu() {
  mobileMenuOpen.value = !mobileMenuOpen.value
}
</script>

<template>
  <Toast />

  <!-- Mobile hamburger -->
  <Button
    v-if="showSidebar"
    icon="pi pi-bars"
    text
    rounded
    class="mobile-menu-btn"
    @click="toggleMobileMenu"
  />

  <!-- Mobile overlay -->
  <div v-if="mobileMenuOpen && showSidebar" class="mobile-overlay" @click="mobileMenuOpen = false"></div>

  <Sidebar
    v-if="showSidebar"
    :is-mobile-open="mobileMenuOpen"
    @update:expanded="sidebarExpanded = $event"
  />

  <main class="main-content" :style="{ marginLeft: mainMargin }">
    <RouterView />
  </main>
</template>

<style scoped>
.main-content {
  transition: margin-left 0.3s ease;
  min-height: 100vh;
  padding: 1rem;
}

.mobile-menu-btn {
  display: none;
  position: fixed;
  top: 12px;
  left: 12px;
  z-index: 1100;
}

.mobile-overlay {
  display: none;
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 999;
}

@media (max-width: 768px) {
  .mobile-menu-btn {
    display: block;
  }

  .mobile-overlay {
    display: block;
  }

  .main-content {
    margin-left: 0 !important;
    padding-top: 3.5rem;
  }
}
</style>
