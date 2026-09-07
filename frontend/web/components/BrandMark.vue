<script setup lang="ts">
withDefaults(defineProps<{ size?: number }>(), { size: 30 })
</script>

<template>
  <!-- Монограмма из разделителя: пустая рамка знака, точка на месте региона -->
  <svg :width="size" :height="size * 0.7" viewBox="0 0 40 28" aria-hidden="true" class="mark">
    <rect x="1.8" y="1.8" width="36.4" height="24.4" rx="3"
          fill="none" stroke="currentColor" stroke-width="3.6" />
    <line x1="27" y1="1.8" x2="27" y2="26.2" stroke="currentColor" stroke-width="3.6" />
    <circle cx="33" cy="14" r="3.9" fill="var(--accent)" />
  </svg>
</template>

<style scoped>
.mark { flex: none; display: block; }
</style>
