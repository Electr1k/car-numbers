<script setup lang="ts">
import type { PlateItem, PlatesResponse } from '~/types/api'
import { categoryLabel, knownCategories, type CategoryId } from '@shared/categories'
import type { Filters } from '~/components/SearchFilters.vue'

const route = useRoute()
const router = useRouter()

const q = computed(() => String(route.query.q || ''))
const activeCategories = computed(() => knownCategories(String(route.query.categories || '').split(',')))
const region = computed(() => String(route.query.region || ''))
const reissue = computed(() => String(route.query.reissue_included || ''))
const priceMin = computed(() => String(route.query.price_min || ''))
const priceMax = computed(() => String(route.query.price_max || ''))
const sort = computed(() => String(route.query.sort || 'updated_desc'))

/** Фильтры живут в адресе: выдачу можно переслать ссылкой */
const filters = computed<Filters>({
  get: () => ({
    region: region.value, price_min: priceMin.value, price_max: priceMax.value, reissue: reissue.value,
    sort: sort.value, categories: activeCategories.value
  }),
  set: (f) => router.push({
    path: '/',
    query: {
      ...route.query,
      region: f.region || undefined,
      price_min: f.price_min || undefined,
      price_max: f.price_max || undefined,
      reissue_included: f.reissue || undefined,
      categories: f.categories.length ? f.categories.join(',') : undefined,
      sort: f.sort !== 'updated_desc' ? f.sort : undefined
    }
  })
})

const PAGE = 20

const { $api } = useNuxtApp()

/** Адрес страницы — для людей и ссылок, запрос — в терминах core-service */
const apiQuery = computed(() => ({
  query: q.value || undefined,
  region_id: Number(region.value) || undefined,
  price_from: Number(priceMin.value) || undefined,
  price_to: Number(priceMax.value) || undefined,
  reissue_included: ['true', 'false'].includes(reissue.value) ? reissue.value : undefined,
  category_ids: activeCategories.value.length ? activeCategories.value : undefined,
  sort: sort.value !== 'updated_desc' ? sort.value : undefined,
  limit: PAGE
}))

const { data: res, pending, error, refresh } = useApi<PlatesResponse>('/api/v1/plates', {
  query: apiQuery,
  lazy: true,
  server: false
})

/* Ручное обновление не прячет ленту за скелетом: крутится только иконка */
const reloading = ref(false)
const loading = computed(() => !error.value && ((pending.value && !reloading.value) || !res.value))

const reload = async () => {
  if (reloading.value) return
  reloading.value = true
  try {
    await refresh()
  } finally {
    reloading.value = false
  }
}

/* Догруженные страницы живут отдельно: первая приходит из useFetch и сбрасывает их при смене фильтров */
const loaded = ref<PlateItem[]>([])
const cursor = ref<string | null>(null)
const loadingMore = ref(false)
const moreError = ref(false)

watch(res, (r) => {
  loaded.value = []
  cursor.value = r?.next_cursor ?? null
  moreError.value = false
}, { immediate: true })

const items = computed(() => [...(res.value?.items ?? []), ...loaded.value])

const loadMore = async () => {
  if (!cursor.value || loadingMore.value) return

  const base = res.value
  loadingMore.value = true
  moreError.value = false
  try {
    const next = await $api<PlatesResponse>('/api/v1/plates', {
      query: { ...apiQuery.value, cursor: cursor.value }
    })
    // Пока грузили, фильтры сменились: эта страница уже чужая
    if (res.value !== base) return
    loaded.value = [...loaded.value, ...next.items]
    cursor.value = next.next_cursor
  } catch {
    moreError.value = true
  } finally {
    loadingMore.value = false
  }
}

/* Лента догружается сама, когда низ списка подходит к экрану; после ошибки — только по кнопке */
const sentinel = ref<HTMLElement | null>(null)
const nearEnd = ref(false)
let observer: IntersectionObserver | null = null

