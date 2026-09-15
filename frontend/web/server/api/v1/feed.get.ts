import type { PlatesResponse } from '~/types/api'
import { coreFetch } from '~~/server/utils/core'

export default defineEventHandler(async (event) => {
  const { limit, cursor } = getQuery(event)

  return coreFetch<PlatesResponse>('/api/v1/plates', {
    limit: Number(limit) || undefined,
    cursor: String(cursor || '') || undefined
  })
})
