import type { PlateDetail } from '~/types/api'
import { coreFetch } from '~~/server/utils/core'

/**
 * Страница номера адресуется идентификатором: он приходит в карточке ленты и выдачи.
 * Номер, которого нет в базе, идентификатора не имеет — такой запрос идёт в /valuation.
 */
export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id') || ''
  return coreFetch<PlateDetail>(`/api/v1/plates/${encodeURIComponent(id)}`)
})
