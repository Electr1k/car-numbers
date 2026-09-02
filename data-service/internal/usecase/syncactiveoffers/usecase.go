package syncactiveoffers

import (
	"context"
	"data-service/internal/domain"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

type offerStore interface {
	GetOffers(ctx context.Context, status domain.OfferStatus, provider domain.Provider) ([]domain.OfferWithNumber, error)
}

type detailDispatcher interface {
	DispatchImportOfferDetail(ctx context.Context, offerId uuid.UUID) (bool, error)
}

type feature interface {
	Enabled(ctx context.Context, key domain.FeatureKey) (bool, error)
}

// UseCase - синхронизация статуса активных офферов
type UseCase struct {
	repository       offerStore
	detailDispatcher detailDispatcher
	features         feature
	logger           *slog.Logger
}

func New(
	repository offerStore,
	detailDispatcher detailDispatcher,
	features feature,
	logger *slog.Logger,
) *UseCase {
	return &UseCase{
		repository:       repository,
		detailDispatcher: detailDispatcher,
		features:         features,
		logger:           logger,
	}
}

// Handle - синхронизирует активные офферы (часть провайдеров неактивные офферы удаляет)
func (uc *UseCase) Handle(ctx context.Context, params Params) error {

	enabled, err := uc.features.Enabled(ctx, domain.FeatureKeySyncActiveOffers)
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}

	providers := params.providers()

	uc.logger.Info("sync started", "providers", providers)

	dispatched := 0
	for _, provider := range providers {
		dispatchedOnProvider, err := uc.syncByProvider(ctx, provider)
		if err != nil {
			return err
		}
		dispatched += dispatchedOnProvider
	}

	uc.logger.Info("import finished", "dispatched", dispatched)

	return nil
}

func (uc *UseCase) syncByProvider(ctx context.Context, provider domain.Provider) (int, error) {
	dispatched := 0

	offers, err := uc.repository.GetOffers(ctx, domain.OfferStatusActive, provider)
	if err != nil {
		return dispatched, err
	}

	for _, offer := range offers {
		success, err := uc.detailDispatcher.DispatchImportOfferDetail(ctx, offer.Offer.Id)
		if err != nil {
			return dispatched, fmt.Errorf("dispatch offer details for offer id %s: %w", offer.Offer.Id, err)
		}
		if success {
			dispatched++
		}
	}

	return dispatched, nil
}
