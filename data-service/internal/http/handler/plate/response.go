package plate

import (
	"data-service/internal/domain/data"
	"time"

	"github.com/google/uuid"
)

type numberResponse struct {
	ID     uuid.UUID `json:"id"`
	Number string    `json:"number"`
	Region *region   `json:"region"`
	Offers []offer   `json:"offers"`
}

type region struct {
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

// mapNumber - маппинг номера в ответ API
func mapNumber(number data.Number) numberResponse {
	var regionJSON *region
	if number.RegionName != nil && number.RegionCode != nil {
		regionJSON = &region{
			Name: *number.RegionName,
			Code: *number.RegionCode,
		}
	}

	offersJSON := make([]offer, 0, len(number.Offers))
	for _, offerDomain := range number.Offers {
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

	return numberResponse{
		ID:     number.ID,
		Number: number.Number,
		Region: regionJSON,
		Offers: offersJSON,
	}
}
