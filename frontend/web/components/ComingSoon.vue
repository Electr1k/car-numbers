<script setup lang="ts">
/**
 * Раздел заявлен в меню, но не готов. Заглушка обязана говорить три вещи:
 * что здесь будет, почему пока нет и куда пойти сейчас. Тупик без выхода —
 * худший вариант для аудитории, которая и так осторожно относится к незнакомым сайтам.
 */
defineProps<{ title: string; lead: string }>()
</script>

<template>
  <div class="container wrap">
    <p class="badge">Раздел в разработке</p>
    <h1>{{ title }}</h1>
    <p class="lead">{{ lead }}</p>

    <div class="body">
      <slot />
    </div>

    <div class="exits">
      <slot name="exits" />
    </div>
  </div>
</template>

<style scoped>
.wrap { padding: 52px 24px 64px; max-width: 760px; }
.badge {
  display: inline-flex; align-items: center; margin-bottom: 16px;
  background: var(--warn-bg); color: var(--warn);
  font-size: 13.5px; font-weight: 600; padding: 5px 11px; border-radius: var(--r-sm);
}
.lead { font-size: 18px; color: var(--text-muted); margin: 14px 0 0; }
.body { margin-top: 32px; display: grid; gap: 20px; }
.exits { margin-top: 36px; display: flex; flex-wrap: wrap; gap: 12px; }
</style>
