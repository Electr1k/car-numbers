package provider

import "errors"

var (
	// Ошибки запроса

	// ErrProviderUnavailable - не смогли достучаться до поставщика
	ErrProviderUnavailable = errors.New("provider unavailable")
	// ErrInvalidResponse - ошибка получения респонса
	ErrInvalidResponse = errors.New("invalid provider response")
	// ErrRateLimitExceeded - уперлись в рейт лимиты
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// Ошибки маппинга

	// ErrBrokenOffer - ошибка получения оффера (вероятно изменилась верстка)
	ErrBrokenOffer = errors.New("failed to parse offer")
	// ErrRowSkipped - ошибка валидация в домене (не всегда крит) - собрать смогли, но не поддерживает бизнес правила - скипаем
	ErrRowSkipped = errors.New("row skipped: does not satisfy domain rules")
)
