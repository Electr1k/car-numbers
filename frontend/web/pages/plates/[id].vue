<script setup lang="ts">
import type { PlateDetail, PlatesResponse, Valuation } from '~/types/api'
import { OFFERS, plural } from '@shared/plural'

const route = useRoute()
const id = computed(() => String(route.params.id))
const event = useRequestEvent()

/* Страница индексируется: номер и объявления рендерятся на сервере */
const { data: plate, pending, error } = await useApi<PlateDetail>(() => `/api/v1/plates/${id.value}`, {
  lazy: true
})

/* Поисковику — настоящий статус, а не 200 со страницей ошибки */
if (event && error.value) setResponseStatus(event, error.value.statusCode ?? 502)

/**
 * Оценка живёт отдельным эндпоинтом, поэтому запрашивается вторым запросом.
 * Её отсутствие — штатное состояние: объявления показываем и без неё.
 */
const { data: valuation, refresh: refreshValuation } = useApi<Valuation>('/api/v1/valuation', {
  query: computed(() => ({ number: plate.value?.number ?? '' })),
  immediate: false,
  lazy: true,
  server: false,
  default: () => null
})

/* Похожие в продаже core-service не отдаёт — добираем выдачей по региону */
const { data: similar, refresh: refreshSimilar } = useApi<PlatesResponse>('/api/v1/plates', {
  query: computed(() => ({ region_id: plate.value?.region?.id, limit: 4 })),
  immediate: false,
  lazy: true,
  server: false,
  default: () => ({ items: [], next_cursor: null })
})

const similarItems = computed(() =>
  (similar.value?.items ?? []).filter(c => c.id !== id.value).slice(0, 3))

const loading = computed(() => !error.value && (pending.value || !plate.value))

/* Оценка и похожие зависят от номера и региона, поэтому ждут первый ответ */
const valuationDone = ref(false)
const similarDone = ref(false)

watch(plate, (p) => {
  if (!p || import.meta.server) return

  refreshValuation().finally(() => { valuationDone.value = true })

  if (p.region) refreshSimilar().finally(() => { similarDone.value = true })
  else similarDone.value = true
}, { immediate: true })

const archive = computed(() => plate.value?.archive_offers ?? [])
const hasActive = computed(() => (plate.value?.active_offers?.length ?? 0) > 0)

const money = (v: number) => v.toLocaleString('ru-RU').replace(/ /g, ' ') + ' ₽'
const monthYear = (iso: string) =>
  new Date(iso).toLocaleDateString('ru-RU', { month: 'long', year: 'numeric' })

/** Последнее, что видели: цена свежайшего архивного объявления */
const lastSeen = computed(() => {
  if (!archive.value.length) return null
  const sorted = [...archive.value].sort((a, b) => b.posted_at.localeCompare(a.posted_at))
  return sorted[0]
})

/* Цена трёхлетней давности — факт, а не ориентир: рынок с тех пор ушёл */
const archiveIsStale = computed(() => {
  const at = lastSeen.value?.posted_at
  return at ? (Date.now() - new Date(at).getTime()) / 31_557_600_000 >= 1 : false
})

const restriction = computed(() => plate.value?.region
  ? `Номер с кодом ${plate.value.region.code} можно поставить только на автомобиль, зарегистрированный в этом регионе.`
  : '')

useHead(() => ({
  title: plate.value ? `${plate.value.number} — сколько стоит номер` : 'Номер'
}))
</script>

