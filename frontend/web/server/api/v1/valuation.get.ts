import type { Valuation } from '~/types/api'
import { coreFetch } from '~~/server/utils/core'
import { normalize, validate } from '~~/server/utils/plate'

/** Оценка любого номера, включая те, которых нет в базе, — это 99% запросов. */
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

  return coreFetch<Valuation>('/api/v1/valuation', { number })
})
