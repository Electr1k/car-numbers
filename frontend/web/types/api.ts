/** Контракт — frontend/api/README.md */

export type Provider = 'gosnomeru' | 'autonomera' | 'anomera'
export type OfferStatus = 'active' | 'archive' | 'sold'
export type Whereabouts = 'on-car' | 'on-storage' | null
export type Confidence = 'high' | 'medium' | 'low'

/** true — включено, false — нет, null — площадка не сказала */
export type Reissue = boolean | null

export interface RegionRef { code: string; name: string }

export interface NumberCard {
  id: string
  number: string
  region: RegionRef
  price: number
  type: string
  reissue_included: Reissue
  offers_count: number
  updated_at: string
}

export interface Offer {
  id: string
  provider: Provider
  price: number
  status: OfferStatus
  reissue_included: Reissue
  whereabouts: Whereabouts
  views: number | null
  comment: string | null
  posted_at: string
  refreshed_at: string
  url: string
}

export interface BreakdownItem {
  code: string
  value: string
  title: string
  multiplier: number
  exact: boolean
}

/** Оценка живёт отдельным эндпоинтом: /api/v1/valuation?number=X */
export interface Valuation {
  number: string
  region: RegionRef
  p25: number
  p50: number
  p75: number
  confidence: Confidence
  breakdown: { base: number; items: BreakdownItem[] } | null
  mask?: { variants: number; cheapest: number; dearest: number } | null
  basis: { model_version: string; trained_on: string }
}

export interface Plate {
  id: string
  number: string
  region: RegionRef
  active_offers: Offer[]
  archive_offers: Offer[] | null
  similar: NumberCard[]
}

export interface FeedResponse { items: NumberCard[]; cursor: string | null }

export interface SearchResponse {
  query: Record<string, unknown>
  total: number
  items: NumberCard[]
  cursor: string | null
}

export interface RegionGroup { name: string; codes: number[]; active_count?: number }
export interface RegionsResponse { items: RegionGroup[] }

export interface ApiError { error: { code: string; message: string } }
