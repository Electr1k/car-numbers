/** Разбор номера. Один источник правды для клиента и сервера. */

export const LETTERS = 'АВЕКМНОРСТУХ'
export const MASK_CHAR = '*'

/** Латиница-двойник → кириллица: раскладка у человека любая, P должно стать Р */
const LOOKALIKE: Record<string, string> = {
  A: 'А', B: 'В', E: 'Е', K: 'К', M: 'М', H: 'Н',
  O: 'О', P: 'Р', C: 'С', T: 'Т', Y: 'У', X: 'Х'
}

const ALLOWED = new Set([...LETTERS, ...'0123456789', MASK_CHAR])

export const CANONICAL = new RegExp(`^[${LETTERS}][0-9]{3}[${LETTERS}]{2}[0-9]{2,3}$`)
export const MASKED = new RegExp(`^[${LETTERS}*][0-9*]{3}[${LETTERS}*]{2}[0-9*]{2,3}$`)
export const MOTO = new RegExp(`^[0-9]{4}[${LETTERS}]{2}[0-9]{2,3}$`)
export const MAX_MASKS = 2
export const MAX_LENGTH = 9

/** Приводит к верхнему регистру и переводит латиницу-двойник */
export function normalizePlate (raw: string): string {
  return raw.toUpperCase().replace(/\s+/g, '')
    .replace(/[ABEKMHOPCTYX]/g, ch => LOOKALIKE[ch] ?? ch)
}

/**
 * То же плюс выброс недопустимых символов и ограничение длины.
 * Для поля ввода: человек физически не может напечатать «Ж» или «@».
 */
export function sanitizePlateInput (raw: string): string {
  return [...normalizePlate(raw)].filter(ch => ALLOWED.has(ch)).slice(0, MAX_LENGTH).join('')
}
