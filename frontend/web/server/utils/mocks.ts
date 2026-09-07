import { readFile } from 'node:fs/promises'
import { resolve } from 'node:path'

/**
 * Пока BFF не написан, эндпоинты отдают моки из frontend/api/mocks.
 * Форма ответов боевая — при переходе на настоящий сервис менять её не придётся.
 */
export async function readMock<T> (name: string): Promise<T> {
  const dir = useRuntimeConfig().mocksDir
  const raw = await readFile(resolve(process.cwd(), dir, name), 'utf-8')
  return JSON.parse(raw) as T
}
