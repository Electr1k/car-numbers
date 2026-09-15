import type { PlatesResponse, RegionsResponse } from '~/types/api'
import { coreFetch } from '~~/server/utils/core'
import { normalize } from '~~/server/utils/plate'

/** core-service фильтрует по идентификатору региона, а в адресе страницы живут коды */
const resolveRegionId = async (codes: string[]): Promise<number | undefined> => {
  if (!codes.length) return undefined

  const { items } = await coreFetch<RegionsResponse>('/api/v1/regions')
  return items.find(r => r.codes.some(c => codes.includes(c)))?.id
}

const parseReissue = (raw: string): boolean | undefined => {
  if (raw === 'true') return true
  if (raw === 'false') return false
  return undefined
}

export default defineEventHandler(async (event) => {
  const p = getQuery(event)
  const codes = String(p.region || '').split(',').filter(Boolean)

  return coreFetch<PlatesResponse>('/api/v1/plates', {
    query: p.q ? normalize(String(p.q)) : undefined,
    region_id: await resolveRegionId(codes),
    price_from: Number(p.price_min) || undefined,
    price_to: Number(p.price_max) || undefined,
    reissue_included: parseReissue(String(p.reissue_included ?? '')),
    limit: Math.min(Number(p.limit) || 24, 48),
    cursor: String(p.cursor || '') || undefined
  })
})
