<script setup lang="ts">
useHead({ title: 'Вход — Номерограф' })
</script>

<template>
  <ComingSoon
    title="Вход"
    lead="Всё, что умеет сайт сегодня, работает без регистрации."
  >
    <section>
      <h2>Что сейчас доступно без входа</h2>
      <ul class="list">
        <li>Оценка любого номера, включая те, которых нет в продаже.</li>
        <li>Объявления трёх площадок в одной карточке и переход к продавцу.</li>
        <li>Поиск по номеру, маске и узору.</li>
      </ul>
    </section>

    <template #exits>
      <NuxtLink to="/estimate" class="btn-primary">Оценить номер</NuxtLink>
      <NuxtLink to="/" class="btn-quiet">На главную</NuxtLink>
    </template>
  </ComingSoon>
</template>

<style scoped>
section { display: grid; gap: 10px; }
h2 { font-size: 20px; }
p { color: var(--text-muted); }
.list { margin: 0; padding-left: 22px; display: grid; gap: 8px; color: var(--text-muted); }

.btn-primary, .btn-quiet {
  min-height: 44px; padding: 11px 20px; display: inline-flex; align-items: center;
  border-radius: var(--r-md); font-size: 16px; font-weight: 600; text-decoration: none;
}
.btn-primary { background: var(--accent); color: #fff; }
.btn-primary:hover { background: var(--accent-hover); color: #fff; text-decoration: none; }
.btn-quiet { background: var(--surface); color: var(--text); border: 1px solid var(--border-strong); }
.btn-quiet:hover { border-color: var(--text-muted); color: var(--text); text-decoration: none; }
</style>
