<script setup lang="ts">
import type { PlatesResponse, Valuation } from '~/types/api'

const route = useRoute()
const number = computed(() => String(route.params.number).toUpperCase())
const event = useRequestEvent()

/* Страница индексируется: оценка рендерится на сервере */
const { data: valuation, pending, error } = await useApi<Valuation>('/api/v1/valuation', {
  query: computed(() => ({ number: number.value })),
  lazy: true
})

/* Поисковику — настоящий статус, а не 200 со страницей отказа */
if (event && error.value) setResponseStatus(event, error.value.statusCode ?? 502)

/* Оценка есть для любой комбинации, но в продаже её обычно нет — предлагаем похожие */
const { data: similar, refresh: refreshSimilar } = useApi<PlatesResponse>('/api/v1/plates', {
  query: computed(() => ({ region_id: valuation.value?.region?.id, limit: 3 })),
  immediate: false,
  lazy: true,
  server: false,
  default: () => ({ items: [], next_cursor: null })
})

const similarDone = ref(false)

watch(valuation, (v) => {
  if (!v || import.meta.server) return
  if (v.region) refreshSimilar().finally(() => { similarDone.value = true })
  else similarDone.value = true
}, { immediate: true })

const refusal = computed(() => {
  const body = error.value?.data as { error?: { code: string; message: string } } | undefined
  return body?.error ?? null
})

useHead(() => ({ title: `${number.value} — сколько стоит номер` }))
</script>

<template>
  <div class="container page">
    <section v-if="refusal" class="refusal">
      <h1>Оценить не получится</h1>
      <p>{{ refusal.message }}</p>
      <NuxtLink to="/estimate" class="back">Попробовать другой номер</NuxtLink>
    </section>

    <p v-else-if="error" class="fail">
      Не удалось получить оценку. Обновите страницу или попробуйте позже.
    </p>

    <template v-else>
      <nav class="crumb" aria-label="Хлебные крошки">
        <NuxtLink to="/estimate">Оценка номера</NuxtLink>
        <span aria-hidden="true">→</span>
        <span class="cur">{{ number }}</span>
      </nav>

      <div class="cols">
        <div class="left">
          <h1 class="visually-hidden">Оценка номера {{ number }}</h1>
          <PlateSign :number="number" />
          <EstimateCard :estimate="valuation" :pending="pending || !valuation" />
        </div>

        <div class="right">
          <div class="empty">
            <h2>В продаже сейчас нет</h2>
            <p>
              Оценка слева построена по объявлениям трёх площадок и работает
              для любого номера. Как только этот появится в продаже — сообщим.
            </p>
            <NuxtLink to="/triggers" class="cta">Сообщить, когда появится</NuxtLink>
          </div>

          <aside v-if="valuation?.region" class="restrict">
            Номер с кодом {{ valuation.region.code }} можно поставить только на автомобиль,
            зарегистрированный в этом регионе.
            <NuxtLink to="/reissue">Как это устроено</NuxtLink>
          </aside>
        </div>
      </div>

      <!-- Похожие идут во всю ширину: под левой колонкой всё равно пусто -->
      <section v-if="valuation && (!similarDone || similar?.items.length)" class="similar">
        <div class="sect">
          <h2>Похожие в продаже</h2>
          <NuxtLink v-if="valuation?.region" :to="`/?region=${valuation.region.id}`">
            Все в регионе {{ valuation.region.code }} →
          </NuxtLink>
        </div>
        <div class="grid" :aria-busy="!similarDone">
          <template v-if="!similarDone">
            <SkeletonCard v-for="i in 3" :key="i" />
          </template>
          <PlateCard v-for="c in similar.items" v-else :key="c.id" :card="c" />
        </div>
      </section>
    </template>
  </div>
</template>

<style scoped>
.page { padding: 26px 24px 56px; }
.fail { padding: 40px 0; color: var(--alert); }

.refusal { padding: 48px 0 64px; display: grid; gap: 14px; justify-items: start; max-width: 620px; }
.refusal p { color: var(--text-muted); font-size: 17px; }
.back {
  min-height: 44px; padding: 11px 20px; display: inline-flex; align-items: center; margin-top: 6px;
  background: var(--accent); color: #fff; border-radius: var(--r-md); font-weight: 600;
}
.back:hover { background: var(--accent-hover); color: #fff; text-decoration: none; }

.crumb { display: flex; flex-wrap: wrap; gap: 8px; font-size: 14.5px; color: var(--text-faint); margin-bottom: 20px; }
.crumb .cur { color: var(--text-muted); }

.cols { display: grid; grid-template-columns: 2fr 3fr; gap: 32px; align-items: start; }
.left { display: grid; gap: 20px; justify-items: start; align-content: start; }
.left > * { max-width: 100%; }
.left :deep(.est) { width: 100%; }
.right { display: grid; gap: 20px; align-content: start; }

.sect { display: flex; justify-content: space-between; align-items: baseline; gap: 16px; flex-wrap: wrap; margin-bottom: 14px; }
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
  border-radius: var(--r-md); background: var(--accent); color: #fff;
  font-size: 16px; font-weight: 600; text-decoration: none;
}
.cta:hover { background: var(--accent-hover); color: #fff; text-decoration: none; }

.restrict {
  background: var(--warn-bg); color: var(--warn); border-radius: var(--r-md);
  padding: 16px 18px; font-size: 15px; font-weight: 500;
}
.restrict a { color: var(--warn); text-decoration: underline; margin-left: 6px; }

@media (max-width: 900px) { .cols { grid-template-columns: 1fr; gap: 24px; } }
</style>
