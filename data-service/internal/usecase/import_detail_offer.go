package usecase

import (
	"context"
	"data-service/internal/domain"
	"log/slog"

	"github.com/google/uuid"
)

type OfferDetailProvider interface {
	FetchOfferDetail(ctx context.Context, offer *domain.OfferWithNumber) (*domain.OfferWithNumber, error)
}

// Resolver - выбирает провайдера деталки по имени провайдера оффера
type Resolver interface {
	Resolve(provider domain.Provider) (OfferDetailProvider, error)
}

type OfferRepository interface {
	GetOfferById(ctx context.Context, id uuid.UUID) (domain.OfferWithNumber, error)

	UpdateOffer(ctx context.Context, offer *domain.Offer) error
}

// ImportOfferDetailUseCase - догрузка деталки по офферу
type ImportOfferDetailUseCase struct {
	resolver   Resolver
	repository OfferRepository
	logger     *slog.Logger
}

func NewImportOfferDetailUseCase(
	resolver Resolver,
	repository OfferRepository,
	logger *slog.Logger,
) *ImportOfferDetailUseCase {
	return &ImportOfferDetailUseCase{
		resolver:   resolver,
		repository: repository,
		logger:     logger,
	}
}

// Handle - собирает предложения поставщика и сохраняет их в базу
func (uc *ImportOfferDetailUseCase) Handle(ctx context.Context, id uuid.UUID) error {
	logger := uc.logger.With("offer_id", id)
	logger.Info("import detail started")

	offer, err := uc.repository.GetOfferById(ctx, id)
	if err != nil {
		return err
	}

	detailProvider, err := uc.resolver.Resolve(offer.Offer.Provider)
	if err != nil {
		return err
	}

	_, err = detailProvider.FetchOfferDetail(ctx, &offer)
	if err != nil {
		return err
	}

	err = uc.repository.UpdateOffer(ctx, offer.Offer)
	if err != nil {
		return err
	}
	logger.Info("import detail finished")

	return nil
}
