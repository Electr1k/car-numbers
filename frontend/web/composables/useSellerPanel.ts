/** Панель продавца одна на приложение: открывается из карточки, закрывается из панели */
export const useSellerPanel = () => {
  const profileId = useState<string | null>('seller-panel', () => null)

  return {
    profileId,
    open: (id: string) => { profileId.value = id },
    close: () => { profileId.value = null }
  }
}
