<script setup lang="ts">
import type { SearchResponse } from '~/types/api'
import { PATTERN_LIST } from '@shared/patterns'
import type { Filters } from '~/components/SearchFilters.vue'

const route = useRoute()
const router = useRouter()

const q = computed(() => String(route.query.q || ''))
const activePatterns = computed(() => String(route.query.categories || '').split(',').filter(Boolean))
const region = computed(() => String(route.query.region || ''))
const reissue = computed(() => String(route.query.reissue_included || ''))
const priceMax = computed(() => String(route.query.price_max || ''))
const sort = computed(() => String(route.query.sort || 'updated_desc'))

/** Фильтры живут в адресе: выдачу можно переслать ссылкой */
const filters = computed<Filters>({
  get: () => ({
    region: region.value, price_max: priceMax.value, reissue: reissue.value,
    sort: sort.value, pattern: activePatterns.value
  }),
  set: (f) => router.push({
    path: '/search',
    query: {
      ...route.query,
      region: f.region || undefined,
      price_max: f.price_max || undefined,
      reissue_included: f.reissue || undefined,
      categories: f.pattern.length ? f.pattern.join(',') : undefined,
      sort: f.sort !== 'updated_desc' ? f.sort : undefined
    }
  })
})

const { data: res, pending, error, refresh } = await useFetch<SearchResponse>('/api/v1/search', {
  query: computed(() => ({ ...route.query, limit: 24 }))
})

/** Правка одного параметра, остальные сохраняются */
const setParam = (key: string, value: string | null) => {
  const next = { ...route.query }
  if (value) next[key] = value
  else delete next[key]
  router.push({ path: '/search', query: next })
}

const togglePattern = (code: string) => {
  const set = new Set(activePatterns.value)
  set.has(code) ? set.delete(code) : set.add(code)
  setParam('categories', [...set].join(',') || null)
}

/** Какие фильтры сейчас сужают выдачу — их и предлагаем ослабить, когда пусто */
const applied = computed(() => {
  const list: { key: string; label: string }[] = []
  if (region.value) list.push({ key: 'region', label: `выбранный регион` })
  if (priceMax.value) list.push({ key: 'price_max', label: `цена до ${Number(priceMax.value).toLocaleString('ru-RU')} ₽` })
  if (reissue.value) {
    list.push({ key: 'reissue_included', label: reissue.value === 'true' ? 'с переоформлением' : 'без переоформления' })
  }
  for (const code of activePatterns.value) {
    const p = PATTERN_LIST.find(x => x.code === code)
    if (p) list.push({ key: `pattern:${code}`, label: p.label.toLowerCase() })
  }
  return list
})

const dropFilter = (key: string) => {
  if (key.startsWith('pattern:')) return togglePattern(key.slice(8))
  setParam(key, null)
}

const plural = (n: number) => {
  const t = n % 100, o = n % 10
  if (t >= 11 && t <= 14) return 'номеров'
  if (o === 1) return 'номер'
  if (o >= 2 && o <= 4) return 'номера'
  return 'номеров'
}

const heading = computed(() => {
  if (!res.value) return 'Поиск'
  const n = res.value.total
  return `${n} ${plural(n)}` + (q.value ? ` по маске ${q.value}` : '')
})

useHead(() => ({ title: q.value ? `${q.value} — поиск номеров` : 'Поиск номеров' }))
</script>

<template>
  <div class="container page">
    <PlateSearch />

    <SearchFilters v-model="filters" class="page-filters" />

    <h1 class="heading">{{ heading }}</h1>

    <!-- Скелет держит высоту, чтобы страница не прыгала -->
    <div v-if="pending" class="grid" aria-hidden="true">
      <div v-for="i in 6" :key="i" class="skeleton" />
    </div>

    <div v-else-if="error" class="fail">
      <p>Не удалось загрузить результаты. Фильтры сохранены.</p>
      <button type="button" class="btn-primary" @click="refresh()">Повторить</button>
    </div>

    <template v-else-if="res?.items.length">
      <div class="grid">
        <NumberCard v-for="c in res.items" :key="c.number" :card="c" />
      </div>
      <button v-if="res.cursor" type="button" class="more">Показать ещё 24</button>
    </template>

    <div v-else class="empty">
      <h2>Ничего не подошло</h2>

      <template v-if="applied.length">
        <p>Попробуйте убрать одно из условий — возможно, оно и сузило выдачу до нуля.</p>
        <div class="drop">
          <button v-for="f in applied" :key="f.key" type="button" class="chip drop-chip" @click="dropFilter(f.key)">
            Убрать: {{ f.label }}
          </button>
        </div>
      </template>
      <p v-else>
        По этой маске сейчас нет ни одного объявления. Это обычное дело:
        в продаже находится меньше процента возможных комбинаций.
      </p>

      <NuxtLink to="/triggers" class="btn-primary">Следить за этим запросом</NuxtLink>
    </div>
  </div>
</template>

<style scoped>
.page { padding: 28px 24px 56px; }

.page-filters { margin: 26px 0 30px; }

.chip {
  font-size: 14.5px; font-weight: 500; padding: 8px 14px; min-height: 44px; cursor: pointer;
  border: 1px solid var(--border-strong); border-radius: var(--r-sm);
  background: var(--surface); color: var(--text-muted);
}
.chip[aria-pressed='true'] { border-color: var(--accent); background: var(--accent-sunk); color: var(--accent); font-weight: 600; }

.heading { font-size: 24px; margin-bottom: 18px; }

.grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 16px; }
.skeleton {
  height: 172px; border-radius: var(--r-lg);
  background: var(--surface); border: 1px solid var(--border);
}

.more, .btn-primary {
  min-height: 44px; padding: 11px 20px; cursor: pointer;
  border-radius: var(--r-md); font-size: 16px; font-weight: 600;
  display: inline-flex; align-items: center; text-decoration: none;
}
.more { margin-top: 24px; background: var(--surface); color: var(--text); border: 1px solid var(--border-strong); }
.more:hover { border-color: var(--text-muted); }
.btn-primary { background: var(--accent); color: #fff; border: 0; }
.btn-primary:hover { background: var(--accent-hover); color: #fff; text-decoration: none; }

.fail { display: grid; gap: 14px; justify-items: start; padding: 32px 0; color: var(--alert); }

.empty {
  background: var(--surface); border: 1px dashed var(--border-strong); border-radius: var(--r-lg);
  padding: 32px 28px; display: grid; gap: 16px; justify-items: center; text-align: center;
}
.empty p { color: var(--text-muted); max-width: 560px; }
.drop { display: flex; flex-wrap: wrap; gap: 8px; justify-content: center; }
.drop-chip { color: var(--accent); border-color: var(--accent); }
</style>