onMounted(() => {
  observer = new IntersectionObserver(([e]) => { nearEnd.value = !!e?.isIntersecting }, { rootMargin: '600px 0px' })
  if (sentinel.value) observer.observe(sentinel.value)
})
onBeforeUnmount(() => observer?.disconnect())

watch([nearEnd, cursor, loadingMore, moreError], () => {
  if (nearEnd.value && !moreError.value) loadMore()
})

/** Правка одного параметра, остальные сохраняются */
const setParam = (key: string, value: string | null) => {
  const next = { ...route.query }
  if (value) next[key] = value
  else delete next[key]
  router.push({ path: '/', query: next })
}

const toggleCategory = (id: CategoryId) => {
  const set = new Set(activeCategories.value)
  set.has(id) ? set.delete(id) : set.add(id)
  setParam('categories', [...set].join(',') || null)
}

/** Какие фильтры сейчас сужают выдачу — их и предлагаем ослабить, когда пусто */
const applied = computed(() => {
  const list: { key: string; label: string }[] = []
  if (region.value) list.push({ key: 'region', label: `выбранный регион` })
  if (priceMin.value) list.push({ key: 'price_min', label: `цена от ${Number(priceMin.value).toLocaleString('ru-RU')} ₽` })
  if (priceMax.value) list.push({ key: 'price_max', label: `цена до ${Number(priceMax.value).toLocaleString('ru-RU')} ₽` })
  if (reissue.value) {
    list.push({ key: 'reissue_included', label: reissue.value === 'true' ? 'с переоформлением' : 'без переоформления' })
  }
  for (const id of activeCategories.value) {
    const label = categoryLabel(id)
    if (label) list.push({ key: `category:${id}`, label: label.toLowerCase() })
  }
  return list
})

const dropFilter = (key: string) => {
  if (key.startsWith('category:')) return toggleCategory(key.slice(9) as CategoryId)
  setParam(key, null)
}

/* Счётчик не выводим: с догрузкой он рос бы на глазах и вводил в заблуждение */
const heading = computed(() => {
  if (q.value) return `Номера по маске ${q.value}`
  if (applied.value.length) return 'Подходящие номера'
  return 'Свежие предложения'
})

useHead(() => ({ title: q.value ? `${q.value} — поиск номеров` : 'Номерограф — объявления о продаже автономеров' }))
</script>

<template>
  <div>
    <section class="hero container">
      <h1>Найти красивый номер</h1>
      <p class="lede">
        Объявления трёх площадок в одном месте. Ищите по номеру, маске или узору —
        а если нужного нет в продаже, мы всё равно скажем, сколько он стоит.
      </p>
      <PlateSearch />

      <details class="filters" :open="applied.length > 0">
        <summary>Фильтры</summary>
        <div class="filters-body">
          <SearchFilters v-model="filters" />
        </div>
      </details>
    </section>

    <section class="feed container">
      <div class="sect">
        <h2>{{ heading }}</h2>
        <button type="button" class="reload" :disabled="loading || reloading" @click="reload">
          <svg :class="{ spin: reloading }" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor"
               stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="M20 12a8 8 0 1 1-2.34-5.66" />
            <path d="M20 4v5h-5" />
          </svg>
          {{ reloading ? 'Обновляем…' : 'Обновить' }}
        </button>
      </div>

      <!-- Скелет держит высоту, чтобы страница не прыгала -->
      <div v-if="loading" class="grid" aria-busy="true">
        <SkeletonCard v-for="i in 6" :key="i" />
      </div>

      <div v-else-if="error" class="fail">
        <p>Не удалось загрузить предложения. Фильтры сохранены.</p>
        <button type="button" class="btn-primary" @click="refresh()">Повторить</button>
      </div>

      <template v-else-if="items.length">
        <div class="grid" :aria-busy="loadingMore">
          <PlateCard v-for="card in items" :key="card.id" :card="card" />
          <SkeletonCard v-for="i in (loadingMore ? 2 : 0)" :key="`sk-${i}`" />
        </div>
        <div v-if="moreError" class="fail more-fail">
          <p>Не удалось загрузить следующие номера.</p>
          <button type="button" class="more" @click="loadMore">Повторить</button>
        </div>
      </template>

      <div v-else class="empty">
        <h2>Ничего не подошло</h2>

        <template v-if="applied.length">
          <p>Попробуйте убрать одно из условий — возможно, именно оно сузило выдачу до нуля.</p>
          <div class="drop">
            <button v-for="f in applied" :key="f.key" type="button" class="chip drop-chip" @click="dropFilter(f.key)">
              Убрать: {{ f.label }}
            </button>
          </div>
        </template>
        <p v-else-if="q">
          По этой маске сейчас нет ни одного объявления. Это обычное дело:
          в продаже находится меньше процента возможных комбинаций.
        </p>
        <p v-else>Предложений пока нет.</p>

        <NuxtLink v-if="q || applied.length" to="/triggers" class="btn-primary">Следить за этим запросом</NuxtLink>
      </div>

      <div ref="sentinel" aria-hidden="true" />
    </section>
  </div>
