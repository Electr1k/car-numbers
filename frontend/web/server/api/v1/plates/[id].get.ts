import type { Plate } from '~/types/api'
import { readMock } from '~~/server/utils/mocks'

/**
 * Страница номера адресуется идентификатором: он приходит в карточке ленты и выдачи.
 * Номер, которого нет в базе, идентификатора не имеет — такой запрос идёт в /valuation.
 */
const MOCKS = ['plate.json', 'plate-archive.json']

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id') || ''

  for (const file of MOCKS) {
    const plate = await readMock<Plate>(file)
    if (plate.id === id) return plate
  }

  // В моках всего два номера; остальные идентификаторы отдаём как первый,
  // чтобы страницу можно было открыть из любой карточки.
  return { ...(await readMock<Plate>('plate.json')), id }
})
