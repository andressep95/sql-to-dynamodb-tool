<script setup lang="ts">
</script>

<template>
  <div class="history-view">
    <h2>Historial de Conversiones</h2>
    <p class="subtitle">Aquí aparecerán tus conversiones anteriores</p>

    <div class="empty-state">
      <i class="pi pi-history empty-icon"></i>
      <p>No hay conversiones registradas</p>
    </div>
  </div>
</template>

<style scoped>
.history-view {
  max-width: 900px;
  margin: 0 auto;
  padding: 2rem;
}

.subtitle {
  color: #6b7280;
  margin-bottom: 2rem;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  color: #9ca3af;
  gap: 1rem;
}

.empty-icon {
  font-size: 3rem;
}
</style>
