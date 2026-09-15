package plate

import (
	"core-service/internal/usecase/fetchplate"
	"core-service/internal/usecase/fetchplates"
	"time"

	"github.com/google/uuid"
)

type fetchPlatesResponse struct {
	Items      []fetchPlatesItem `json:"items"`
	NextCursor *string           `json:"next_cursor"`
}

type fetchPlatesItem struct {
	ID              uuid.UUID          `json:"id"`
	Number          string             `json:"number"`
	Region          *fetchPlatesRegion `json:"region"`
	Price           *float64           `json:"price"`
	Type            string             `json:"type"`
	Count           int                `json:"count"`
	RefreshedAt     time.Time          `json:"refreshed_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	ReissueIncluded *bool              `json:"reissue_included"`
}

type fetchPlatesRegion struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func mapFetchPlatesResponse(result fetchplates.Result) fetchPlatesResponse {
	items := make([]fetchPlatesItem, 0, len(result.Items))
	for _, item := range result.Items {
		var region *fetchPlatesRegion
		if item.Region != nil {
			region = &fetchPlatesRegion{
				Code: item.Region.Code,
				Name: item.Region.Name,
			}
		}
		items = append(items, fetchPlatesItem{
			ID:              item.ID,
			Number:          item.Number,
			Region:          region,
			Price:           item.Price,
			Type:            item.Type,
			Count:           item.Count,
			RefreshedAt:     item.RefreshedAt,
			UpdatedAt:       item.UpdatedAt,
			ReissueIncluded: item.ReissueIncluded,
		})
	}

	return fetchPlatesResponse{
		Items:      items,
		NextCursor: result.NextCursor,
	}
}

type fetchPlateByIDResponse struct {
	ID            uuid.UUID             `json:"id"`
	Number        string                `json:"number"`
	Region        *fetchPlateByIDRegion `json:"region"`
	ActiveOffers  []fetchPlateByIDOffer `json:"active_offers"`
	ArchiveOffers []fetchPlateByIDOffer `json:"archive_offers"`
}

type fetchPlateByIDOffer struct {
	ID              uuid.UUID `json:"id"`
	Provider        string    `json:"provider"`
	Price           *float64  `json:"price"`
	Status          string    `json:"status"`
	ReissueIncluded *bool     `json:"reissue_included"`
	Whereabouts     *string   `json:"whereabouts"`
	ViewCount       *int      `json:"view_count"`
	Comment         *string   `json:"comment"`
	RefreshedAt     time.Time `json:"refreshed_at"`
	PostedAt        time.Time `json:"posted_at"`
	URL             string    `json:"url"`
}

type fetchPlateByIDRegion struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func mapFetchPlateByIDResponse(result fetchplate.Result) fetchPlateByIDResponse {
	var region *fetchPlateByIDRegion
	if result.Region != nil {
		region = &fetchPlateByIDRegion{
			Code: result.Region.Code,
			Name: result.Region.Name,
		}
	}

	mapOffer := func(offer fetchplate.Offer) fetchPlateByIDOffer {
		return fetchPlateByIDOffer{
			ID:              offer.ID,
			Provider:        offer.Provider,
			Price:           offer.Price,
			Status:          offer.Status,
			ReissueIncluded: offer.ReissueIncluded,
			Whereabouts:     offer.Whereabouts,
			ViewCount:       offer.ViewCount,
			Comment:         offer.Comment,
			RefreshedAt:     offer.RefreshedAt,
			PostedAt:        offer.PostedAt,
			URL:             offer.URL,
		}
	}

	activeOffers := make([]fetchPlateByIDOffer, 0, len(result.ActiveOffers))
	for _, offer := range result.ActiveOffers {
		activeOffers = append(activeOffers, mapOffer(offer))
	}

	archiveOffers := make([]fetchPlateByIDOffer, 0, len(result.ArchiveOffers))
	for _, offer := range result.ArchiveOffers {
		archiveOffers = append(archiveOffers, mapOffer(offer))
	}

	return fetchPlateByIDResponse{
		ID:            result.ID,
		Number:        result.Number,
		Region:        region,
		ActiveOffers:  activeOffers,
		ArchiveOffers: archiveOffers,
	}
}
