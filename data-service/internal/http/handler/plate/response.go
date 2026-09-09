package plate

import (
	"data-service/internal/domain/data"
	"time"

	"github.com/google/uuid"
)

type numberResponse struct {
	Id     uuid.UUID `json:"id"`
	Number string    `json:"number"`
	Region *region   `json:"region"`
	Offers []offer   `json:"offers"`
}

type region struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type offer struct {
	Id              uuid.UUID `json:"id"`
	Provider        string    `json:"provider"`
	Price           *float64  `json:"price"`
	Status          string    `json:"status"`
	ReissueIncluded *bool     `json:"reissue_included"`
	Whereabouts     *string   `json:"whereabouts"`
	ViewCount       *int      `json:"view_count"`
	Comment         *string   `json:"comment"`
	PostedAt        time.Time `json:"posted_at"`
	RefreshedAt     time.Time `json:"refreshed_at"`
	Url             string    `json:"url"`
}

// mapFeedNumbers - маппинг свежих номеров в ответ API
func mapNumber(number data.Number) numberResponse {
	var regionJson *region = nil
	if number.RegionName != nil && number.RegionCode != nil {
		regionJson = &region{
			*number.RegionName,
			*number.RegionCode,
		}
	}

	offersJson := make([]offer, 0, len(number.Offers))
	for _, offerDomain := range number.Offers {
		var whereabouts *string = nil
		if offerDomain.Whereabouts != nil {
			w := string(*offerDomain.Whereabouts)
			whereabouts = &w
		}
		offersJson = append(offersJson, offer{
			Id:              offerDomain.Id,
			Provider:        string(offerDomain.Provider),
			Price:           offerDomain.Price,
			Status:          string(offerDomain.Status),
			ReissueIncluded: offerDomain.ReissueIncluded,
			Whereabouts:     whereabouts,
			ViewCount:       offerDomain.ViewCount,
			Comment:         offerDomain.Comment,
			PostedAt:        *offerDomain.PostedAt,
			RefreshedAt:     *offerDomain.RefreshedAt,
			Url:             offerDomain.Url,
		})
	}

	return numberResponse{
		Id:     number.Id,
		Number: number.Number,
		Region: regionJson,
		Offers: offersJson,
	}
}
