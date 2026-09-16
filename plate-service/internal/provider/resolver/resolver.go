package resolver

import (
	"fmt"
	"log/slog"
	"plate-service/config"
	"plate-service/internal/domain"
	"plate-service/internal/provider"
	"plate-service/internal/provider/anomera"
	"plate-service/internal/provider/autonomera"
	"plate-service/internal/provider/gosnomeru"
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

// New - сборка резолвера со всеми провайдерами из конфига
func New(cfg *config.Config, logger *slog.Logger) *Resolver {
	return NewResolver(
		autonomera.NewService(
			autonomera.NewClient(cfg.AutoNomeraConfig.BaseURL, logger),
			autonomera.NewMapper(cfg.AutoNomeraConfig.BaseURL),
		),
		gosnomeru.NewService(
			gosnomeru.NewClient(cfg.GosnomeruConfig.BaseURL, logger),
			gosnomeru.NewMapper(cfg.GosnomeruConfig.BaseURL),
		),
		anomera.NewService(
			anomera.NewClient(cfg.AnomeraConfig.BaseURL, logger),
			anomera.NewMapper(cfg.AnomeraConfig.BaseURL),
		),
	)
}

// Autonomera - провайдер autonomera
func (r *Resolver) Autonomera() *autonomera.Service {
	return r.autonomeraProvider
}

// Gosnomeru - провайдер gosnomeru
func (r *Resolver) Gosnomeru() *gosnomeru.Service {
	return r.gosnomeruProvider
}

// Anomera - провайдер anomera
func (r *Resolver) Anomera() *anomera.Service {
	return r.anomeraProvider
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
