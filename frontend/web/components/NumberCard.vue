<script setup lang="ts">
import type { NumberCard } from '~/types/api'

const props = defineProps<{ card: NumberCard }>()

const money = (v: number) => v.toLocaleString('ru-RU').replace(/ /g, ' ') + ' ₽'
const updated = computed(() =>
  new Date(props.card.updated_at).toLocaleDateString('ru-RU', { day: 'numeric', month: 'long' }))

const many = computed(() => props.card.offers_count > 1)

/* У многоофферной карточки подпись говорит о комплектации той цены, что показана */
const hint = computed(() =>
  many.value && props.card.reissue_included === false ? 'Дешевле — без переоформления' : null)
</script>

<template>
  <article class="card">
    <div class="head">
      <PlateNumber :number="card.number" size="md" />
      <p class="price">
        <span v-if="many" class="from">от</span>{{ money(card.price) }}
      </p>
    </div>

    <div class="sub">
      <BundleTag :reissue="card.reissue_included" />
      <span>{{ card.region.name }}</span>
    </div>

    <p v-if="hint" class="hint">{{ hint }}</p>

    <div class="foot">
      <span class="meta">
        <template v-if="many">{{ card.offers_count }} предложения · </template>
        обновлено {{ updated }}
      </span>
      <NuxtLink :to="`/plates/${card.id}`" class="go">
        {{ many ? 'Сравнить цены' : 'Смотреть объявление' }} →
      </NuxtLink>
    </div>
  </article>
</template>

<style scoped>
.card {
  background: var(--surface); border: 1px solid var(--border); border-radius: var(--r-lg);
  padding: 18px 20px; box-shadow: var(--sh-1);
  display: grid; gap: 12px; align-content: start;
}
.head { display: flex; justify-content: space-between; align-items: flex-start; gap: 16px; flex-wrap: wrap; }
.price {
  font-size: 26px; font-weight: 700; letter-spacing: -.02em; line-height: 1.1;
  font-variant-numeric: tabular-nums; white-space: nowrap;
}
.from { font-size: 15px; font-weight: 500; color: var(--text-muted); margin-right: .3em; }
.sub { display: flex; flex-wrap: wrap; gap: 8px 14px; align-items: center; font-size: 14.5px; color: var(--text-muted); }
.hint { font-size: 14.5px; color: var(--alert); }
.foot {
  display: flex; justify-content: space-between; align-items: center; gap: 14px; flex-wrap: wrap;
  border-top: 1px solid var(--border); padding-top: 12px;
}
.meta { font-size: 14.5px; color: var(--text-faint); }
.go { font-weight: 600; font-size: 15px; }
</style>
