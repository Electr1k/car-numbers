import type { FeedResponse } from '~/types/api'
import { readMock } from '~~/server/utils/mocks'

export default defineEventHandler(async (event) => {
  const { limit } = getQuery(event)
  const feed = await readMock<FeedResponse>('feed.json')
  const n = Number(limit) || feed.items.length
  return { ...feed, items: feed.items.slice(0, n) }
})
