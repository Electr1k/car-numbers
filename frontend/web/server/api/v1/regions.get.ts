import type { RegionsResponse } from '~/types/api'
import { coreFetch } from '~~/server/utils/core'

/** Человек выбирает регион, а не код: у Москвы их семь. Порядок — алфавитный. */
export default defineEventHandler(async () => {
  const data = await coreFetch<RegionsResponse>('/api/v1/regions')
  return { items: [...data.items].sort((a, b) => a.name.localeCompare(b.name, 'ru')) }
})
