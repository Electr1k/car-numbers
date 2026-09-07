<script setup lang="ts">
import type { RegionsResponse } from '~/types/api'
import { PATTERN_LIST } from '@shared/patterns'

export interface Filters {
  region: string
  price_max: string
  reissue: string
  sort: string
  pattern: string[]
}

const model = defineModel<Filters>({ required: true })

const REISSUE = [
  { code: 'true',  label: 'С переоформлением' },
  { code: 'false', label: 'Без переоформления' }
]
const SORTS = [
  { code: 'updated_desc', label: 'Сначала свежие' },
  { code: 'price_asc',    label: 'Сначала дешёвые' },
  { code: 'price_desc',   label: 'Сначала дорогие' }
]
const PRICES = [
  { value: '50000',   label: '50 тыс. ₽' },
  { value: '150000',  label: '150 тыс. ₽' },
  { value: '500000',  label: '500 тыс. ₽' },
  { value: '2000000', label: '2 млн ₽' }
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
</script>

<template>
  <div class="filters">
    <div class="selects">
      <label class="sel">
        <span>Регион</span>
        <select :value="model.region" @change="set('region', ($event.target as HTMLSelectElement).value)">
          <option value="">Любой</option>
          <option v-for="r in regions?.items" :key="r.name" :value="r.codes.join(',')">
            {{ r.name }}
          </option>
        </select>
      </label>

      <label class="sel">
        <span>Цена до</span>
        <select :value="model.price_max" @change="set('price_max', ($event.target as HTMLSelectElement).value)">
          <option value="">Без ограничения</option>
          <option v-for="p in PRICES" :key="p.value" :value="p.value">{{ p.label }}</option>
        </select>
      </label>

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
.selects { display: flex; flex-wrap: wrap; gap: 14px; }
.sel { display: grid; gap: 6px; }
.sel > span { font-size: 14px; font-weight: 500; color: var(--text-muted); }
select {
  font: inherit; font-size: 15.5px; min-height: 44px; padding: 10px 12px; cursor: pointer;
  background: var(--surface); color: var(--text);
  border: 1px solid var(--border-strong); border-radius: var(--r-md);
}

.group { border: 0; margin: 0; padding: 0; display: grid; gap: 8px; }
legend {
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