<template>
  <div class="container page">
    <p v-if="error" class="fail">
      Не удалось загрузить страницу номера. Обновите страницу или попробуйте позже.
    </p>

    <template v-else>
      <nav class="crumb" aria-label="Хлебные крошки">
        <NuxtLink to="/">Поиск</NuxtLink>
        <span aria-hidden="true">→</span>
        <template v-if="plate?.region">
          <span>{{ plate.region.name }}</span>
          <span aria-hidden="true">→</span>
        </template>
        <span v-if="plate" class="cur">{{ plate.number }}</span>
        <span v-else class="sk crumb-sk" aria-hidden="true" />
      </nav>

      <div class="cols">
        <!-- Левая колонка одинакова во всех состояниях -->
        <div class="left">
          <template v-if="plate">
            <h1 class="visually-hidden">Номер {{ plate.number }}</h1>
            <PlateTilt :number="plate.number" />
          </template>
          <span v-else class="sk plate-sk" aria-hidden="true" />

          <EstimateCard :estimate="valuation" :pending="loading || !valuationDone" />
        </div>

        <div class="right">
          <template v-if="loading">
            <div class="sect">
              <span class="sk head-sk" aria-hidden="true" />
            </div>
            <div class="stack" aria-busy="true">
              <SkeletonOffer v-for="i in 2" :key="i" />
            </div>
          </template>

          <template v-else-if="hasActive">
            <div class="sect">
              <h2>{{ plate.active_offers.length }} {{ plural(plate.active_offers.length, OFFERS) }}</h2>
              <span class="note">Цены отличаются тем, что в них входит</span>
            </div>
            <div class="stack">
              <OfferCard v-for="o in plate.active_offers" :key="o.id" :offer="o" />
            </div>
          </template>

          <template v-else>
            <div class="empty">
              <h2>Сейчас в продаже нет</h2>
              <p>Мы знаем этот номер и сообщим, как только он снова появится.</p>
              <NuxtLink to="/triggers" class="cta">Сообщить, когда появится</NuxtLink>
            </div>

            <div v-if="lastSeen" class="arch">
              <p class="arch-lbl">Продавался раньше</p>
              <p class="arch-val">
                {{ monthYear(lastSeen.posted_at) }}<template v-if="lastSeen.price !== null"> — {{ money(lastSeen.price) }}</template>
              </p>
              <p class="arch-meta">
                Всего объявлений в архиве: {{ archive.length }}
              </p>
              <p v-if="archiveIsStale" class="arch-note">
                Цена не сегодняшняя: с тех пор рынок ушёл вверх, и ориентироваться
                на неё не стоит. Актуальную цену показывает оценка слева.
              </p>
            </div>
          </template>

          <aside v-if="restriction" class="restrict">
            {{ restriction }}
            <NuxtLink to="/reissue">Как это устроено</NuxtLink>
          </aside>
        </div>
      </div>

      <!-- Похожие идут во всю ширину: под левой колонкой всё равно пусто -->
      <section v-if="plate && (!similarDone || similarItems.length)" class="similar">
        <div class="sect">
          <h2>Похожие в продаже</h2>
          <NuxtLink v-if="plate?.region" :to="`/?region=${plate.region.id}`">
            Все в регионе {{ plate.region.code }} →
          </NuxtLink>
        </div>
        <div class="grid" :aria-busy="!similarDone">
          <template v-if="!similarDone">
            <SkeletonCard v-for="i in 3" :key="i" />
          </template>
          <PlateCard v-for="c in similarItems" v-else :key="c.id" :card="c" />
        </div>
      </section>
    </template>
  </div>
</template>

<style scoped>
.page { padding: 26px 24px 56px; }
.fail { padding: 40px 0; color: var(--alert); }

.crumb { display: flex; flex-wrap: wrap; gap: 8px; font-size: 14.5px; color: var(--text-faint); margin-bottom: 20px; }
.crumb .cur { color: var(--text-muted); }
.crumb-sk { display: inline-block; width: 104px; height: 15px; }
.plate-sk { display: block; width: 340px; max-width: 100%; height: 92px; border-radius: 9px; }
.head-sk { display: block; width: 168px; height: 24px; }

.cols { display: grid; grid-template-columns: 2fr 3fr; gap: 32px; align-items: start; }
.left { display: grid; gap: 20px; justify-items: start; align-content: start; }
.left > * { max-width: 100%; }
.left :deep(.est) { width: 100%; }
.right { display: grid; gap: 20px; align-content: start; }

.sect { display: flex; justify-content: space-between; align-items: baseline; gap: 16px; flex-wrap: wrap; margin-bottom: 14px; }
.note { font-size: 15px; color: var(--text-muted); }
.stack { display: grid; gap: 14px; }
.similar { margin-top: 36px; }
/* 320px — минимум, в который знак помещается со своими полями; на 1180px это ровно три колонки */
.grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 16px; }

.empty {
  background: var(--surface); border: 1px dashed var(--border-strong); border-radius: var(--r-lg);
  padding: 28px 26px; display: grid; gap: 14px; justify-items: center; text-align: center;
}
.empty p { color: var(--text-muted); max-width: 520px; }
.cta {
  min-height: 44px; padding: 11px 20px; display: inline-flex; align-items: center;
  border: 0; border-radius: var(--r-md);
  background: var(--accent); color: #fff; font-size: 16px; font-weight: 600; text-decoration: none;
}
.cta:hover { background: var(--accent-hover); color: #fff; text-decoration: none; }

.arch { background: var(--surface-sunk); border-radius: var(--r-lg); padding: 20px 22px; display: grid; gap: 8px; }
.arch-lbl { font-size: 13px; font-weight: 700; text-transform: uppercase; letter-spacing: .1em; color: var(--text-faint); }
.arch-val { font-size: 22px; font-weight: 700; letter-spacing: -.02em; }
.arch-meta { font-size: 14.5px; color: var(--text-faint); }
.arch-note { font-size: 15px; color: var(--text-muted); }

.restrict {
  background: var(--warn-bg); color: var(--warn); border-radius: var(--r-md);
  padding: 16px 18px; font-size: 15px; font-weight: 500;
}
.restrict a { color: var(--warn); text-decoration: underline; margin-left: 6px; }

@media (max-width: 900px) {
  .cols { grid-template-columns: 1fr; gap: 24px; }
}
</style>
