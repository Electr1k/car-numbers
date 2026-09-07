import type { RegionsResponse } from '~/types/api'
import { readMock } from '~~/server/utils/mocks'

/** Человек выбирает регион, а не код: у Москвы их семь. Порядок — алфавитный. */
export default defineEventHandler(async () => {
  const data = await readMock<RegionsResponse>('regions.json')
  return { items: [...data.items].sort((a, b) => a.name.localeCompare(b.name, 'ru')) }
})
