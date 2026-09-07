<script setup lang="ts">
const theme = useTheme()

const nav = [
  { to: '/',                label: 'Поиск' },
  { to: '/estimate',        label: 'Оценка номера' },
  { to: '/triggers',        label: 'Триггеры' },
  { to: '/reissue',         label: 'Переоформление' }
]
</script>

<template>
  <header class="bar">
    <div class="inner container">
      <NuxtLink to="/" class="brand">
        <BrandMark :size="30" />
        <span class="name">Номеро<span class="g">граф</span></span>
      </NuxtLink>

      <nav class="nav" aria-label="Основные разделы">
        <NuxtLink v-for="i in nav" :key="i.to" :to="i.to">{{ i.label }}</NuxtLink>
      </nav>

      <div class="util">
        <!-- Иконка в одиночку кнопкой не является: у каждой есть подпись -->
        <button type="button" class="ghost" @click="theme.toggle()">
          {{ theme.isDark.value ? 'Светлая тема' : 'Тёмная тема' }}
        </button>
        <button type="button" class="ghost">Рус</button>
        <NuxtLink to="/login" class="signin">Войти</NuxtLink>
      </div>
    </div>
  </header>
</template>

<style scoped>
.bar { background: var(--surface); border-bottom: 1px solid var(--border); }
.inner { display: flex; align-items: center; gap: 24px; padding-top: 12px; padding-bottom: 12px; }
.brand { display: flex; align-items: center; gap: 10px; color: var(--text); text-decoration: none; }
.brand:hover { text-decoration: none; }
.name { font-weight: 700; font-size: 20px; letter-spacing: -.022em; }
.g { color: var(--text-faint); font-weight: 600; }

.nav { display: flex; gap: 4px; flex: 1; flex-wrap: wrap; }
.nav a {
  font-size: 15.5px; font-weight: 500; color: var(--text-muted); text-decoration: none;
  padding: 11px 13px; min-height: 44px; display: flex; align-items: center; border-radius: var(--r-sm);
}
.nav a:hover { color: var(--text); text-decoration: none; }
.nav a.router-link-active { color: var(--text); font-weight: 600; background: var(--surface-sunk); }

.util { display: flex; gap: 8px; align-items: center; }
.ghost {
  font-size: 14.5px; font-weight: 500; min-height: 44px; padding: 10px 13px; cursor: pointer;
  background: transparent; border: 1px solid var(--border); border-radius: var(--r-sm); color: var(--text-muted);
}
.ghost:hover { color: var(--text); border-color: var(--border-strong); }
.signin {
  font-size: 16px; font-weight: 600; min-height: 44px; padding: 11px 20px;
  display: inline-flex; align-items: center; text-decoration: none;
  background: var(--surface); color: var(--text);
  border: 1px solid var(--border-strong); border-radius: var(--r-md);
}
.signin:hover { border-color: var(--text-muted); color: var(--text); text-decoration: none; }
@media (max-width: 900px) {
  .inner { flex-wrap: wrap; }
  .nav { order: 3; flex-basis: 100%; }
}
</style>
