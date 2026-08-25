package resolver

import (
	"data-service/internal/domain"
	"data-service/internal/provider"
	"data-service/internal/provider/autonomera"
	"fmt"
)

type Resolver struct {
	autonomeraProvider *autonomera.Service
}

func NewResolver(autonomeraProvider *autonomera.Service) *Resolver {
	return &Resolver{
		autonomeraProvider: autonomeraProvider,
	}
}

// Resolve - провайдер деталки по имени поставщика оффера
func (r *Resolver) Resolve(name domain.Provider) (provider.OfferDetailProvider, error) {
	switch name {
	case domain.ProviderAutonomera:
		return r.autonomeraProvider, nil
	}

	return nil, fmt.Errorf("provider %s is not supported", name)
}
