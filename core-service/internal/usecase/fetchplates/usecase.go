package fetchplates

import (
	"context"
	"core-service/internal/service/plate"
)

type plateProvider interface {
	FetchPlates(ctx context.Context, params plate.FetchPlatesParams) (*plate.FetchPlatesResponse, error)
}

// UseCase - возвращает номер с предложениями
type UseCase struct {
	plateClient plateProvider
}

func New(
	plateClient plateProvider,
) *UseCase {
	return &UseCase{
		plateClient: plateClient,
	}
}

// Handle - возвращает список номеров
func (uc *UseCase) Handle(ctx context.Context, params Params) (*Result, error) {
	if err := params.validate(); err != nil {
		return nil, err
	}

	plates, err := uc.plateClient.FetchPlates(ctx, params.toClientParams())
	if err != nil {
		return nil, err
	}

	result := Result{
		Items:      make([]Plate, len(plates.Items)),
		NextCursor: plates.NextCursor,
	}
	for i, item := range plates.Items {
		var region *Region
		if item.Region != nil {
			region = &Region{
				Code: item.Region.Code,
				Name: item.Region.Name,
			}
		}

		result.Items[i] = Plate{
			ID:              item.ID,
			Number:          item.Number,
			Region:          region,
			Price:           item.Price,
			Type:            item.Type,
			Count:           item.Count,
			RefreshedAt:     item.RefreshedAt,
			UpdatedAt:       item.UpdatedAt,
			ReissueIncluded: item.ReissueIncluded,
		}
	}

	return &result, nil
}
