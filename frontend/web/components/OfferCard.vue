<script setup lang="ts">
import type { Offer } from '~/types/api'

const props = defineProps<{ offer: Offer }>()

const money = (v: number) => v.toLocaleString('ru-RU').replace(/ /g, ' ') + ' ₽'
const date = (iso: string) =>
  new Date(iso).toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' })

const where = computed(() =>
  ({ 'on-car': 'На автомобиле', 'on-storage': 'На хранении' }[props.offer.whereabouts!] ?? null))
const site = { gosnomeru: 'gosnomeru.com', autonomera: 'autonomera777.ru', anomera: 'anomera6.ru' }
</script>

<template>
  <article class="offer">
    <div class="top">
      <p class="price">{{ money(offer.price) }}</p>
      <span class="src">{{ site[offer.provider] }}</span>
    </div>

    <BundleTag :reissue="offer.reissue_included" />

    <!-- Есть значение — показываем, нет — строки не существует -->
    <dl>
      <dt>Опубликовано</dt><dd>{{ date(offer.posted_at) }}</dd>

      <template v-if="offer.refreshed_at !== offer.posted_at">
        <dt>Обновлено</dt><dd>{{ date(offer.refreshed_at) }}</dd>
      </template>

      <template v-if="where">
        <dt>Где номер</dt><dd>{{ where }}</dd>
      </template>

      <template v-if="offer.views">
        <dt>Просмотров</dt><dd>{{ offer.views }}</dd>
      </template>
    </dl>

    <p v-if="offer.comment" class="cmt">{{ offer.comment }}</p>

    <a :href="offer.url" target="_blank" rel="noopener noreferrer nofollow" class="go">
      Перейти к продавцу →
    </a>
  </article>
</template>

<style scoped>
.offer {
  background: var(--surface); border: 1px solid var(--border); border-left: 3px solid var(--accent);
  border-radius: var(--r-md); padding: 16px 18px; box-shadow: var(--sh-1);
  display: grid; gap: 11px; justify-items: start;
}
.top { display: flex; justify-content: space-between; align-items: baseline; gap: 14px; width: 100%; flex-wrap: wrap; }
.price { font-size: 23px; font-weight: 700; letter-spacing: -.02em; font-variant-numeric: tabular-nums; }
.src { font-family: var(--font-plate); font-size: 14px; font-weight: 700; text-transform: uppercase; letter-spacing: .08em; color: var(--text-faint); }
dl { margin: 0; display: grid; grid-template-columns: auto 1fr; gap: 5px 14px; font-size: 14.5px; }
dt { color: var(--text-faint); }
dd { margin: 0; }
.cmt { font-size: 14.5px; color: var(--text-muted); background: var(--surface-sunk); padding: 10px 12px; border-radius: var(--r-sm); line-height: 1.45; }
.go {
  min-height: 44px; padding: 11px 20px; display: inline-flex; align-items: center;
  background: var(--surface); color: var(--text); text-decoration: none;
  border: 1px solid var(--border-strong); border-radius: var(--r-md);
  font-size: 16px; font-weight: 600;
}
.go:hover { border-color: var(--text-muted); color: var(--text); text-decoration: none; }
</style>
