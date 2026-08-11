package usecase

import (
	"context"
	"core/config"
	"core/internal/domain"
	"core/internal/providers/autonomera"
	"core/internal/repository/postgres"
	"fmt"
	"log/slog"
	"time"
)

type FetchAutonomeraUseCase struct {
	provider         autonomera.Service
	config           config.AutoNomeraConfig
	numberRepository postgres.NumberRepository
	offerRepository  postgres.OfferRepository
	logger           *slog.Logger
}

func NewParseAutonomeraUseCase(
	provider autonomera.Service,
	config config.AutoNomeraConfig,
	numberRepository postgres.NumberRepository,
	offerRepository postgres.OfferRepository,
	logger *slog.Logger,
) *FetchAutonomeraUseCase {
	return &FetchAutonomeraUseCase{
		provider:         provider,
		config:           config,
		numberRepository: numberRepository,
		offerRepository:  offerRepository,
		logger:           logger,
	}
}

func (uc *FetchAutonomeraUseCase) Handle(ctx context.Context) error {

	currentDate := time.Now()
	stopDate := time.Now().Add(-uc.config.StopAfter)
	count := 0
	uc.logger.Info("Starting fetching numbers from autonomera777", "stop_date", stopDate)

	for currentDate.After(stopDate) {
		batch, err := uc.provider.FetchOffers(ctx, count)
		if err != nil {
			uc.logger.Error("Failed fetching numbers", "error", err)
			return fmt.Errorf("failed fetching numbers: %w", err)
		}

		var lastOffer *domain.Offer
		var lastVehicleNumber *domain.Number
		for _, offer := range batch {
			lastVehicleNumber, err = uc.numberRepository.UpdateOrCreate(ctx, offer.Number)
			if err != nil {
				uc.logger.Error("Failed creating number", "error", err)
				return fmt.Errorf("failed creating number: %w", err)
			}
			offer.Offer.NumberId = lastVehicleNumber.Id

			lastOffer, err = uc.offerRepository.UpdateOrCreate(ctx, offer.Offer)
			if err != nil {
				uc.logger.Error("Failed creating offer", "error", err)
				return fmt.Errorf("failed creating offer: %w", err)
			}
		}
		count += 20

		if lastOffer != nil {
			currentDate = *lastOffer.PostedAt
		}

		ln := ""
		if lastVehicleNumber != nil {
			ln = lastVehicleNumber.Number
		}
		uc.logger.Info("Last offer", "last date", currentDate, "last number", ln)
	}

	uc.logger.Info("Success fetching numbers from autonomera777", "count", count)

	return nil
}
