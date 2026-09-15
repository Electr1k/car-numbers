package fetchplate

import (
	"context"
	"core-service/internal/service/plate"

	"github.com/google/uuid"
)

const activeOffer = "active"

type plateProvider interface {
	FetchPlateByID(ctx context.Context, id uuid.UUID) (*plate.FetchPlateByIDResponse, error)
}

// UseCase - возвращает номер с предложениями
type UseCase struct {
	plateClient plateProvider
}

func New(
	plateClient plateProvider,
) *UseCase {
	return &UseCase{
		plateClient: plateClient,
	}
}

// Handle - возвращает деталку номера
func (uc *UseCase) Handle(ctx context.Context, id uuid.UUID) (*Result, error) {
	plateResult, err := uc.plateClient.FetchPlateByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var region *Region
	if plateResult.Region != nil {
		region = &Region{
			Code: plateResult.Region.Code,
			Name: plateResult.Region.Name,
		}
	}

	activeOffers := make([]Offer, 0, len(plateResult.Offers))
	archiveOffers := make([]Offer, 0, len(plateResult.Offers))
	for _, o := range plateResult.Offers {
		offer := Offer{
			ID:              o.ID,
			Provider:        o.Provider,
			Price:           o.Price,
			Status:          o.Status,
			ReissueIncluded: o.ReissueIncluded,
			Whereabouts:     o.Whereabouts,
			ViewCount:       o.ViewCount,
			Comment:         o.Comment,
			RefreshedAt:     o.RefreshedAt,
			PostedAt:        o.PostedAt,
			URL:             o.URL,
		}
		if o.Status == activeOffer {
			activeOffers = append(activeOffers, offer)
		} else {
			archiveOffers = append(archiveOffers, offer)
		}
	}

	return &Result{
		ID:            plateResult.ID,
		Number:        plateResult.Number,
		Region:        region,
		ActiveOffers:  activeOffers,
		ArchiveOffers: archiveOffers,
	}, nil
}
