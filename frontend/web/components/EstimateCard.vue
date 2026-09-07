<script setup lang="ts">
import type { Valuation } from '~/types/api'

const props = defineProps<{ estimate: Valuation | null }>()

const money = (v: number) => v.toLocaleString('ru-RU').replace(/ /g, ' ') + ' ₽'
const short = (v: number) =>
  v >= 1_000_000 ? (v / 1_000_000).toFixed(1).replace('.', ',') + ' млн'
                 : Math.round(v / 1000) + ' тыс.'

/* Точка p50 внутри полосы: положение отражает, где медиана лежит между p25 и p75 */
const dotAt = computed(() => {
  const e = props.estimate
  if (!e || e.p75 === e.p25) return 50
  return Math.min(92, Math.max(8, ((e.p50 - e.p25) / (e.p75 - e.p25)) * 100))
})

const confidenceNote = computed(() => ({
  high:   null,
  medium: 'Похожих номеров в этом ценовом сегменте мало — оценка приблизительная.',
  low:    'Оценка ориентировочная: в этом сегменте рынок расходится сильнее всего.'
}[props.estimate?.confidence ?? 'high']))
</script>

<template>
  <section v-if="estimate" class="est" aria-labelledby="est-h">
    <div>
      <h2 id="est-h" class="lbl">За похожие номера просят</h2>
      <p class="p50">{{ money(estimate.p50) }}</p>
    </div>

    <div class="track" role="img"
         :aria-label="`Половина похожих номеров стоит от ${money(estimate.p25)} до ${money(estimate.p75)}`">
      <span class="span" />
      <span class="dot" :style="{ left: dotAt + '%' }" />
      <span class="tk lo">{{ short(estimate.p25) }}</span>
      <span class="tk hi">{{ short(estimate.p75) }}</span>
    </div>

    <p v-if="confidenceNote" class="warn">{{ confidenceNote }}</p>

    <div v-if="estimate.mask" class="mask">
      Неизвестные знаки раскрыты перебором: {{ estimate.mask.variants }} вариантов,
      от {{ money(estimate.mask.cheapest) }} до {{ money(estimate.mask.dearest) }}.
    </div>

    <details v-if="estimate.breakdown" class="why">
      <summary>Из чего сложилась цена</summary>
      <p class="base">Средний уровень рынка — {{ money(estimate.breakdown.base) }}</p>
      <ul class="factors">
        <li v-for="item in estimate.breakdown.items" :key="item.code + item.value">
          <span>{{ item.title }}</span>
          <span class="m" :class="item.multiplier >= 1 ? 'up' : 'down'">
            ×{{ item.multiplier.toFixed(2).replace('.', ',') }}
          </span>
        </li>
      </ul>
      <p class="src">
        Модель от {{ estimate.basis.trained_on }}. Это цена запроса, а не цена сделки.
      </p>
    </details>
  </section>

  <section v-else class="est unavailable">
    <p>Оценка сейчас недоступна. Объявления ниже показаны как есть.</p>
  </section>
</template>

<style scoped>
.est {
  background: var(--surface); border: 1px solid var(--border); border-radius: var(--r-lg);
  padding: 22px 24px; box-shadow: var(--sh-2); display: grid; gap: 14px;
}
.unavailable { box-shadow: var(--sh-1); color: var(--text-muted); }
.lbl { font-size: 14px; font-weight: 500; color: var(--text-muted); letter-spacing: 0; }
.p50 { font-size: 40px; font-weight: 700; letter-spacing: -.03em; line-height: 1; font-variant-numeric: tabular-nums; margin-top: 4px; }

.track { position: relative; height: 10px; background: var(--surface-sunk); border-radius: 5px; margin: 22px 0 32px; }
.span { position: absolute; inset: 0 22% 0 8%; background: var(--accent-sunk); border-radius: 5px; }
.dot {
  position: absolute; top: 50%; width: 16px; height: 16px; margin: -8px 0 0 -8px;
  background: var(--accent); border: 3px solid var(--surface); border-radius: 50%;
}
.tk { position: absolute; top: 18px; font-size: 13.5px; color: var(--text-muted); font-variant-numeric: tabular-nums; white-space: nowrap; }
.tk.lo { left: 8%; transform: translateX(-50%); }
.tk.hi { right: 22%; transform: translateX(50%); }

.warn { font-size: 14.5px; color: var(--warn); background: var(--warn-bg); padding: 10px 12px; border-radius: var(--r-sm); }
.mask { font-size: 14.5px; color: var(--text-muted); background: var(--surface-sunk); padding: 10px 12px; border-radius: var(--r-sm); }

.why { border-top: 1px solid var(--border); padding-top: 12px; }
.why summary { font-size: 15px; font-weight: 600; color: var(--accent); cursor: pointer; min-height: 44px; display: flex; align-items: center; }
.base { font-size: 14.5px; color: var(--text-faint); margin: 4px 0 8px; }
.factors { list-style: none; margin: 0; padding: 0; }
.factors li {
  display: flex; justify-content: space-between; gap: 16px; align-items: baseline;
  padding: 10px 0; border-bottom: 1px solid var(--border); font-size: 15px;
}
.factors li:last-child { border-bottom: none; }
.m { font-variant-numeric: tabular-nums; font-weight: 600; white-space: nowrap; }
.up { color: var(--good); }
.down { color: var(--alert); }
.src { font-size: 14px; color: var(--text-faint); margin-top: 12px; }
</style>
