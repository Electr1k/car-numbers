/** Форма слова для числа: plural(5, ['предложение', 'предложения', 'предложений']) */
export const plural = (n: number, [one, few, many]: [string, string, string]) => {
  const t = n % 100, o = n % 10
  if (t >= 11 && t <= 14) return many
  if (o === 1) return one
  if (o >= 2 && o <= 4) return few
  return many
}

export const OFFERS: [string, string, string] = ['предложение', 'предложения', 'предложений']
