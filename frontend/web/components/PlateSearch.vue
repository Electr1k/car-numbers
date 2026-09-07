<script setup lang="ts">
import { sanitizePlateInput } from '@shared/plate'
import type { Filters } from './SearchFilters.vue'

/** Одно поле принимает и полный номер, и маску: конструктор из селектов не берём */
const props = withDefaults(defineProps<{
  /** search — в выдачу; estimate — сразу на страницу номера */
  mode?: 'search' | 'estimate'
  /** показывать ли свёрнутый блок фильтров под полем */
  advanced?: boolean
  label?: string
  cta?: string
}>(), { mode: 'search', advanced: false })

const query = ref('')
const router = useRouter()

const filters = ref<Filters>({ region: '', price_max: '', reissue: '', sort: 'updated_desc', pattern: [] })

/**
 * Печатать можно только то, что бывает в номере.
 * Латиница переводится на лету: P становится Р, а не отбрасывается.
 */
const onInput = (event: Event) => {
  const el = event.target as HTMLInputElement
  const before = el.selectionStart ?? el.value.length
  const removed = el.value.length - sanitizePlateInput(el.value).length

  query.value = sanitizePlateInput(el.value)
  nextTick(() => {
    el.value = query.value
    const at = Math.max(0, before - removed)
    el.setSelectionRange(at, at)
  })
}

const submit = () => {
  const q = query.value.trim()

  if (props.mode === 'estimate') {
    if (!q) return
    router.push(`/valuation/${encodeURIComponent(q)}`)
    return
  }

  const f = filters.value
  router.push({
    path: '/search',
    query: {
      q: q || undefined,
      region: f.region || undefined,
      price_max: f.price_max || undefined,
      reissue_included: f.reissue || undefined,
      categories: f.pattern.length ? f.pattern.join(',') : undefined,
      sort: f.sort !== 'updated_desc' ? f.sort : undefined
    }
  })
}
</script>

<template>
  <form class="search" @submit.prevent="submit">
    <div class="field">
      <label for="q">{{ label ?? 'Номер или маска' }}</label>
      <input
        id="q"
        :value="query"
        type="text"
        autocomplete="off"
        autocapitalize="characters"
        spellcheck="false"
        :maxlength="9"
        placeholder="А123ВС777 или А*2*ВС77"
        aria-describedby="q-hint"
        @input="onInput"
      >
    </div>
    <button type="submit" class="submit">{{ cta ?? 'Найти' }}</button>

    <p id="q-hint" class="hint">
      Неизвестные знаки заменяйте звёздочкой: <code>А*2*ВС77</code>.
      Латинская раскладка переводится сама.
    </p>

    <details v-if="advanced" class="more">
      <summary>Дополнительные фильтры</summary>
      <div class="more-body">
        <SearchFilters v-model="filters" />
      </div>
    </details>
  </form>
</template>

<style scoped>
.search { display: grid; grid-template-columns: minmax(0, 1fr) auto; gap: 12px; max-width: 760px; }
.field { display: grid; gap: 6px; }
label { font-size: 14.5px; font-weight: 500; color: var(--text-muted); }
input {
  font-family: var(--font-plate); font-weight: 700; font-size: 26px; letter-spacing: .06em;
  width: 100%; min-height: 56px; padding: 12px 16px;
  border: 2px solid var(--border-strong); border-radius: var(--r-md);
  background: var(--surface); color: var(--text);
}
input::placeholder {
  font-family: var(--font-ui); font-weight: 400; font-size: 19px;
  color: var(--text-faint); letter-spacing: 0;
}
input:focus { outline: none; border-color: var(--accent); box-shadow: 0 0 0 3px var(--accent-sunk); }

.submit {
  align-self: end; min-height: 56px; padding: 11px 24px; border: 0; border-radius: var(--r-md);
  background: var(--accent); color: #fff; font-size: 17px; font-weight: 600; cursor: pointer;
}
.submit:hover { background: var(--accent-hover); }

.hint { grid-column: 1 / -1; font-size: 14px; color: var(--text-faint); }
.hint code { font-family: var(--font-plate); background: var(--surface-sunk); padding: 1px 5px; border-radius: 2px; }

.more { grid-column: 1 / -1; border-top: 1px solid var(--border); padding-top: 4px; }
.more summary {
  font-size: 15.5px; font-weight: 600; color: var(--accent); cursor: pointer;
  min-height: 44px; display: flex; align-items: center;
}
.more-body { padding: 8px 0 4px; }

@media (max-width: 620px) {
  .search { grid-template-columns: 1fr; }
  .submit { width: 100%; }
}
</style>
