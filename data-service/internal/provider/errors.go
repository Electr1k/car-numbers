package provider

import "errors"

var (
	// Ошибки запроса

	// ErrProviderUnavailable - не смогли достучаться до поставщика
	ErrProviderUnavailable = errors.New("provider unavailable")
	// ErrInvalidResponse - ошибка получения респонса
	ErrInvalidResponse = errors.New("invalid provider response")
	// ErrRateLimitExceeded - рейт лимит
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	// ErrNotFound - сущеность не найдена
	ErrNotFound = errors.New("not found")

	// Ошибки маппинга

	// ErrBrokenOffer - ошибка получения оффера (вероятно изменилась верстка)
	ErrBrokenOffer = errors.New("failed to parse offer")
	// ErrRowSkipped - ошибка валидация в домене (не всегда крит) - собрать смогли, но не поддерживает бизнес правила - скипаем
	ErrRowSkipped = errors.New("row skipped: does not satisfy domain rules")
	// ErrMapOffer - ошибка маппинга оффера (например изменился номер в оффере)
	ErrMapOffer = errors.New("failed to map offer")
)
