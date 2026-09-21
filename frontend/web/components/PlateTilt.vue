<script setup lang="ts">
/**
 * Крупный знак для страницы номера: наклоняется за курсором, блик и надписи смещаются с разной глубиной.
 * Только мышь: на тач-экранах и при «меньше движения» знак стоит ровно.
 */
defineProps<{ number: string }>()

const stage = ref<HTMLElement | null>(null)
const tilt = reactive({ x: 0, y: 0, active: false })
let frame = 0

const onMove = (e: PointerEvent) => {
  if (e.pointerType !== 'mouse' || !stage.value) return
  const r = stage.value.getBoundingClientRect()
  const x = (e.clientX - r.left) / r.width * 2 - 1
  const y = (e.clientY - r.top) / r.height * 2 - 1
  cancelAnimationFrame(frame)
  frame = requestAnimationFrame(() => {
    tilt.x = Math.max(-1, Math.min(1, x))
    tilt.y = Math.max(-1, Math.min(1, y))
    tilt.active = true
  })
}

const onLeave = () => {
  cancelAnimationFrame(frame)
  Object.assign(tilt, { x: 0, y: 0, active: false })
}

onBeforeUnmount(() => cancelAnimationFrame(frame))

const vars = computed(() => ({ '--tx': tilt.x.toFixed(3), '--ty': tilt.y.toFixed(3) }))
</script>

<template>
  <div ref="stage" class="stage" :class="{ active: tilt.active }" :style="vars" @pointermove="onMove" @pointerleave="onLeave">
    <div class="tilt">
      <PlateSign :number="number" large />
      <span class="glare" aria-hidden="true" />
    </div>
  </div>
</template>

<style scoped>
/* Поля вокруг знака — зона, где курсор уже наклоняет его */
.stage { perspective: 900px; padding: 18px; margin: -18px; max-width: calc(100% + 36px); }

.tilt {
  position: relative; display: inline-block; max-width: 100%;
  transform: rotateX(calc(var(--ty) * -12deg)) rotateY(calc(var(--tx) * 12deg));
  transition: transform .6s cubic-bezier(.2, .8, .2, 1), filter .6s ease;
  will-change: transform;
}
.active .tilt {
  transition: transform .15s ease-out, filter .15s ease-out;
  filter: drop-shadow(calc(var(--tx) * -10px) calc(12px - var(--ty) * 6px) 14px rgba(0, 0, 0, .22));
}

.tilt :deep(.plate) { box-shadow: var(--sh-2); }

/* Надписи «ближе» к зрителю, чем пластина: сдвигаются сильнее её наклона */
.tilt :deep(.main),
.tilt :deep(.reg) {
  transform: translate(calc(var(--tx) * 4px), calc(var(--ty) * 3px));
  transition: transform .6s cubic-bezier(.2, .8, .2, 1);
}
.active .tilt :deep(.main),
.active .tilt :deep(.reg) { transition-duration: .15s; }

/* Пластина белая, осветлять нечего: «блик» — это тень, отступающая от места, куда светит лампа */
.glare {
  position: absolute; inset: 0; border-radius: 9px; pointer-events: none;
  background: radial-gradient(
    circle at calc(50% - var(--tx) * 45%) calc(50% - var(--ty) * 60%),
    rgba(0, 0, 0, 0) 20%, rgba(0, 0, 0, .14) 85%
  );
  mix-blend-mode: multiply; opacity: 0; transition: opacity .4s ease;
}
.active .glare { opacity: 1; }

@media (prefers-reduced-motion: reduce) {
  .tilt, .tilt :deep(.main), .tilt :deep(.reg) { transform: none; }
  .active .tilt { filter: none; }
  .glare { display: none; }
}
</style>
