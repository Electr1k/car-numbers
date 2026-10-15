<script setup lang="ts">
import type { OfferStatus, ProfileDetail, ProfileMatch, ProfileOffer, ProfileOffersResponse } from '~/types/api'
import { PROVIDER_SITE, formatPhone, sellerTitle } from '@shared/providers'
import { OFFERS, plural } from '@shared/plural'

/*
 * Профиль грузится только в браузере по клику: контакты продавца не попадают
 * в серверную разметку и в поисковую выдачу.
 */
const { profileId, close } = useSellerPanel()
const { $api } = useNuxtApp()
const route = useRoute()

/* Переходы по связанным аккаунтам складываются в цепочку, «Назад» возвращает к предыдущему */
const trail = ref<string[]>([])
const current = computed(() => trail.value.at(-1) ?? null)

const profile = ref<ProfileDetail | null>(null)
const status = ref<'loading' | 'error' | 'ready'>('loading')
const errorText = ref('')

/* Уже открытые профили не запрашиваем повторно: у core лимит запросов в минуту */
const cache = new Map<string, ProfileDetail>()
let requestSeq = 0

const load = async (id: string) => {
  const seq = ++requestSeq
  const cached = cache.get(id)
  if (cached) {
    profile.value = cached
    status.value = 'ready'
    return
  }

  status.value = 'loading'
  try {
    const res = await $api<ProfileDetail>(`/api/v1/profiles/${id}`)
    cache.set(id, res)
    if (seq !== requestSeq) return
    profile.value = res
    status.value = 'ready'
  } catch (e: any) {
    if (seq !== requestSeq) return
    const code = e?.statusCode ?? e?.response?.status
    errorText.value = code === 404
      ? 'Профиль не найден.'
      : code === 429
        ? 'Слишком много запросов. Попробуйте через минуту.'
        : 'Не удалось загрузить профиль. Попробуйте ещё раз.'
    status.value = 'error'
  }
}

watch(profileId, (id) => {
  offersMode.value = null
  trail.value = id ? [id] : []
})
watch(current, (id) => {
  if (!id) return
  load(id)
  nextTick(() => { bodyEl.value?.scrollTo({ top: 0 }) })
})

/* Список предложений открывается поверх профиля; «Назад» сначала закрывает его */
type OffersMode = 'active' | 'archive'

/* Какие статусы показывает каждая вкладка: архив — снятые и проданные */
const MODE_STATUSES: Record<OffersMode, OfferStatus[]> = {
  active: ['active'],
  archive: ['sold', 'inactive']
}
const offersMode = ref<OffersMode | null>(null)

const openRelated = (id: string) => {
  offersMode.value = null
  trail.value = [...trail.value, id]
}
const back = () => {
  if (offersMode.value) offersMode.value = null
  else trail.value = trail.value.slice(0, -1)
}
const retry = () => { if (current.value) load(current.value) }
const canGoBack = computed(() => offersMode.value !== null || trail.value.length > 1)

/* Предложения грузятся страницами; уже загруженные храним по профилю и режиму */
const OFFERS_PAGE = 30
const offers = ref<ProfileOffer[]>([])
/* Курсор следующей страницы; null — список загружен целиком */
const offersCursor = ref<string | null>(null)
const offersStatus = ref<'loading' | 'more' | 'error' | 'ready'>('loading')
const offersCache = new Map<string, { items: ProfileOffer[]; cursor: string | null }>()
let offersSeq = 0

const loadOffers = async (append = false) => {
  const id = current.value
  const mode = offersMode.value
  if (!id || !mode) return

  const key = `${id}:${mode}`
  const cached = offersCache.get(key)
  if (!append && cached) {
    offers.value = cached.items
    offersCursor.value = cached.cursor
    offersStatus.value = 'ready'
    return
  }

  const seq = ++offersSeq
  offersStatus.value = append ? 'more' : 'loading'
  try {
    const res = await $api<ProfileOffersResponse>(`/api/v1/profiles/${id}/offers`, {
      query: {
        statuses: MODE_STATUSES[mode],
        limit: OFFERS_PAGE,
        cursor: append ? offersCursor.value ?? undefined : undefined
      }
    })
    if (seq !== offersSeq) return
    const items = append ? [...offers.value, ...res.items] : res.items
    offersCache.set(key, { items, cursor: res.next_cursor })
    offers.value = items
    offersCursor.value = res.next_cursor
    offersStatus.value = 'ready'
  } catch {
    if (seq !== offersSeq) return
    offersStatus.value = 'error'
  }
}

