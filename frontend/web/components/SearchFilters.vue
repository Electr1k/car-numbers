<script setup lang="ts">
import type { RegionsResponse } from '~/types/api'
import { PATTERN_LIST } from '@shared/patterns'

export interface Filters {
  region: string
  price_min: string
  price_max: string
  reissue: string
  sort: string
  pattern: string[]
}

const model = defineModel<Filters>({ required: true })
const uid = useId()

const REISSUE = [
  { code: 'true',  label: 'С переоформлением' },
  { code: 'false', label: 'Без переоформления' }
]
const SORTS = [
  { code: 'updated_desc', label: 'Сначала свежие' },
  { code: 'price_asc',    label: 'Сначала дешёвые' },
  { code: 'price_desc',   label: 'Сначала дорогие' }
]

/* Регионы приходят из справочника уже в алфавитном порядке */
const { data: regions } = await useFetch<RegionsResponse>('/api/v1/regions')

const digitPatterns = PATTERN_LIST.filter(p => p.group === 'digits')
const letterPatterns = PATTERN_LIST.filter(p => p.group === 'letters')

const toggle = (code: string) => {
  const set = new Set(model.value.pattern)
  set.has(code) ? set.delete(code) : set.add(code)
  model.value = { ...model.value, pattern: [...set] }
}

const set = (key: keyof Filters, value: string) => {
  model.value = { ...model.value, [key]: value }
}

const region = computed({
  get: () => model.value.region,
  set: (v: string) => set('region', v)
})

/* Цена вводится руками; разряды разделяем пробелом только когда поле оставили */
const digits = (v: string) => v.replace(/\D/g, '').slice(0, 12)
const money = (v: string) => (v ? Number(v).toLocaleString('ru-RU') : '')

const shownPrice = reactive({
  price_min: money(model.value.price_min),
  price_max: money(model.value.price_max)
})

watch(() => [model.value.price_min, model.value.price_max] as const, ([min, max]) => {
  shownPrice.price_min = money(min)
  shownPrice.price_max = money(max)
})

const onPriceInput = (key: 'price_min' | 'price_max', event: Event) => {
  shownPrice[key] = digits((event.target as HTMLInputElement).value)
}

/** Границы, введённые наоборот, меняем местами: иначе выдача всегда пуста */
const commitPrice = (key: 'price_min' | 'price_max') => {
  const next = { ...model.value, [key]: digits(shownPrice[key]) }
  if (next.price_min && next.price_max && Number(next.price_min) > Number(next.price_max)) {
    [next.price_min, next.price_max] = [next.price_max, next.price_min]
  }
  shownPrice.price_min = money(next.price_min)
  shownPrice.price_max = money(next.price_max)
  model.value = next
}
</script>

<template>
  <div class="filters">
    <div class="selects">
      <RegionSelect v-model="region" :items="regions?.items ?? []" />

      <div class="price" role="group" :aria-labelledby="`${uid}-price`">
        <span :id="`${uid}-price`">Цена, ₽</span>
        <div class="price-row">
          <input
            :value="shownPrice.price_min" class="money" type="text" inputmode="numeric"
            autocomplete="off" placeholder="от" aria-label="Цена от, рублей"
            @input="onPriceInput('price_min', $event)"
            @change="commitPrice('price_min')" @blur="commitPrice('price_min')"
            @keydown.enter="commitPrice('price_min')"
          >
          <span class="dash" aria-hidden="true">—</span>
          <input
            :value="shownPrice.price_max" class="money" type="text" inputmode="numeric"
            autocomplete="off" placeholder="до" aria-label="Цена до, рублей"
            @input="onPriceInput('price_max', $event)"
            @change="commitPrice('price_max')" @blur="commitPrice('price_max')"
            @keydown.enter="commitPrice('price_max')"
          >
        </div>
      </div>

      <label class="sel">
        <span>Переоформление</span>
        <select :value="model.reissue" @change="set('reissue', ($event.target as HTMLSelectElement).value)">
          <option value="">Неважно</option>
          <option v-for="r in REISSUE" :key="r.code" :value="r.code">{{ r.label }}</option>
        </select>
      </label>

      <label class="sel">
        <span>Сортировка</span>
        <select :value="model.sort" @change="set('sort', ($event.target as HTMLSelectElement).value)">
          <option v-for="s in SORTS" :key="s.code" :value="s.code">{{ s.label }}</option>
        </select>
      </label>
    </div>

    <fieldset class="group">
      <legend>Узор цифр</legend>
      <div class="chips">
        <button
          v-for="p in digitPatterns" :key="p.code" type="button" class="chip"
          :aria-pressed="model.pattern.includes(p.code)" @click="toggle(p.code)"
        >{{ p.label }}</button>
      </div>
    </fieldset>

    <fieldset class="group">
      <legend>Буквы и регион</legend>
      <div class="chips">
        <button
          v-for="p in letterPatterns" :key="p.code" type="button" class="chip"
          :aria-pressed="model.pattern.includes(p.code)" @click="toggle(p.code)"
        >{{ p.label }}</button>
      </div>
    </fieldset>
  </div>
</template>

<style scoped>
.filters { display: grid; gap: 20px; }
.selects { display: flex; flex-wrap: wrap; align-items: end; gap: 14px; }
.sel { display: grid; gap: 6px; }
.sel > span { font-size: 14px; font-weight: 500; color: var(--text-muted); }
select {
  font: inherit; font-size: 15.5px; min-height: 44px; padding: 10px 12px; cursor: pointer;
  background: var(--surface); color: var(--text);
  border: 1px solid var(--border-strong); border-radius: var(--r-md);
}

.price { display: grid; gap: 6px; }
.price > span { font-size: 14px; font-weight: 500; color: var(--text-muted); }
.price-row { display: flex; align-items: center; gap: 8px; }
.money {
  font: inherit; font-size: 15.5px; font-variant-numeric: tabular-nums;
  width: 116px; min-height: 44px; padding: 10px 12px;
  background: var(--surface); color: var(--text);
  border: 1px solid var(--border-strong); border-radius: var(--r-md);
}
.money::placeholder { color: var(--text-faint); }
.money:focus { outline: none; border-color: var(--accent); box-shadow: 0 0 0 3px var(--accent-sunk); }
.dash { color: var(--text-faint); }

.group { border: 0; margin: 0; padding: 0; display: grid; gap: 8px; }
.group legend {
  padding: 0; font-size: 12px; font-weight: 700; text-transform: uppercase;
  letter-spacing: .09em; color: var(--text-faint);
}
.chips { display: flex; flex-wrap: wrap; gap: 8px; }
.chip {
  font-size: 14.5px; font-weight: 500; padding: 8px 14px; min-height: 44px; cursor: pointer;
  border: 1px solid var(--border-strong); border-radius: var(--r-sm);
  background: var(--surface); color: var(--text-muted);
}
.chip[aria-pressed='true'] {
  border-color: var(--accent); background: var(--accent-sunk); color: var(--accent); font-weight: 600;
}
</style>
