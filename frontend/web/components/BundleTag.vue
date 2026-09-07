<script setup lang="ts">
import type { Reissue } from '~/types/api'

const props = defineProps<{ reissue: Reissue; short?: boolean }>()

/** null — это «площадка не сказала», а не «не включено» */
const label = computed(() => {
  if (props.reissue === true) return props.short ? 'С оформлением' : 'С переоформлением'
  if (props.reissue === false) return props.short ? 'Без оформления' : 'Без переоформления'
  return 'Оформление не указано'
})

const kind = computed(() =>
  props.reissue === true ? 'yes' : props.reissue === false ? 'no' : 'unknown')
</script>

<template>
  <!-- Смысл никогда не передаётся одним цветом: всегда слово плюс фон -->
  <span class="tag" :class="kind">{{ label }}</span>
</template>

<style scoped>
.tag {
  font-size: 13.5px; font-weight: 600; padding: 4px 9px; border-radius: var(--r-sm);
  display: inline-flex; align-items: center; white-space: nowrap;
}
.yes     { background: var(--good-bg);  color: var(--good); }
.no      { background: var(--alert-bg); color: var(--alert); }
.unknown { background: var(--warn-bg);  color: var(--warn); }
</style>
