<script setup lang="ts">
/** Иконка продавца с сайта площадки; нет иконки или она не открылась — первая буква имени */
const props = defineProps<{ src: string | null; title: string; size?: number }>()

const failed = ref(false)
watch(() => props.src, () => { failed.value = false })

const initial = computed(() => props.title.charAt(0).toUpperCase())
</script>

<template>
  <img v-if="src && !failed" :src="src" alt="" class="ava" :style="{ '--size': `${size ?? 36}px` }"
       loading="lazy" referrerpolicy="no-referrer" @error="failed = true">
  <span v-else class="ava letter" :style="{ '--size': `${size ?? 36}px` }" aria-hidden="true">{{ initial }}</span>
</template>

<style scoped>
.ava {
  width: var(--size); height: var(--size); border-radius: 50%; flex: none;
  object-fit: cover; background: var(--surface-sunk);
}
.letter {
  display: inline-grid; place-items: center;
  font-weight: 600; font-size: calc(var(--size) * .42); color: var(--text-muted);
}
</style>
