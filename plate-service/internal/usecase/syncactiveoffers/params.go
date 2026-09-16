package syncactiveoffers

import (
	"fmt"
	"plate-service/internal/domain"
	"slices"
)

// Params - входные параметры для синхронизации активных офферов
type Params struct {
	// Provider - провайдер, для которого производится синхронизация
	Provider domain.Provider
}

func (p Params) validate() error {
	if !slices.Contains(domain.GetAllProviders(), p.Provider) {
		return fmt.Errorf("unknown provider %q", p.Provider)
	}

	return nil
}
