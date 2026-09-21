<script setup lang="ts">
/** Крупный знак для страницы номера: наклоняется за курсором целиком, надписи и блик — внутри PlateSign */
defineProps<{ number: string }>()

const { el, style, active, onMove, onLeave } = useTilt(() => 12)
</script>

<template>
  <div ref="el" class="stage" :class="{ active }" :style="style" @pointermove="onMove" @pointerleave="onLeave">
    <div class="tilt">
      <PlateSign :number="number" large />
    </div>
  </div>
</template>

<style scoped>
/* Поля вокруг знака — зона, где курсор уже наклоняет его */
.stage { perspective: 900px; padding: 18px; margin: -18px; max-width: calc(100% + 36px); }

.tilt {
  display: inline-block; max-width: 100%;
  transform: rotateX(calc(var(--ty) * var(--deg) * -1)) rotateY(calc(var(--tx) * var(--deg)));
  transition: transform var(--tilt-speed) cubic-bezier(.2, .8, .2, 1), filter var(--tilt-speed) ease;
}
.active .tilt {
  will-change: transform;
  filter: drop-shadow(calc(var(--tx) * -10px) calc(12px - var(--ty) * 6px) 14px rgba(0, 0, 0, .22));
}
.tilt :deep(.plate) { box-shadow: var(--sh-2); }
</style>
