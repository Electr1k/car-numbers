package fetchplates

import (
	"core-service/internal/domain"
	"core-service/internal/service"
	"core-service/internal/service/plate"
	"fmt"
)

const (
	minLimit = 1
	maxLimit = 25
)

// Params - входные параметры выборки свежих номеров
type Params struct {
	Query           *string
	RegionID        *int
	PriceFrom       *float64
	PriceTo         *float64
	ReissueIncluded *bool
	CategoryIDs     []int
	Sort            domain.PlateSort
	Limit           int
	Cursor          string
}

func (p Params) normalize() Params {
	if p.Query != nil {
		if query := domain.NormalizePlate(*p.Query); query != "" {
			p.Query = &query
		} else {
			p.Query = nil
		}
	}

	return p
}

func (p Params) validate() error {
	switch {
	case p.Limit < minLimit:
		return fmt.Errorf("%w: limit must be at least %d, got %d", service.ErrBadRequest, minLimit, p.Limit)
	case p.Limit > maxLimit:
		return fmt.Errorf("%w: limit must not exceed %d, got %d", service.ErrBadRequest, maxLimit, p.Limit)
	case p.PriceFrom != nil && p.PriceTo != nil && *p.PriceFrom > *p.PriceTo:
		return fmt.Errorf("%w: price_from cannot be greater than price_to", service.ErrBadRequest)
	case p.Sort != "" && !p.Sort.Valid():
		return fmt.Errorf("%w: unknown sort %q", service.ErrBadRequest, p.Sort)
	}

	return nil
}

func (p Params) toClientParams() plate.FetchPlatesParams {
	return plate.FetchPlatesParams{
		Query:           p.Query,
		RegionID:        p.RegionID,
		PriceFrom:       p.PriceFrom,
		PriceTo:         p.PriceTo,
		ReissueIncluded: p.ReissueIncluded,
		CategoryIDs:     p.CategoryIDs,
		Sort:            string(p.Sort),
		Limit:           p.Limit,
		Cursor:          p.Cursor,
	}
}