const showOffers = (mode: OffersMode) => { offersMode.value = mode }
watch(offersMode, (mode) => {
  if (!mode) return
  loadOffers()
  nextTick(() => { bodyEl.value?.scrollTo({ top: 0 }) })
})

const STATUS_LABEL: Record<string, string> = { sold: 'Продано', inactive: 'Снято с продажи' }
const money = (v: number) => v.toLocaleString('ru-RU').replace(/ /g, ' ') + ' ₽'
const date = (iso: string) =>
  new Date(iso).toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' })

/* Уход со страницы закрывает панель */
watch(() => route.fullPath, () => close())

/* Фокус уходит в панель и возвращается на кнопку, которая её открыла; страница под панелью не скроллится */
const panelEl = ref<HTMLElement | null>(null)
const bodyEl = ref<HTMLElement | null>(null)
let opener: HTMLElement | null = null

watch(profileId, (id, prev) => {
  if (!import.meta.client) return
  if (id && !prev) {
    opener = document.activeElement as HTMLElement | null
    document.documentElement.style.overflow = 'hidden'
    nextTick(() => panelEl.value?.focus())
  } else if (!id && prev) {
    document.documentElement.style.overflow = ''
    opener?.focus()
    opener = null
  }
})

onBeforeUnmount(() => {
  if (import.meta.client) document.documentElement.style.overflow = ''
})

const FOCUSABLE = 'a[href], button:not([disabled]), [tabindex]:not([tabindex="-1"])'

/* Esc закрывает, Tab не выпускает фокус из панели */
const onKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Escape') {
    close()
    return
  }
  if (e.key !== 'Tab' || !panelEl.value) return

  const items = [...panelEl.value.querySelectorAll<HTMLElement>(FOCUSABLE)]
  if (!items.length) return
  const first = items[0]!
  const last = items.at(-1)!
  if (e.shiftKey && (document.activeElement === first || document.activeElement === panelEl.value)) {
    e.preventDefault()
    last.focus()
  } else if (!e.shiftKey && document.activeElement === last) {
    e.preventDefault()
    first.focus()
  }
}

const seller = computed(() => profile.value ? sellerTitle(profile.value) : null)
const archiveCount = computed(() => profile.value ? profile.value.offers.total - profile.value.offers.active : 0)

const MATCH_LABEL: Record<ProfileMatch, string> = { phone: 'телефон', email: 'email', login: 'логин' }
const matchText = (matches: ProfileMatch[]) => matches.map(m => MATCH_LABEL[m]).join(', ')
</script>

