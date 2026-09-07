/** Узоры номера. Классы те же, что у модели оценки: ML/predictor/features.py */

const digits = (n: string) => n.slice(1, 4)
const series = (n: string) => n[0] + n.slice(4, 6)
const region = (n: string) => n.slice(6)

export const PATTERNS = {
  /* цифры */
  same_digits:      (n: string) => new Set(digits(n)).size === 1,
  pair_digits:      (n: string) => new Set(digits(n)).size === 2,
  round_hundred:    (n: string) => /^[1-9]00$/.test(digits(n)),
  first_ten:        (n: string) => /^00[1-9]$/.test(digits(n)),
  mirror:           (n: string) => digits(n)[0] === digits(n)[2] && digits(n)[0] !== digits(n)[1],
  zero_edges:       (n: string) => digits(n)[0] === '0' && digits(n)[2] === '0',
  ladder:           (n: string) => {
    const d = digits(n).split('').map(Number)
    return (d[1] === d[0] + 1 && d[2] === d[1] + 1) || (d[1] === d[0] - 1 && d[2] === d[1] - 1)
  },
  /* буквы */
  same_letters:     (n: string) => new Set(series(n)).size === 1,
  pair_letters:     (n: string) => new Set(series(n)).size === 2,
  /* пересечение — самый сильный единичный признак модели, ×4,18 */
  digits_as_region: (n: string) => Number(digits(n)) === Number(region(n))
} as const

export type PatternCode = keyof typeof PATTERNS

/** Подписи — из словаря рынка: так же они называются у площадок */
export const PATTERN_LIST: { code: PatternCode; label: string; group: 'digits' | 'letters' }[] = [
  { code: 'same_digits',      label: 'Одинаковые цифры', group: 'digits' },
  { code: 'pair_digits',      label: 'Пара цифр',        group: 'digits' },
  { code: 'round_hundred',    label: 'Ровная сотня',     group: 'digits' },
  { code: 'first_ten',        label: 'Первая десятка',   group: 'digits' },
  { code: 'mirror',           label: 'Зеркальные цифры', group: 'digits' },
  { code: 'zero_edges',       label: 'Нули по краям',    group: 'digits' },
  { code: 'ladder',           label: 'Лесенка',          group: 'digits' },
  { code: 'same_letters',     label: 'Одинаковые буквы', group: 'letters' },
  { code: 'pair_letters',     label: 'Пара букв',        group: 'letters' },
  { code: 'digits_as_region', label: 'Цифры как регион', group: 'letters' }
]
