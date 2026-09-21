<script setup lang="ts">
import type { PlateItem } from '~/types/api'
import { OFFERS, plural } from '@shared/plural'

const props = defineProps<{ card: PlateItem }>()

const money = (v: number) => v.toLocaleString('ru-RU').replace(/ /g, ' ') + ' ₽'
const updated = computed(() =>
  new Date(props.card.updated_at).toLocaleDateString('ru-RU', { day: 'numeric', month: 'long' }))

const many = computed(() => props.card.count > 1)

/* У многоофферной карточки подпись говорит о комплектации той цены, что показана */
const hint = computed(() =>
  many.value && props.card.reissue_included === false ? 'Дешевле — без переоформления' : null)
</script>

<template>
  <!-- Обёртка нужна только как контейнер запросов: полка перестраивается по своей ширине -->
  <div class="hold">
    <NuxtLink :to="`/plates/${card.id}`" class="row">
      <PlateSign :number="card.number" class="sign" />

      <span class="meta">
        <span v-if="card.region" class="region">{{ card.region.name }}</span>
        <span class="line">
          <BundleTag :reissue="card.reissue_included" />
          <span v-if="hint" class="hint">{{ hint }}</span>
        </span>
        <span class="faint">
          <template v-if="many">{{ card.count }} {{ plural(card.count, OFFERS) }} · </template>
          обновлено {{ updated }}
        </span>
      </span>

      <span class="right">
        <span class="price">
          <template v-if="card.price !== null">
            <span v-if="many" class="from">от</span>{{ money(card.price) }}
          </template>
          <template v-else>Цена не указана</template>
        </span>
        <span class="go">{{ many ? 'Сравнить цены' : 'Смотреть объявление' }} →</span>
      </span>
    </NuxtLink>
  </div>
</template>

<style scoped>
.hold { container-type: inline-size; display: grid; }

.row {
  background: var(--surface); border: 1px solid var(--border); border-radius: var(--r-lg);
  padding: 20px 24px; box-shadow: var(--sh-1); color: var(--text); text-decoration: none;
  display: grid; grid-template-columns: auto 1fr auto; grid-template-areas: 'sign meta price';
  align-items: center; gap: 12px 28px;
}
.row:hover { border-color: var(--border-strong); box-shadow: var(--sh-2); color: var(--text); text-decoration: none; }
.row:hover .go { color: var(--accent-hover); text-decoration: underline; }

.sign { grid-area: sign; justify-self: start; }

.meta { grid-area: meta; display: grid; gap: 8px; justify-items: start; }
.line { display: flex; flex-wrap: wrap; gap: 8px 12px; align-items: center; }
.region { font-size: 16px; color: var(--text-muted); }
.hint { font-size: 14.5px; color: var(--alert); }
.faint { font-size: 14.5px; color: var(--text-faint); }

.right { grid-area: price; display: grid; gap: 8px; justify-items: end; text-align: right; }
.price {
  font-size: 26px; font-weight: 700; letter-spacing: -.02em; line-height: 1.1;
  font-variant-numeric: tabular-nums; white-space: nowrap;
}
.from { font-size: 15px; font-weight: 500; color: var(--text-muted); margin-right: .3em; }
.go { font-size: 15px; font-weight: 600; color: var(--accent); white-space: nowrap; }

/* Знак не ужимается — при нехватке ширины на свою строку уходит он, а следом и цена */
@container (max-width: 760px) {
  .row {
    grid-template-columns: 1fr auto; grid-template-areas: 'sign sign' 'meta price';
    padding: 16px 18px; gap: 16px; align-items: end;
  }
}
@container (max-width: 440px) {
  .row { grid-template-columns: 1fr; grid-template-areas: 'sign' 'meta' 'price'; gap: 14px; }
  .right { justify-items: start; text-align: left; }
}
</style>
