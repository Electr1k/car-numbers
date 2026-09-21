/** Запросы в core-service: на сервере — по внутреннему адресу, в браузере — через прокси /api того же хоста */
export default defineNuxtPlugin(() => {
  const api = $fetch.create({
    baseURL: import.meta.server ? useRuntimeConfig().coreUrl : '/',
    headers: import.meta.server ? clientIPHeaders() : undefined
  })

  return { provide: { api } }
})

/** IP посетителя для рейт лимита core: от nginx, иначе адрес соединения */
function clientIPHeaders (): Record<string, string> {
  const ip = useRequestHeaders(['x-real-ip'])['x-real-ip'] || useRequestEvent()?.node.req.socket.remoteAddress

  return ip ? { 'X-Real-IP': ip } : {}
}
