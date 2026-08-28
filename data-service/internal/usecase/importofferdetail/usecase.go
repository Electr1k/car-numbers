package importofferdetail

import (
	"context"
	"data-service/internal/domain"
	"data-service/internal/provider"
	"log/slog"

	"github.com/google/uuid"
)

// resolver - выбирает провайдера деталки по имени поставщика оффера
type resolver interface {
	Resolve(name domain.Provider) (provider.OfferDetailProvider, error)
}

type offerRepository interface {
	GetOfferById(ctx context.Context, id uuid.UUID) (domain.OfferWithNumber, error)

	UpdateOffer(ctx context.Context, offer *domain.Offer) error
}

type feature interface {
	Enabled(ctx context.Context, key domain.FeatureKey) (bool, error)
}

// UseCase - догрузка деталки по офферу
type UseCase struct {
	resolver   resolver
	repository offerRepository
	features   feature
	logger     *slog.Logger
}

func New(
	resolver resolver,
	repository offerRepository,
	features feature,
	logger *slog.Logger,
) *UseCase {
	return &UseCase{
		resolver:   resolver,
		repository: repository,
		features:   features,
		logger:     logger,
	}
}

// Handle - догружает деталку оффера и сохраняет её в базу
func (uc *UseCase) Handle(ctx context.Context, id uuid.UUID) error {
	logger := uc.logger.With("offer_id", id)

	enabled, err := uc.features.Enabled(ctx, domain.FeatureKeyImportOfferDetail)
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}

	logger.Info("import detail started")

	offer, err := uc.repository.GetOfferById(ctx, id)
	if err != nil {
		return err
	}

	detailProvider, err := uc.resolver.Resolve(offer.Offer.Provider)
	if err != nil {
		return err
	}

	enriched, err := detailProvider.FetchOfferDetail(ctx, offer)
	if err != nil {
		return err
	}

	err = uc.repository.UpdateOffer(ctx, enriched.Offer)
	if err != nil {
		return err
	}
	logger.Info("import detail finished")

	return nil
}
