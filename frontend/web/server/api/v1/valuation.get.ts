import type { Valuation } from '~/types/api'
import { readMock } from '~~/server/utils/mocks'
import { normalize, validate } from '~~/server/utils/plate'

/**
 * Оценка любого номера, включая те, которых нет в базе, — это 99% запросов.
 * Пока BFF не написан, отдаём мок с подставленным номером и регионом.
 */
export default defineEventHandler(async (event) => {
  const raw = String(getQuery(event).number || '')
  const number = normalize(raw)

  if (!number) {
    throw createError({
      statusCode: 400, statusMessage: 'number_required',
      data: { error: { code: 'number_required', message: 'Укажите номер для оценки.' } }
    })
  }

  const problem = validate(number)
  if (problem) {
    throw createError({ statusCode: 422, statusMessage: problem.code, data: { error: problem } })
  }

  const base = await readMock<Valuation>('valuation.json')
  if (number === base.number) return base

  const code = number.match(/[0-9]{2,3}$/)?.[0] ?? ''
  return { ...base, number, region: { ...base.region, code } }
})
