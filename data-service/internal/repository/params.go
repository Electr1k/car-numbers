package repository

import "data-service/internal/domain/data"

type GetPlatesParams struct {
	Query           *string
	RegionId        *int
	PriceFrom       *float64
	PriceTo         *float64
	ReissueIncluded *bool
	CategoryIds     []int
	Limit           int
	Cursor          *data.FeedCursor
}
