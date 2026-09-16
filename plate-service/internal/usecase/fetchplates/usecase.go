package fetchplates

import (
	"context"
	"plate-service/internal/domain/data"
	"plate-service/internal/repository"
	"strconv"
)

type plateStore interface {
	GetPlates(ctx context.Context, params repository.GetPlatesParams) ([]data.FeedPlate, bool, error)
}

// UseCase - возвращает свежие номера
type UseCase struct {
	repository plateStore
}

func New(
	repository plateStore,
) *UseCase {
	return &UseCase{
		repository: repository,
	}
}

// Handle - возвращает список свежих номеров
func (uc *UseCase) Handle(ctx context.Context, params Params) (Result, error) {
	err := params.validate()
	if err != nil {
		return Result{}, err
	}

	plates, hasNext, err := uc.repository.GetPlates(ctx, repository.GetPlatesParams(params))
	if err != nil {
		return Result{}, err
	}

	if !hasNext {
		return Result{Plates: plates}, nil
	}

	return Result{
		Plates:     plates,
		NextCursor: nextCursor(params.Sort, plates[len(plates)-1]),
	}, nil
}

func nextCursor(sort data.PlateSort, last data.FeedPlate) *data.FeedCursor {
	cursor := &data.FeedCursor{Sort: sort, ID: last.ID}

	switch sort {
	case data.PlateSortPriceAsc, data.PlateSortPriceDesc:
		if last.Price != nil {
			price := strconv.FormatFloat(*last.Price, 'f', -1, 64)
			cursor.Price = &price
		}
	default:
		cursor.RefreshedAt = &last.RefreshedAt
		cursor.UpdatedAt = &last.UpdatedAt
	}

	return cursor
}