<template>
  <Teleport to="body">
    <Transition name="panel">
      <div v-if="profileId" class="layer" @keydown="onKeydown">
        <div class="scrim" @click="close" />

        <aside ref="panelEl" class="panel" role="dialog" aria-modal="true" aria-labelledby="seller-panel-title" tabindex="-1">
          <header class="bar">
            <button v-if="canGoBack" type="button" class="icon-btn" aria-label="Назад" @click="back">←</button>
            <h2 id="seller-panel-title">{{ offersMode ? 'Предложения продавца' : 'Продавец' }}</h2>
            <button type="button" class="icon-btn" aria-label="Закрыть" @click="close">✕</button>
          </header>

          <div ref="bodyEl" class="body">
            <div v-if="status === 'loading'" class="stack" aria-busy="true">
              <div class="who">
                <span class="sk sk-ava" aria-hidden="true" />
                <span class="sk sk-line" aria-hidden="true" />
              </div>
              <span class="sk sk-block" aria-hidden="true" />
              <span class="sk sk-block" aria-hidden="true" />
            </div>

            <div v-else-if="status === 'error'" class="stack">
              <p class="fail">{{ errorText }}</p>
              <button type="button" class="btn" @click="retry">Повторить</button>
            </div>

            <div v-else-if="profile && seller && offersMode" class="stack">
              <div class="who who-sm">
                <SellerAvatar :src="profile.icon_url" :title="seller.title" />
                <p class="title-sm">{{ seller.title }}</p>
              </div>

              <div class="seg" role="group" aria-label="Какие предложения показать">
                <button type="button" :aria-pressed="offersMode === 'active'" @click="showOffers('active')">
                  В продаже · {{ profile.offers.active }}
                </button>
                <button type="button" :aria-pressed="offersMode === 'archive'" @click="showOffers('archive')">
                  Архив · {{ archiveCount }}
                </button>
              </div>

              <div v-if="offersStatus === 'loading'" class="offers" aria-busy="true">
                <span v-for="i in 4" :key="i" class="sk sk-offer" aria-hidden="true" />
              </div>

              <div v-else-if="offersStatus === 'error' && !offers.length" class="stack">
                <p class="fail">Не удалось загрузить предложения.</p>
                <button type="button" class="btn" @click="loadOffers()">Повторить</button>
              </div>

              <p v-else-if="!offers.length" class="muted">
                {{ offersMode === 'active' ? 'Сейчас у продавца нет предложений в продаже.' : 'В архиве у продавца пусто.' }}
              </p>

              <template v-else>
                <ul class="offers">
                  <li v-for="o in offers" :key="o.id">
                    <NuxtLink :to="`/plates/${o.plate_id}`" class="offer-row">
                      <PlateSign :number="o.number" class="sign" />
                      <span class="offer-main">
                        <span class="offer-price">{{ o.price === null ? 'Цена не указана' : money(o.price) }}</span>
                        <span class="offer-meta">
                          <template v-if="STATUS_LABEL[o.status]">{{ STATUS_LABEL[o.status] }} · </template>{{ date(o.refreshed_at) }}
                        </span>
                      </span>
                      <span class="rel-go" aria-hidden="true">→</span>
                    </NuxtLink>
                  </li>
                </ul>

                <p v-if="offersStatus === 'error'" class="fail">Не удалось загрузить ещё. Попробуйте снова.</p>
                <button v-if="offersCursor" type="button" class="btn" :disabled="offersStatus === 'more'"
                        @click="loadOffers(true)">
                  {{ offersStatus === 'more' ? 'Загружаем…' : 'Показать ещё' }}
                </button>
              </template>
            </div>

            <div v-else-if="profile && seller" class="stack">
              <div class="who">
                <SellerAvatar :src="profile.icon_url" :title="seller.title" :size="56" />
                <div class="names">
                  <p class="title">{{ seller.title }}</p>
                  <p v-if="seller.login" class="login">{{ seller.login }}</p>
                  <a :href="profile.url" target="_blank" rel="noopener noreferrer nofollow" class="site">
                    Профиль на {{ PROVIDER_SITE[profile.provider] }} →
                  </a>
                </div>
              </div>

              <div class="stats">
                <button type="button" class="stat" :disabled="!profile.offers.active" @click="showOffers('active')">
                  <span class="num">{{ profile.offers.active }}</span>
                  <span class="lbl">{{ plural(profile.offers.active, OFFERS) }} в продаже</span>
                </button>
                <button type="button" class="stat" :disabled="!archiveCount" @click="showOffers('archive')">
                  <span class="num">{{ archiveCount }}</span>
                  <span class="lbl">в архиве</span>
                </button>
              </div>

              <dl v-if="profile.phone || profile.email" class="contacts">
                <template v-if="profile.phone">
                  <dt>Телефон</dt>
                  <dd><a :href="`tel:+${profile.phone}`">{{ formatPhone(profile.phone) }}</a></dd>
                </template>
                <template v-if="profile.email">
                  <dt>Email</dt>
                  <dd><a :href="`mailto:${profile.email}`">{{ profile.email }}</a></dd>
                </template>
              </dl>
              <p v-else class="muted">Контакты продавец на площадке не указал.</p>

              <section v-if="profile.related.length" class="related">
                <h3>Возможно, это тот же продавец</h3>
                <p class="muted">Совпадает телефон, email или логин на этой или другой площадке.</p>

                <ul>
                  <li v-for="r in profile.related" :key="r.id">
                    <button type="button" class="rel" @click="openRelated(r.id)">
                      <SellerAvatar :src="r.icon_url" :title="sellerTitle(r).title" />
                      <span class="rel-main">
                        <span class="rel-title">{{ sellerTitle(r).title }}</span>
                        <span class="rel-meta">
                          {{ PROVIDER_SITE[r.provider] }} · {{ r.offers.active }} в продаже · {{ r.offers.total - r.offers.active }} в архиве
                        </span>
                        <span class="rel-match">Совпадает: {{ matchText(r.matches) }}</span>
                      </span>
                      <span class="rel-go" aria-hidden="true">→</span>
                    </button>
                  </li>
                </ul>
              </section>
            </div>
          </div>
        </aside>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.layer { position: fixed; inset: 0; z-index: 100; }
.scrim { position: absolute; inset: 0; background: rgba(17, 22, 29, .38); }

.panel {
  position: absolute; top: 0; right: 0; bottom: 0;
  width: min(560px, 100%); display: flex; flex-direction: column;
  background: var(--surface); border-left: 1px solid var(--border); box-shadow: var(--sh-3);
}
.panel:focus { outline: none; }

