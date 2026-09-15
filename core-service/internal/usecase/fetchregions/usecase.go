package fetchregions

import (
	"context"
	"core-service/internal/service/plate"
)

type regionsProvider interface {
	FetchRegions(ctx context.Context) (*plate.FetchRegionsResponse, error)
}

// UseCase - возвращает список регионов с кодами
type UseCase struct {
	regionProvider regionsProvider
}

func New(
	regionProvider regionsProvider,
) *UseCase {
	return &UseCase{
		regionProvider: regionProvider,
	}
}

// Handle - возвращает список регионов с кодами
func (uc *UseCase) Handle(ctx context.Context) (*Result, error) {
	regions, err := uc.regionProvider.FetchRegions(ctx)
	if err != nil {
		return nil, err
	}

	result := Result{
		Items: make([]Region, len(regions.Items)),
	}

	for i, item := range regions.Items {
		result.Items[i] = Region{
			ID:    item.ID,
			Name:  item.Name,
			Codes: item.Codes,
		}
	}

	return &result, nil
}
