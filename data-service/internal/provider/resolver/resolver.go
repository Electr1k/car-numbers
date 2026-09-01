package resolver

import (
	"data-service/internal/domain"
	"data-service/internal/provider"
	"data-service/internal/provider/anomera"
	"data-service/internal/provider/autonomera"
	"data-service/internal/provider/gosnomeru"
	"fmt"
)

type Resolver struct {
	autonomeraProvider *autonomera.Service
	gosnomeruProvider  *gosnomeru.Service
	anomeraProvider    *anomera.Service
}

func NewResolver(
	autonomeraProvider *autonomera.Service,
	gosnomeruProvider *gosnomeru.Service,
	anomeraProvider *anomera.Service,
) *Resolver {
	return &Resolver{
		autonomeraProvider: autonomeraProvider,
		gosnomeruProvider:  gosnomeruProvider,
		anomeraProvider:    anomeraProvider,
	}
}

// Resolve - провайдер деталки по имени поставщика оффера
func (r *Resolver) Resolve(name domain.Provider) (provider.OfferDetailProvider, error) {
	switch name {
	case domain.ProviderAutonomera:
		return r.autonomeraProvider, nil
	case domain.ProviderGosnomeru:
		return r.gosnomeruProvider, nil
	case domain.ProviderAnomera:
		return r.anomeraProvider, nil
	}

	return nil, fmt.Errorf("provider %s is not supported", name)
}
