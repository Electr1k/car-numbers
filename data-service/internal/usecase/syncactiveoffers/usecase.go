package syncactiveoffers

import (
	"context"
	"data-service/internal/domain"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

type offerStore interface {
	GetOfferIdsByProviderAndStatus(ctx context.Context, provider domain.Provider, status domain.OfferStatus) ([]uuid.UUID, error)
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

// Handle - синхронизирует активные офферы провайдера (часть провайдеров неактивные офферы удаляет)
func (uc *UseCase) Handle(ctx context.Context, params Params) error {
	if err := params.validate(); err != nil {
		return err
	}

	enabled, err := uc.features.Enabled(ctx, domain.FeatureKeySyncActiveOffers)
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}

	logger := uc.logger.With("provider", params.Provider)
	logger.Info("sync started")

	offerIds, err := uc.repository.GetOfferIdsByProviderAndStatus(ctx, params.Provider, domain.OfferStatusActive)
	if err != nil {
		return err
	}

	dispatched := 0
	for _, offerId := range offerIds {
		success, err := uc.detailDispatcher.DispatchImportOfferDetail(ctx, offerId)
		if err != nil {
			return fmt.Errorf("dispatch offer details for offer id %s: %w", offerId, err)
		}
		if success {
			dispatched++
		}
	}

	logger.Info("sync finished", "offers", len(offerIds), "dispatched", dispatched)

	return nil
}
