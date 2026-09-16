package fetchplates

import (
	"context"
	"data-service/internal/domain/data"
	"data-service/internal/repository"
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

	last := plates[len(plates)-1]

	return Result{
		Plates: plates,
		NextCursor: &data.FeedCursor{
			RefreshedAt: last.RefreshedAt,
			UpdatedAt:   last.UpdatedAt,
			ID:          last.ID,
		},
	}, nil
}