</template>

<style scoped>
.hero { padding: 52px 24px 32px; }
.lede { font-size: 18px; color: var(--text-muted); margin: 14px 0 24px; max-width: 620px; }

.filters { max-width: 760px; margin-top: 16px; border-top: 1px solid var(--border); padding-top: 4px; }
.filters summary {
  font-size: 15.5px; font-weight: 600; color: var(--accent); cursor: pointer;
  min-height: 44px; display: flex; align-items: center;
}
.filters-body { padding: 8px 0 4px; }

.feed { padding: 0 24px 56px; }
.sect { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 8px 16px; margin-bottom: 16px; }

.reload {
  display: inline-flex; align-items: center; gap: 8px; min-height: 44px; padding: 8px 14px; cursor: pointer;
  background: var(--surface); color: var(--text-muted);
  border: 1px solid var(--border-strong); border-radius: var(--r-md);
  font-size: 15px; font-weight: 500;
}
.reload:hover:not(:disabled) { color: var(--text); border-color: var(--text-muted); }
.reload:disabled { opacity: .6; cursor: default; }
.spin { animation: spin .8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
@media (prefers-reduced-motion: reduce) { .spin { animation: none; } }

.grid { display: grid; grid-template-columns: 1fr; gap: 12px; }

.chip {
  font-size: 14.5px; font-weight: 500; padding: 8px 14px; min-height: 44px; cursor: pointer;
  border: 1px solid var(--border-strong); border-radius: var(--r-sm);
  background: var(--surface); color: var(--text-muted);
}

.more, .btn-primary {
  min-height: 44px; padding: 11px 20px; cursor: pointer;
  border-radius: var(--r-md); font-size: 16px; font-weight: 600;
  display: inline-flex; align-items: center; text-decoration: none;
}
.more { background: var(--surface); color: var(--text); border: 1px solid var(--border-strong); }
.more:hover { border-color: var(--text-muted); }
.btn-primary { background: var(--accent); color: #fff; border: 0; }
.btn-primary:hover { background: var(--accent-hover); color: #fff; text-decoration: none; }

.fail { display: grid; gap: 14px; justify-items: start; padding: 32px 0; color: var(--alert); }
.more-fail { padding: 24px 0 0; }

.empty {
  background: var(--surface); border: 1px dashed var(--border-strong); border-radius: var(--r-lg);
  padding: 32px 28px; display: grid; gap: 16px; justify-items: center; text-align: center;
}
.empty p { color: var(--text-muted); max-width: 560px; }
.drop { display: flex; flex-wrap: wrap; gap: 8px; justify-content: center; }
.drop-chip { color: var(--accent); border-color: var(--accent); }
</style>
