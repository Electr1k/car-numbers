<script setup lang="ts">
/* Поиск живёт на главной; старые ссылки ведут туда же с теми же параметрами */
definePageMeta({
  middleware: to => navigateTo({ path: '/', query: to.query }, { redirectCode: 301 })
})
</script>

<template>
  <div />
</template>
