package repository

import "plate-service/internal/domain/data"

type GetPlatesParams struct {
	Query           *string
	RegionId        *int
	PriceFrom       *float64
	PriceTo         *float64
	ReissueIncluded *bool
	CategoryIds     []int
	Sort            data.PlateSort
	Limit           int
	Cursor          *data.FeedCursor
}
