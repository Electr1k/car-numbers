package repository

import (
	"plate-service/internal/domain"
	"plate-service/internal/domain/data"
)

type GetPlatesParams struct {
	Query           *string
	RegionId        *int
	PriceFrom       *float64
	PriceTo         *float64
	ReissueIncluded *bool
	CategoryIds     []domain.CategoryID
	Sort            data.PlateSort
	Limit           int
	Cursor          *data.FeedCursor
}
