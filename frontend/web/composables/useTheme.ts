/** Тема хранится у зрителя; на сервере всегда светлая, чтобы разметка совпала. */
export const useTheme = () => {
  const isDark = useState('theme-dark', () => false)

  const apply = (dark: boolean) => {
    if (import.meta.client) {
      document.documentElement.dataset.theme = dark ? 'dark' : 'light'
      try { localStorage.setItem('theme', dark ? 'dark' : 'light') } catch {}
    }
  }

  onMounted(() => {
    let stored: string | null = null
    try { stored = localStorage.getItem('theme') } catch {}
    const dark = stored
      ? stored === 'dark'
      : window.matchMedia('(prefers-color-scheme: dark)').matches
    isDark.value = dark
    apply(dark)
  })

  const toggle = () => { isDark.value = !isDark.value; apply(isDark.value) }

  return { isDark, toggle }
}
