<script setup lang="ts">
/**
 * Настоящий знак: в списках обычный размер, на странице номера — крупный.
 * Знак не инвертируется в тёмной теме — это физический предмет.
 */
const props = defineProps<{ number: string; large?: boolean }>()

const parts = computed(() => {
  const m = props.number.match(/^(.*?)(\d{2,3})$/)
  return m ? { main: m[1], region: m[2] } : { main: props.number, region: '' }
})
</script>

<template>
  <span class="plate" :class="{ large }" :aria-label="`Номер ${number}`">
    <span class="main">{{ parts.main }}</span>
    <span v-if="parts.region" class="reg">
      <span class="n">{{ parts.region }}</span>
      <span class="rus">
        <svg viewBox="0 0 12 8" width="12" height="8" aria-hidden="true" class="flag">
          <rect width="12" height="2.67" fill="#fff" />
          <rect y="2.67" width="12" height="2.67" fill="#0039a6" />
          <rect y="5.34" width="12" height="2.66" fill="#d52b1e" />
        </svg>RUS
      </span>
    </span>
  </span>
</template>

<style scoped>
.plate {
  display: inline-flex; align-items: stretch;
  background: var(--plate-face); color: var(--plate-ink);
  border: 2px solid var(--plate-edge); border-radius: 6px; overflow: hidden;
  font-family: var(--font-plate); font-weight: 700; font-variant-numeric: tabular-nums;
  box-shadow: var(--sh-1);
}
.main { padding: 8px 16px 6px; font-size: 44px; letter-spacing: .06em; line-height: 1; }
.reg {
  border-left: 2px solid var(--plate-edge); padding: 6px 11px 5px;
  display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 2px;
}
.n { font-size: 29px; line-height: 1; }
.rus { font-size: 9px; letter-spacing: .06em; line-height: 1; display: flex; align-items: center; gap: 3px; }
.flag { display: block; border: .5px solid #9a9a9a; }

.large { border-width: 3px; border-radius: 9px; }
.large .main { padding: 12px 22px 9px; font-size: 68px; }
.large .reg { border-left-width: 3px; padding: 9px 15px 7px; gap: 4px; }
.large .n { font-size: 44px; }
.large .rus { font-size: 13px; gap: 4px; }
.large .flag { width: 17px; height: 11px; }

@media (max-width: 620px) {
  .main { font-size: 34px; padding: 6px 12px 5px; }
  .n { font-size: 22px; }
  .large .main { font-size: 50px; padding: 9px 16px 7px; }
  .large .n { font-size: 33px; }
  .large .rus { font-size: 11px; }
}
</style>