.bar {
  display: flex; align-items: center; gap: 8px;
  padding: 12px 12px 12px 20px; border-bottom: 1px solid var(--border);
}
.bar h2 { flex: 1; font-size: 17px; font-weight: 600; }
.icon-btn {
  width: 40px; height: 40px; display: inline-grid; place-items: center;
  background: none; border: 1px solid transparent; border-radius: var(--r-md);
  font-size: 18px; color: var(--text-muted); cursor: pointer;
}
.icon-btn:hover { background: var(--surface-sunk); color: var(--text); }

.body { flex: 1; overflow-y: auto; padding: 20px; overscroll-behavior: contain; }
.stack { display: grid; gap: 20px; }

.who { display: flex; align-items: center; gap: 14px; min-width: 0; }
.names { display: grid; gap: 2px; min-width: 0; }
.title { font-size: 20px; font-weight: 700; letter-spacing: -.01em; overflow-wrap: anywhere; }
.login { font-size: 14.5px; color: var(--text-faint); overflow-wrap: anywhere; }
.site { font-size: 14.5px; margin-top: 4px; }

.stats { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
.stat {
  display: grid; gap: 2px; text-align: left; cursor: pointer;
  background: var(--surface-sunk); border: 1px solid transparent; border-radius: var(--r-md); padding: 12px 14px;
}
.stat:hover:not(:disabled) { border-color: var(--border-strong); }
.stat:disabled { cursor: default; }
.stat:not(:disabled) .lbl::after { content: ' →'; }
.num { font-size: 24px; font-weight: 700; font-variant-numeric: tabular-nums; }
.lbl { font-size: 13.5px; color: var(--text-muted); }

.who-sm { gap: 10px; }
.title-sm { font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.seg { display: inline-flex; justify-self: start; border: 1px solid var(--border-strong); border-radius: var(--r-md); overflow: hidden; }
.seg button { padding: 8px 14px; background: var(--surface); border: 0; cursor: pointer; font-size: 14.5px; }
.seg button + button { border-left: 1px solid var(--border-strong); }
.seg button[aria-pressed="true"] { background: var(--accent); color: #fff; }

.offers { list-style: none; margin: 0; padding: 0; display: grid; gap: 8px; }
.offer-row {
  display: flex; align-items: center; gap: 14px; padding: 10px 12px;
  border: 1px solid var(--border); border-radius: var(--r-md);
  color: var(--text); text-decoration: none;
}
.offer-row:hover { border-color: var(--border-strong); background: var(--surface-sunk); color: var(--text); text-decoration: none; }
.sign { zoom: .5; flex: none; }
.offer-main { flex: 1; display: grid; gap: 1px; min-width: 0; }
.offer-price { font-weight: 700; font-variant-numeric: tabular-nums; }
.offer-meta { font-size: 13.5px; color: var(--text-faint); }
.sk-offer { display: block; height: 58px; border-radius: var(--r-md); }

.contacts { margin: 0; display: grid; grid-template-columns: auto 1fr; gap: 8px 16px; font-size: 15px; }
.contacts dt { color: var(--text-faint); }
.contacts dd { margin: 0; overflow-wrap: anywhere; }

.muted { font-size: 14.5px; color: var(--text-muted); }
.fail { color: var(--alert); }
.btn {
  justify-self: start; min-height: 40px; padding: 8px 16px;
  background: var(--surface); border: 1px solid var(--border-strong); border-radius: var(--r-md);
  font-weight: 600; cursor: pointer;
}

.related { display: grid; gap: 8px; border-top: 1px solid var(--border); padding-top: 18px; }
.related h3 { font-size: 16px; font-weight: 600; }
.related ul { list-style: none; margin: 4px 0 0; padding: 0; display: grid; gap: 8px; }
.rel {
  width: 100%; display: flex; align-items: center; gap: 12px; text-align: left;
  padding: 10px 12px; background: var(--surface); cursor: pointer;
  border: 1px solid var(--border); border-radius: var(--r-md);
}
.rel:hover { border-color: var(--border-strong); background: var(--surface-sunk); }
.rel-main { flex: 1; display: grid; gap: 1px; min-width: 0; }
.rel-title { font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.rel-meta, .rel-match { font-size: 13.5px; color: var(--text-faint); }
.rel-match { color: var(--text-muted); }
.rel-go { color: var(--text-faint); }

.sk-ava { width: 56px; height: 56px; border-radius: 50%; flex: none; }
.sk-line { width: 160px; height: 20px; }
.sk-block { display: block; height: 64px; border-radius: var(--r-md); }

.panel-enter-active, .panel-leave-active { transition: opacity .2s ease; }
.panel-enter-active .panel, .panel-leave-active .panel { transition: transform .22s ease; }
.panel-enter-from, .panel-leave-to { opacity: 0; }
.panel-enter-from .panel, .panel-leave-to .panel { transform: translateX(100%); }
</style>
