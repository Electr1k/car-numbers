package plate

import (
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
