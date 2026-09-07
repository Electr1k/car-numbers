<script setup lang="ts">
/** Номер набором, а не мини-знаком: узкое начертание не переносит малый размер. */
withDefaults(defineProps<{ number: string; size?: 'sm' | 'md' | 'lg' }>(), { size: 'md' })

const parts = (n: string) => {
  const m = n.match(/^(.*?)(\d{2,3})$/)
  return m ? { main: m[1], region: m[2] } : { main: n, region: '' }
}
</script>

<template>
  <span class="num" :class="size">
    <span>{{ parts(number).main }}</span>
    <span v-if="parts(number).region" class="rg">{{ parts(number).region }}</span>
  </span>
</template>

<style scoped>
.num {
  font-family: var(--font-plate);
  font-weight: 700;
  letter-spacing: .045em;
  font-variant-numeric: tabular-nums;
  line-height: 1;
  display: inline-flex;
  align-items: baseline;
  gap: .34em;
  white-space: nowrap;
}
.rg { color: var(--text-muted); font-size: .68em; letter-spacing: .02em; }
.sm { font-size: 24px; }
.md { font-size: 30px; }
.lg { font-size: 34px; }
</style>
