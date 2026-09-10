package feed

import (
	"data-service/internal/domain"
	"data-service/internal/domain/data"
	"time"

	"github.com/google/uuid"
)

type feedNumberResponse struct {
	Items      []feedNumberItem `json:"items"`
	NextCursor *string          `json:"next_cursor"`
}

type feedNumberItem struct {
	ID              uuid.UUID         `json:"id"`
	Number          string            `json:"number"`
	Region          *feedRegion       `json:"region"`
	Price           *float64          `json:"price"`
	Type            domain.NumberType `json:"type"`
	Count           int               `json:"count"`
	RefreshedAt     time.Time         `json:"refreshed_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	ReissueIncluded *bool             `json:"reissue_included"`
}

type feedRegion struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// mapFeedNumbers - маппинг свежих номеров в ответ API
func mapFeedNumbers(numbers []data.FeedNumber, nextCursor *string) feedNumberResponse {
	items := make([]feedNumberItem, 0, len(numbers))
	for _, number := range numbers {
		var region *feedRegion
		if number.RegionName != nil && number.RegionCode != nil {
			region = &feedRegion{
				Name: *number.RegionName,
				Code: *number.RegionCode,
			}
		}

		items = append(items, feedNumberItem{
			ID:              number.ID,
			Number:          number.Number,
			Region:          region,
			Price:           number.Price,
			Type:            number.Type,
			Count:           number.Count,
			RefreshedAt:     number.RefreshedAt,
			UpdatedAt:       number.UpdatedAt,
			ReissueIncluded: number.ReissueIncluded,
		})
	}

	return feedNumberResponse{Items: items, NextCursor: nextCursor}
}
