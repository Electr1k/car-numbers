/** Контракт — ответы core-service, GET /api/v1/* */

export type Provider = 'gosnomeru' | 'autonomera' | 'anomera'
export type OfferStatus = 'active' | 'sold' | 'inactive'
export type Whereabouts = 'on-car' | 'on-storage' | null
export type Confidence = 'high' | 'medium' | 'low'

/** true — включено, false — нет, null — площадка не сказала */
export type Reissue = boolean | null

export interface RegionRef { id: number; code: string; name: string }

export interface PlateItem {
  id: string
  number: string
  region: RegionRef | null
  price: number | null
  type: string
  count: number
  refreshed_at: string
  updated_at: string
  reissue_included: Reissue
}

export interface PlatesResponse {
  items: PlateItem[]
  next_cursor: string | null
}

export interface SellerProfile {
  id: string
  login: string
  name: string | null
  icon_url: string | null
}

export interface Offer {
  id: string
  provider: Provider
  price: number | null
  status: OfferStatus
  reissue_included: Reissue
  whereabouts: Whereabouts
  view_count: number | null
  comment: string | null
  posted_at: string
  refreshed_at: string
  url: string
  profile: SellerProfile | null
}

/** Признак, по которому другой аккаунт считается тем же продавцом */
export type ProfileMatch = 'phone' | 'email' | 'login'

export interface ProfileOffers { active: number; total: number }

export interface RelatedProfile {
  id: string
  provider: Provider
  url: string
  icon_url: string | null
  login: string
  name: string | null
  offers: ProfileOffers
  matches: ProfileMatch[]
}

export interface ProfileDetail {
  id: string
  provider: Provider
  url: string
  icon_url: string | null
  login: string
  name: string | null
  /** 11 цифр вида 79991234567 */
  phone: string | null
  email: string | null
  offers: ProfileOffers
  related: RelatedProfile[]
}

export interface ProfileOffer {
  id: string
  plate_id: string
  number: string
  price: number | null
  status: OfferStatus
  posted_at: string
  refreshed_at: string
  url: string
}

export interface ProfileOffersResponse { items: ProfileOffer[]; next_cursor: string | null }

export interface PlateDetail {
  id: string
  number: string
  region: RegionRef | null
  active_offers: Offer[]
  archive_offers: Offer[]
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
  region: RegionRef | null
  price: { p25: number; p50: number; p75: number }
  confidence: Confidence
  breakdown: { base: number; items: BreakdownItem[] }
}

export interface RegionItem { id: number; name: string; codes: string[] }
export interface RegionsResponse { items: RegionItem[] }

export interface ApiError { error: { code: string; message: string; request_id?: string } }
