package syncactiveoffers

import (
	"data-service/internal/domain"
)

// Params - входные параметры для синхронизации оффера
type Params struct {
	// Providers - провайдеры, для которых производится синхронизация
	Providers []domain.Provider
}

// providers - провайдеры, для которых производится синхронизация
func (p Params) providers() []domain.Provider {
	if len(p.Providers) == 0 {
		return domain.GetAllProviders()
	}

	return p.Providers
}
