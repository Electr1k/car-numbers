<script setup lang="ts">
import type { PlateItem, PlatesResponse } from '~/types/api'

const PAGE = 16

const { $api } = useNuxtApp()

const { data: feed, pending, error } = useApi<PlatesResponse>('/api/v1/plates', {
  query: { limit: PAGE },
  lazy: true,
  server: false
})

const loading = computed(() => !error.value && (pending.value || !feed.value))

const loaded = ref<PlateItem[]>([])
const cursor = ref<string | null>(null)

watch(feed, (f) => { cursor.value = f?.next_cursor ?? null }, { immediate: true })
const loadingMore = ref(false)

const items = computed(() => [...(feed.value?.items ?? []), ...loaded.value])

const loadMore = async () => {
  if (!cursor.value || loadingMore.value) return

  loadingMore.value = true
  try {
    const next = await $api<PlatesResponse>('/api/v1/plates', {
      query: { limit: PAGE, cursor: cursor.value }
    })
    loaded.value = [...loaded.value, ...next.items]
    cursor.value = next.next_cursor
  } finally {
    loadingMore.value = false
  }
}

useHead({ title: 'Номерограф — объявления о продаже автономеров' })
</script>

<template>
  <div>
    <section class="hero container">
      <h1>Найти красивый номер</h1>
      <p class="lede">
        Объявления трёх площадок в одном месте. Ищите по номеру, маске или узору —
        а если нужного нет в продаже, мы всё равно скажем, сколько он стоит.
      </p>
      <PlateSearch advanced />
    </section>

    <section class="feed container">
      <h2 class="sect">Свежие предложения</h2>

      <div v-if="loading" class="grid" aria-busy="true">
        <SkeletonCard v-for="i in 6" :key="i" />
      </div>

      <p v-else-if="error" class="state error">
        Не удалось загрузить предложения. Обновите страницу или попробуйте позже.
      </p>

      <template v-else-if="items.length">
        <div class="grid">
          <PlateCard v-for="card in items" :key="card.id" :card="card" />
          <SkeletonCard v-for="i in (loadingMore ? 2 : 0)" :key="`sk-${i}`" />
        </div>
        <button v-if="cursor" type="button" class="more" :disabled="loadingMore" @click="loadMore">
          {{ loadingMore ? 'Загружаем…' : `Показать ещё ${PAGE}` }}
        </button>
      </template>

      <p v-else class="state">Предложений пока нет.</p>
    </section>
  </div>
</template>

<style scoped>
.hero { padding: 52px 24px 40px; }
.lede { font-size: 18px; color: var(--text-muted); margin: 14px 0 24px; max-width: 620px; }

.feed { padding: 0 24px 56px; }
.sect { margin-bottom: 16px; }

.grid { display: grid; grid-template-columns: 1fr; gap: 12px; }

.more {
  margin-top: 24px; min-height: 44px; padding: 11px 20px; cursor: pointer;
  background: var(--surface); color: var(--text);
  border: 1px solid var(--border-strong); border-radius: var(--r-md);
  font-size: 16px; font-weight: 600;
}
.more:hover { border-color: var(--text-muted); }
.more:disabled { opacity: .6; cursor: default; }

.state { padding: 28px 0; color: var(--text-muted); }
.state.error { color: var(--alert); }
</style>
