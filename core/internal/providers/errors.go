package providers

import "errors"

var (
	ErrProviderUnavailable = errors.New("provider unavailable")
	ErrInvalidResponse     = errors.New("invalid provider response")
	ErrRateLimitExceeded   = errors.New("rate limit exceeded")
	ErrNoMoreData          = errors.New("no more data available")
	ErrParsingFailed       = errors.New("failed to parse provider data")
)

// Ошибки получения оффера
//   - ErrSchemaDrift - сломался парсер: разметка провайдера изменилась
//   - ErrRowSkipped - не проходит валидация домена
var (
	ErrSchemaDrift = errors.New("schema drift: provider markup changed")
	ErrRowSkipped  = errors.New("row skipped: does not satisfy domain rules")
)
