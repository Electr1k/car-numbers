package consumer

import (
	"context"
	"log/slog"
	"plate-service/config"
	"plate-service/internal/domain"
	"plate-service/internal/job"
	"plate-service/internal/provider/resolver"
	"plate-service/internal/usecase/importanomera"
	"plate-service/internal/usecase/importautonomera"
	"plate-service/internal/usecase/importgosnomeru"
	"plate-service/internal/usecase/importofferdetail"
	"plate-service/internal/usecase/syncactiveoffers"

	"github.com/google/uuid"
)

type OfferRepository interface {
	SaveBatch(ctx context.Context, items []domain.OfferWithPlate) ([]domain.OfferWithPlate, error)
	GetOfferByID(ctx context.Context, id uuid.UUID) (domain.OfferWithPlate, error)
	GetOfferIdsByProviderAndStatus(ctx context.Context, provider domain.Provider, status domain.OfferStatus) ([]uuid.UUID, error)
	UpdateOffer(ctx context.Context, offer *domain.Offer) error
}

type Features interface {
	Enabled(ctx context.Context, key domain.FeatureKey) (bool, error)
}

// Deps - зависимости консьюмеров джоб
type Deps struct {
	Producer   *job.Producer
	Providers  *resolver.Resolver
	Offers     OfferRepository
	Features   Features
	AutoNomera config.AutoNomeraConfig
	Logger     *slog.Logger
}

// Register - регистрация всех консьюмеров джоб в резолвере
func Register(r *Resolver, d Deps) {
	r.Register(
		domain.JobNameImportAutonomeraOffers,
		NewImportAutonomeraOffersConsumer(importautonomera.New(
			d.Providers.Autonomera(),
			d.Producer,
			d.Offers,
			d.Features,
			d.AutoNomera,
			d.Logger.With("provider", domain.ProviderAutonomera),
		)),
	)

	r.Register(
		domain.JobNameImportGosnomeruOffers,
		NewImportGosnomeruOffersConsumer(importgosnomeru.New(
			d.Providers.Gosnomeru(),
			d.Producer,
			d.Offers,
			d.Features,
			d.Logger.With("provider", domain.ProviderGosnomeru),
		)),
	)

	r.Register(
		domain.JobNameImportAnomeraOffers,
		NewImportAnomeraOffersConsumer(importanomera.New(
			d.Providers.Anomera(),
			d.Producer,
			d.Offers,
			d.Features,
			d.Logger.With("provider", domain.ProviderAnomera),
		)),
	)

	r.Register(
		domain.JobNameImportOfferDetail,
		NewImportOfferDetailConsumer(importofferdetail.New(
			d.Providers,
			d.Offers,
			d.Features,
			d.Logger,
		)),
	)

	r.Register(
		domain.JobNameSyncActiveOffers,
		NewSyncActiveOffersConsumer(syncactiveoffers.New(
			d.Offers,
			d.Producer,
			d.Features,
			d.Logger,
		)),
	)
}
