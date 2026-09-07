import { fileURLToPath } from 'node:url'

export default defineNuxtConfig({
  compatibilityDate: '2026-09-07',

  // Явный алиас: встроенный #shared склеивает путь с двойным слешем и теряет расширение
  alias: {
    '@shared': fileURLToPath(new URL('./shared', import.meta.url))
  },
  devtools: { enabled: false },
  ssr: true,

  css: ['~/assets/css/tokens.css', '~/assets/css/base.css'],

  app: {
    head: {
      htmlAttrs: { lang: 'ru' },
      title: 'Номерограф — цены на автономера',
      meta: [
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: 'Объявления о продаже автономеров с трёх площадок и оценка стоимости любого номера.' }
      ],
      link: [
        { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
        { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' },
        {
          rel: 'stylesheet',
          href: 'https://fonts.googleapis.com/css2?family=Golos+Text:wght@400;500;600;700&family=PT+Sans+Narrow:wght@400;700&display=swap'
        }
      ]
    }
  },

  // Каталог моков монтируется в контейнер; путь можно переопределить переменной
  runtimeConfig: {
    mocksDir: process.env.MOCKS_DIR || '../api/mock'
  },

  nitro: { compressPublicAssets: true }
})
