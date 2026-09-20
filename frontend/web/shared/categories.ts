/**
 * Категории номера. Коды и названия — те же, что в таблице categories plate-service:
 * фронт отправляет их в category_ids как есть. До появления справочника в API список держим здесь.
 */

export const CATEGORY_LIST = [
  /* цифры */
  { id: 'same-digits',      label: 'Одинаковые цифры', group: 'digits' },
  { id: 'pair-digits',      label: 'Пара цифр',        group: 'digits' },
  { id: 'round-hundreds',   label: 'Круглые сотни',    group: 'digits' },
  { id: 'first-ten',        label: 'Первая десятка',   group: 'digits' },
  { id: 'mirrored-digits',  label: 'Зеркальные цифры', group: 'digits' },
  { id: 'zero-edges',       label: 'Ноль по краям',    group: 'digits' },
  { id: 'digits-staircase', label: 'Лесенка цифр',     group: 'digits' },
  /* буквы */
  { id: 'same-letters',     label: 'Одинаковые буквы', group: 'letters' },
  { id: 'pair-letters',     label: 'Пара букв',        group: 'letters' },
  { id: 'mirrored-letters', label: 'Зеркальные буквы', group: 'letters' },
  { id: 'letter-staircase', label: 'Лесенка букв',     group: 'letters' },
  /* пересечение цифр и региона */
  { id: 'digits-as-region', label: 'Цифры как регион', group: 'letters' }
] as const

export type CategoryId = typeof CATEGORY_LIST[number]['id']

const KNOWN = new Set<string>(CATEGORY_LIST.map(c => c.id))

/** Отбор известных кодов: в адресе может оказаться что угодно, а бек на чужой код ответит 400 */
export const knownCategories = (codes: string[]): CategoryId[] =>
  codes.filter((code): code is CategoryId => KNOWN.has(code))

export const categoryLabel = (id: string): string | undefined =>
  CATEGORY_LIST.find(c => c.id === id)?.label
