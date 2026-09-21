<script setup lang="ts">
import { sanitizePlateInput } from '@shared/plate'

/** Одно поле принимает и полный номер, и маску: конструктор из селектов не берём */
const props = withDefaults(defineProps<{
  /** search — в выдачу; estimate — сразу на страницу номера */
  mode?: 'search' | 'estimate'
  label?: string
  cta?: string
}>(), { mode: 'search' })

const route = useRoute()
const router = useRouter()

/* В режиме поиска поле показывает маску из адреса: переход по «Поиск» в шапке его очищает */
const fromRoute = () => (props.mode === 'search' ? String(route.query.q || '') : '')
const query = ref(fromRoute())
watch(() => route.query.q, () => { query.value = fromRoute() })

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

  /* Фильтры остаются в адресе выдачи, меняется только маска */
  const rest = route.path === '/' ? route.query : {}
  router.push({ path: '/', query: { ...rest, q: q || undefined } })
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
        placeholder="А777АА77 или А*7*АА77"
        aria-describedby="q-hint"
        @input="onInput"
      >
    </div>
    <button type="submit" class="submit">{{ cta ?? 'Найти' }}</button>

    <p id="q-hint" class="hint">
      Неизвестные знаки заменяйте звёздочкой: <code>А*7*АА77</code>.
      Латинская раскладка переводится сама.
    </p>

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

@media (max-width: 620px) {
  .search { grid-template-columns: 1fr; }
  .submit { width: 100%; }
}
</style>
