import type { ApiError } from '~/types/api'

/** Запрос в core-service; ошибку апстрима пробрасываем со статусом и телом как есть. */
export async function coreFetch<T> (path: string, query?: Record<string, unknown>): Promise<T> {
  const baseURL = useRuntimeConfig().coreUrl

  try {
    return await $fetch<T>(path, { baseURL, query })
  } catch (e) {
    const err = e as { status?: number; statusCode?: number; data?: ApiError }
    const body = err.data?.error

    throw createError({
      statusCode: err.status ?? err.statusCode ?? 502,
      statusMessage: body?.code ?? 'upstream_error',
      data: { error: body ?? { code: 'upstream_error', message: 'Сервис временно недоступен.' } }
    })
  }
}
