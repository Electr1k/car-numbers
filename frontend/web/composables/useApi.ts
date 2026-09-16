import type { UseFetchOptions } from 'nuxt/app'

/** useFetch поверх $api: те же опции, но запрос уходит в core-service */
export function useApi<T> (url: string | (() => string), options: UseFetchOptions<T> = {}) {
  return useFetch(url, { ...options, $fetch: useNuxtApp().$api as typeof $fetch })
}
