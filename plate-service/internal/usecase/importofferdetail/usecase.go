package importofferdetail

import (
	"context"
	"log/slog"
	"plate-service/internal/domain"
	"plate-service/internal/provider"

	"github.com/google/uuid"
)

// resolver - выбирает провайдера деталки по имени поставщика оффера
type resolver interface {
	Resolve(name domain.Provider) (provider.OfferDetailProvider, error)
}

type offerRepository interface {
	GetOfferByID(ctx context.Context, id uuid.UUID) (domain.OfferWithPlate, error)

	UpdateOffer(ctx context.Context, offer *domain.Offer) error
}

type profileRepository interface {
	GetExistingProfileExternalIDs(ctx context.Context, provider domain.Provider, externalIDs []string) ([]string, error)
}

type profileDispatcher interface {
	DispatchImportProfile(ctx context.Context, provider domain.Provider, profileExternalId string) (bool, error)
}

type feature interface {
	Enabled(ctx context.Context, key domain.FeatureKey) (bool, error)
}

// UseCase - догрузка деталки по офферу
type UseCase struct {
	resolver          resolver
	offerRepository   offerRepository
	profileRepository profileRepository
	profileDispatcher profileDispatcher
	features          feature
	logger            *slog.Logger
}

func New(
	resolver resolver,
	offerRepository offerRepository,
	profileRepository profileRepository,
	profileDispatcher profileDispatcher,
	features feature,
	logger *slog.Logger,
) *UseCase {
	return &UseCase{
		resolver:          resolver,
		offerRepository:   offerRepository,
		profileRepository: profileRepository,
		profileDispatcher: profileDispatcher,
		features:          features,
		logger:            logger,
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

	logger.InfoContext(ctx, "import detail started")

	offer, err := uc.offerRepository.GetOfferByID(ctx, id)
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

	err = uc.offerRepository.UpdateOffer(ctx, enriched.Offer)
	if err != nil {
		return err
	}

	dispatched := false
	if enriched.Offer.ProfileExternalId != nil {
		ids, err := uc.profileRepository.GetExistingProfileExternalIDs(ctx, offer.Offer.Provider, []string{*enriched.Offer.ProfileExternalId})
		if err != nil {
			return err
		}

		if len(ids) == 0 {
			dispatched, err = uc.profileDispatcher.DispatchImportProfile(ctx, offer.Offer.Provider, *enriched.Offer.ProfileExternalId)
			if err != nil {
				return err
			}
		}
	}

	logger.InfoContext(ctx, "import detail finished", "profile_dispatched", dispatched)

	return nil
}
