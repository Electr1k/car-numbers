package importgosnomerubydetail

import (
	"context"
	"data-service/internal/domain"
	"data-service/internal/provider"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
)

type offerProvider interface {
	FetchLatestOffers(ctx context.Context) (provider.FetchResult, error)
	FetchOfferDetail(ctx context.Context, offer domain.OfferWithNumber) (domain.OfferWithNumber, error)
	FetchOfferDetailByExternalId(ctx context.Context, externalId string) (domain.OfferWithNumber, error)
}

type offerRepository interface {
	GetOfferByExternalId(ctx context.Context, provider domain.Provider, externalId string) (domain.OfferWithNumber, error)
	UpdateOffer(ctx context.Context, offer *domain.Offer) error
	UpdateOrCreate(ctx context.Context, offer domain.OfferWithNumber) (domain.OfferWithNumber, error)
}

type feature interface {
	Enabled(ctx context.Context, key domain.FeatureKey) (bool, error)
}

// UseCase - импорт деталок по последовательности id gosnomeru
type UseCase struct {
	provider   offerProvider
	repository offerRepository
	features   feature
	logger     *slog.Logger
}

func New(
	provider offerProvider,
	repository offerRepository,
	features feature,
	logger *slog.Logger,
) *UseCase {
	return &UseCase{
		provider:   provider,
		repository: repository,
		features:   features,
		logger:     logger,
	}
}

// Handle - собирает предложения поставщика и сохраняет их в базу через перебор id деталки
func (uc *UseCase) Handle(ctx context.Context, params Params) error {
	err := params.validate()
	if err != nil {
		return err
	}

	enabled, err := uc.features.Enabled(ctx, domain.FeatureKeyImportGosnomeruOffers)
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}

	startId := params.StartId
	endId := params.EndId

	if endId == 0 {
		var err error
		endId, err = uc.fetchLastExternalId(ctx)
		if err != nil {
			return err
		}
	}

	uc.logger.Info("start import offers by enumeration id", "startId", startId, "endId", endId)

	for ; startId <= endId; startId++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		externalId := strconv.Itoa(startId)

		// Проверка существования оффера в БД
		stored, err := uc.repository.GetOfferByExternalId(ctx, domain.ProviderGosnomeru, externalId)
		if err != nil && !errors.Is(err, domain.ErrOfferNotFound) {
			return fmt.Errorf("get offer by external id %s: %w", externalId, err)
		}

		if errors.Is(err, domain.ErrOfferNotFound) {
			err = uc.importNew(ctx, externalId)
		} else {
			err = uc.refreshStored(ctx, stored)
		}

		if err != nil {
			msg := "failed to process offer"
			if errors.Is(err, provider.ErrNotFound) || errors.Is(err, provider.ErrRowSkipped) {
				uc.logger.Debug(msg, "err", err, "external_id", externalId)
			} else {
				uc.logger.Error(msg, "err", err, "external_id", externalId)
			}
		}
	}
	uc.logger.Info("end import offers by enumeration id", "endId", endId)

	return nil
}

// fetchLastExternalId - возвращает последний Id провайдера
func (uc *UseCase) fetchLastExternalId(ctx context.Context) (int, error) {
	latestOffers, err := uc.provider.FetchLatestOffers(ctx)
	if err != nil {
		return 0, fmt.Errorf("fetch latest offers: %w", err)
	}

	if len(latestOffers.Offers) == 0 {
		return 0, fmt.Errorf("last offer not found")
	}

	endId, err := strconv.Atoi(latestOffers.Offers[0].Offer.ExternalId)
	if err != nil {
		return 0, fmt.Errorf("invalid external offer id: %w", err)
	}

	return endId, nil
}

// refreshStored - догружает деталку поверх уже сохранённого предложения, не трогая поля из выдачи
func (uc *UseCase) refreshStored(ctx context.Context, stored domain.OfferWithNumber) error {
	offer, err := uc.provider.FetchOfferDetail(ctx, stored)
	if err != nil {
		return fmt.Errorf("fetch offer detail: %w", err)
	}

	if err := uc.repository.UpdateOffer(ctx, offer.Offer); err != nil {
		return fmt.Errorf("update offer: %w", err)
	}

	return nil
}

// importNew - сохраняет предложение, которого ещё нет в базе, целиком из деталки
func (uc *UseCase) importNew(ctx context.Context, externalId string) error {
	offer, err := uc.provider.FetchOfferDetailByExternalId(ctx, externalId)
	if err != nil {
		return err
	}

	if _, err := uc.repository.UpdateOrCreate(ctx, offer); err != nil {
		return fmt.Errorf("update or create offer: %w", err)
	}

	return nil
}
