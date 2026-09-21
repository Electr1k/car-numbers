/**
 * Наклон за курсором: отдаёт CSS-переменные, из которых элемент и его содержимое строят параллакс.
 * --tx/--ty — положение курсора от -1 до 1, --deg — предел наклона, --tilt-on — 1, пока курсор над элементом.
 * Только мышь и только без «меньше движения»: иначе переменные остаются нулевыми.
 */
export function useTilt (maxDeg: (width: number) => number) {
  const el = ref<HTMLElement | null>(null)
  const state = reactive({ x: 0, y: 0, deg: 0, active: false })
  let frame = 0

  const onMove = (e: PointerEvent) => {
    if (e.pointerType !== 'mouse' || !el.value) return
    if (matchMedia('(prefers-reduced-motion: reduce)').matches) return
    const r = el.value.getBoundingClientRect()
    const x = (e.clientX - r.left) / r.width * 2 - 1
    const y = (e.clientY - r.top) / r.height * 2 - 1
    cancelAnimationFrame(frame)
    frame = requestAnimationFrame(() => {
      state.x = Math.max(-1, Math.min(1, x))
      state.y = Math.max(-1, Math.min(1, y))
      state.deg = maxDeg(r.width)
      state.active = true
    })
  }

  const onLeave = () => {
    cancelAnimationFrame(frame)
    Object.assign(state, { x: 0, y: 0, active: false })
  }

  onBeforeUnmount(() => cancelAnimationFrame(frame))

  const style = computed(() => ({
    '--tx': state.x.toFixed(3),
    '--ty': state.y.toFixed(3),
    '--deg': `${state.deg.toFixed(2)}deg`,
    '--tilt-on': state.active ? 1 : 0,
    // Пока курсор ведёт — отклик мгновенный, отпустил — плавный возврат
    '--tilt-speed': state.active ? '.15s' : '.6s'
  }))

  return { el, style, active: computed(() => state.active), onMove, onLeave }
}
