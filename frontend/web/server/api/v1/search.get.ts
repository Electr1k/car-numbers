import type { NumberCard } from '~/types/api'
import { readMock } from '~~/server/utils/mocks'
import { matchesMask, normalize } from '~~/server/utils/plate'
import { PATTERNS, type PatternCode } from '@shared/patterns'

const SORTS = {
  price_asc:    (a: NumberCard, b: NumberCard) => a.price - b.price,
  price_desc:   (a: NumberCard, b: NumberCard) => b.price - a.price,
  updated_desc: (a: NumberCard, b: NumberCard) => b.updated_at.localeCompare(a.updated_at)
}

/** `true` / `false` / `null` — последнее означает «площадка не сказала» */
const parseReissue = (raw: string): boolean | null | undefined => {
  if (raw === 'true') return true
  if (raw === 'false') return false
  if (raw === 'null') return null
  return undefined
}

export default defineEventHandler(async (event) => {
  const p = getQuery(event)
  const pool = (await readMock<{ items: NumberCard[] }>('search-pool.json')).items

  const q = p.q ? normalize(String(p.q)) : ''
  const regions = String(p.region || '').split(',').filter(Boolean)
  const reissue = parseReissue(String(p.reissue_included ?? ''))
  const categories = String(p.categories || '').split(',').filter(Boolean) as PatternCode[]
  const min = Number(p.price_min) || 0
  const max = Number(p.price_max) || Infinity
  const sort = String(p.sort || 'updated_desc') as keyof typeof SORTS
  const limit = Math.min(Number(p.limit) || 24, 48)

  const items = pool
    .filter(c => !q || matchesMask(c.number, q))
    .filter(c => !regions.length || regions.includes(c.region.code))
    .filter(c => reissue === undefined || c.reissue_included === reissue)
    .filter(c => c.price >= min && c.price <= max)
    .filter(c => categories.every(code => PATTERNS[code]?.(c.number) ?? true))
    .sort(SORTS[sort] ?? SORTS.updated_desc)

  return {
    query: {
      q: q || null,
      region: regions.join(',') || null,
      reissue_included: reissue === undefined ? null : reissue,
      categories,
      price_min: min || null,
      price_max: Number.isFinite(max) ? max : null,
      sort
    },
    total: items.length,
    items: items.slice(0, limit),
    cursor: items.length > limit ? 'cG9vbDoy' : null
  }
})
