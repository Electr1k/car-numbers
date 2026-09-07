/** Разбор номера на сервере. Правила — в shared/plate.ts, общем с клиентом. */
import { CANONICAL, LETTERS, MASKED, MAX_MASKS, MOTO, normalizePlate } from '@shared/plate'

export { MAX_MASKS, CANONICAL }

export const normalize = normalizePlate

export interface PlateProblem { code: string; message: string }

/** null — номер годится; иначе готовый к показу отказ */
export function validate (number: string): PlateProblem | null {
  if (MOTO.test(number)) {
    return { code: 'moto', message: `Номер ${number} — мото или прицеп, такие мы не оцениваем.` }
  }

  if (!MASKED.test(number)) {
    return { code: 'format', message: 'Номер вводится как А123ВС777.' }
  }

  const masks = (number.match(/\*/g) || []).length
  if (masks > MAX_MASKS) {
    return {
      code: 'too_many_masks',
      message: `В номере ${masks} неизвестных знаков — при таком разбросе оценка перестаёт нести смысл.`
    }
  }

  return null
}

/* --- Маска для выдачи --- */

/**
 * Маска со звёздочками: `А*7*ВС77` совпадает с `А171ВС77`.
 * Строка короче номера считается началом — человек в поиске часто недопечатывает.
 */
export function matchesMask (number: string, mask: string): boolean {
  const m = normalize(mask)
  if (!m) return true
  if (m.length > number.length) return false

  for (let i = 0; i < m.length; i++) {
    if (m[i] !== '*' && m[i] !== number[i]) return false
  }
  return true
}
