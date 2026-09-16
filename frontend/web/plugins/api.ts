/** Запросы в core-service: на сервере — по внутреннему адресу, в браузере — через прокси /api того же хоста */
export default defineNuxtPlugin(() => {
  const api = $fetch.create({
    baseURL: import.meta.server ? useRuntimeConfig().coreUrl : '/'
  })

  return { provide: { api } }
})
