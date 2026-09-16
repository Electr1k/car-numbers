package plate

import (
	"data-service/internal/domain"
	"data-service/internal/domain/data"
	"time"

	"github.com/google/uuid"
)

type platesResponse struct {
	Items      []platesItem `json:"items"`
	NextCursor *string      `json:"next_cursor"`
}

type platesItem struct {
	ID              uuid.UUID        `json:"id"`
	Number          string           `json:"number"`
	Region          *region          `json:"region"`
	Price           *float64         `json:"price"`
	Type            domain.PlateType `json:"type"`
	Count           int              `json:"count"`
	RefreshedAt     time.Time        `json:"refreshed_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	ReissueIncluded *bool            `json:"reissue_included"`
}

// mapPlates - маппинг свежих номеров в ответ API
func mapPlates(plates []data.FeedPlate, nextCursor *string) platesResponse {
	items := make([]platesItem, 0, len(plates))
	for _, plate := range plates {
		var plateRegion *region
		if plate.RegionID != nil && plate.RegionName != nil && plate.RegionCode != nil {
			plateRegion = &region{
				ID:   *plate.RegionID,
				Name: *plate.RegionName,
				Code: *plate.RegionCode,
			}
		}

		items = append(items, platesItem{
			ID:              plate.ID,
			Number:          plate.Number,
			Region:          plateRegion,
			Price:           plate.Price,
			Type:            plate.Type,
			Count:           plate.Count,
			RefreshedAt:     plate.RefreshedAt,
			UpdatedAt:       plate.UpdatedAt,
			ReissueIncluded: plate.ReissueIncluded,
		})
	}

	return platesResponse{Items: items, NextCursor: nextCursor}
}

type plateResponse struct {
	ID     uuid.UUID `json:"id"`
	Number string    `json:"number"`
	Region *region   `json:"region"`
	Offers []offer   `json:"offers"`
}

type region struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

type offer struct {
	ID              uuid.UUID `json:"id"`
	Provider        string    `json:"provider"`
	Price           *float64  `json:"price"`
	Status          string    `json:"status"`
	ReissueIncluded *bool     `json:"reissue_included"`
	Whereabouts     *string   `json:"whereabouts"`
	ViewCount       *int      `json:"view_count"`
	Comment         *string   `json:"comment"`
	PostedAt        time.Time `json:"posted_at"`
	RefreshedAt     time.Time `json:"refreshed_at"`
	URL             string    `json:"url"`
}

// mapPlate - маппинг номера в ответ API
func mapPlate(plate data.Plate) plateResponse {
	var regionJSON *region
	if plate.RegionID != nil && plate.RegionName != nil && plate.RegionCode != nil {
		regionJSON = &region{
			ID:   *plate.RegionID,
			Name: *plate.RegionName,
			Code: *plate.RegionCode,
		}
	}

	offersJSON := make([]offer, 0, len(plate.Offers))
	for _, offerDomain := range plate.Offers {
		var whereabouts *string
		if offerDomain.Whereabouts != nil {
			w := string(*offerDomain.Whereabouts)
			whereabouts = &w
		}
		offersJSON = append(offersJSON, offer{
			ID:              offerDomain.ID,
			Provider:        string(offerDomain.Provider),
			Price:           offerDomain.Price,
			Status:          string(offerDomain.Status),
			ReissueIncluded: offerDomain.ReissueIncluded,
			Whereabouts:     whereabouts,
			ViewCount:       offerDomain.ViewCount,
			Comment:         offerDomain.Comment,
			PostedAt:        offerDomain.PostedAt,
			RefreshedAt:     offerDomain.RefreshedAt,
			URL:             offerDomain.URL,
		})
	}

	return plateResponse{
		ID:     plate.ID,
		Number: plate.Number,
		Region: regionJSON,
		Offers: offersJSON,
	}
}
