<script setup lang="ts">
import type { RegionGroup } from '~/types/api'

/** Нативный select не даёт покрасить коды в сером, поэтому список свой */
const props = withDefaults(defineProps<{
  items: RegionGroup[]
  /** порог, после которого в списке появляется поле поиска */
  searchFrom?: number
}>(), { searchFrom: 10 })

const model = defineModel<string>({ required: true })

const ANY = { value: '', name: 'Любой', codes: [] as number[] }

const open = ref(false)
const filter = ref('')
const active = ref(0)
const root = ref<HTMLElement>()
const search = ref<HTMLInputElement>()
const list = ref<HTMLElement>()
const uid = useId()

const options = computed(() => [
  ANY,
  ...props.items.map(r => ({ value: r.codes.join(','), name: r.name, codes: r.codes }))
])
const searchable = computed(() => options.value.length > props.searchFrom)

/** «Любой» остаётся в списке всегда: иначе выбор не сбросить */
const shown = computed(() => {
  const q = filter.value.trim().toLowerCase()
  if (!q) return options.value
  return options.value.filter(o => !o.value
    || o.name.toLowerCase().includes(q)
    || o.codes.some(c => String(c).startsWith(q)))
})

const current = computed(() => options.value.find(o => o.value === model.value) ?? ANY)

const scrollToActive = () => {
  list.value?.querySelector<HTMLElement>(`[data-idx="${active.value}"]`)
    ?.scrollIntoView({ block: 'nearest' })
}

const show = async () => {
  open.value = true
  filter.value = ''
  active.value = Math.max(0, shown.value.findIndex(o => o.value === model.value))
  await nextTick()
  ;(searchable.value ? search.value : list.value)?.focus()
  scrollToActive()
}

const hide = (focusBack = false) => {
  open.value = false
  if (focusBack) root.value?.querySelector<HTMLElement>('.trigger')?.focus()
}

const pick = (value: string) => {
  model.value = value
  hide(true)
}

const move = (step: number) => {
  const n = shown.value.length
  if (!n) return
  active.value = (active.value + step + n) % n
  scrollToActive()
}

const onKeydown = (e: KeyboardEvent) => {
  if (!open.value) {
    if (['Enter', ' ', 'ArrowDown'].includes(e.key)) { e.preventDefault(); show() }
    return
  }
  switch (e.key) {
    case 'Escape':    e.preventDefault(); hide(true); break
    case 'ArrowDown': e.preventDefault(); move(1); break
    case 'ArrowUp':   e.preventDefault(); move(-1); break
    case 'Home':      e.preventDefault(); active.value = 0; scrollToActive(); break
    case 'End':       e.preventDefault(); active.value = shown.value.length - 1; scrollToActive(); break
    case 'Enter':
      e.preventDefault()
      if (shown.value[active.value]) pick(shown.value[active.value].value)
      break
    case 'Tab': hide(); break
  }
}

/** Фильтр мог отрезать подсвеченную строку — возвращаем подсветку в начало */
watch(filter, () => { active.value = 0 })

const onPointerDown = (e: PointerEvent) => {
  if (open.value && !root.value?.contains(e.target as Node)) hide()
}
onMounted(() => document.addEventListener('pointerdown', onPointerDown))
onBeforeUnmount(() => document.removeEventListener('pointerdown', onPointerDown))
</script>

<template>
  <div ref="root" class="region" @keydown="onKeydown">
    <span :id="`${uid}-label`" class="label">Регион</span>

    <button
      type="button" class="trigger"
      :aria-labelledby="`${uid}-label ${uid}-value`"
      :aria-expanded="open" aria-haspopup="listbox"
      @click="open ? hide() : show()"
    >
      <span :id="`${uid}-value`" class="value">
        {{ current.name }}
        <span v-if="current.codes.length" class="codes">{{ current.codes.join(', ') }}</span>
      </span>
      <svg class="caret" width="10" height="6" viewBox="0 0 10 6" aria-hidden="true">
        <path d="M1 1l4 4 4-4" fill="none" stroke="currentColor" stroke-width="1.6" />
      </svg>
    </button>

    <div v-if="open" class="pop">
      <input
        v-if="searchable" ref="search" v-model="filter" class="filter" type="text"
        autocomplete="off" spellcheck="false" placeholder="Название или код"
        aria-label="Поиск региона"
      >

      <div
        ref="list" class="list" role="listbox" tabindex="-1"
        :aria-labelledby="`${uid}-label`"
        :aria-activedescendant="shown[active] ? `${uid}-opt-${active}` : undefined"
      >
        <div
          v-for="(o, i) in shown" :id="`${uid}-opt-${i}`" :key="o.value || 'any'"
          class="opt" role="option" :data-idx="i"
          :aria-selected="o.value === model" :class="{ on: i === active, sel: o.value === model }"
          @click="pick(o.value)" @mousemove="active = i"
        >
          <span class="name">{{ o.name }}</span>
          <span v-if="o.codes.length" class="codes">{{ o.codes.join(', ') }}</span>
        </div>

        <p v-if="!shown.length" class="none">Ничего не нашлось</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.region { position: relative; display: grid; gap: 6px; min-width: 240px; }
.label { font-size: 14px; font-weight: 500; color: var(--text-muted); }

.trigger {
  font: inherit; font-size: 15.5px; min-height: 44px; padding: 10px 12px; cursor: pointer;
  display: flex; align-items: center; justify-content: space-between; gap: 10px; text-align: left;
  background: var(--surface); color: var(--text);
  border: 1px solid var(--border-strong); border-radius: var(--r-md);
}
.trigger:hover { border-color: var(--text-muted); }
.trigger:focus-visible { outline: none; border-color: var(--accent); box-shadow: 0 0 0 3px var(--accent-sunk); }
.value { display: flex; align-items: baseline; gap: 8px; min-width: 0; }
.caret { flex: none; color: var(--text-faint); }

/* Коды — служебная подпись при названии, поэтому узкое начертание и приглушённый цвет */
.codes {
  font-family: var(--font-plate); font-size: 14.5px; color: var(--text-faint);
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}

.pop {
  position: absolute; z-index: 20; top: calc(100% + 4px); left: 0; min-width: 100%;
  background: var(--surface); border: 1px solid var(--border-strong);
  border-radius: var(--r-lg); box-shadow: var(--sh-3); overflow: hidden;
}
.filter {
  font: inherit; font-size: 15.5px; width: 100%; min-height: 44px; padding: 10px 12px;
  background: var(--surface); color: var(--text);
  border: 0; border-bottom: 1px solid var(--border); border-radius: 0;
}
.filter:focus { outline: none; background: var(--surface-sunk); }

.list { max-height: 320px; overflow-y: auto; padding: 4px; }
.list:focus { outline: none; }
.opt {
  display: flex; align-items: baseline; gap: 10px; justify-content: space-between;
  min-height: 44px; padding: 10px 12px; cursor: pointer; border-radius: var(--r-sm);
}
.opt.on { background: var(--surface-sunk); }
.opt.sel .name { font-weight: 600; color: var(--accent); }
.name { white-space: nowrap; }

.none { padding: 14px 12px; color: var(--text-muted); }
</style>
