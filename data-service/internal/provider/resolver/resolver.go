package resolver

import (
	"data-service/internal/domain"
	"data-service/internal/provider/autonomera"
	"data-service/internal/usecase"
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

func (r *Resolver) Resolve(provider domain.Provider) (usecase.OfferDetailProvider, error) {
	switch provider {
	case domain.ProviderAutonomera:
		return r.autonomeraProvider, nil
	}

	return nil, fmt.Errorf("provider %s is not supported", provider)
}
