/** Сайт площадки для подписи рядом с объявлением или профилем */
export const PROVIDER_SITE = {
  gosnomeru: 'gosnomeru.com',
  autonomera: 'autonomera777.ru',
  anomera: 'anomera6.ru'
} as const

/** Имя продавца показываем, только если оно отличается от логина */
export const sellerTitle = (p: { login: string; name: string | null }) => {
  const name = p.name && p.name !== p.login ? p.name : null
  return { title: name ?? p.login, login: name ? p.login : null }
}

/** 79991234567 → +7 999 123-45-67 */
export const formatPhone = (phone: string) =>
  /^7\d{10}$/.test(phone)
    ? `+7 ${phone.slice(1, 4)} ${phone.slice(4, 7)}-${phone.slice(7, 9)}-${phone.slice(9)}`
    : phone
