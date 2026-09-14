package fetchfeednumbers

import (
	"context"
	"data-service/internal/domain/data"
	"data-service/internal/repository"
)

type numberStore interface {
	GetFeedNumbers(ctx context.Context, params repository.GetNumbersParams) ([]data.FeedNumber, bool, error)
}

// UseCase - возвращает свежие номера
type UseCase struct {
	repository numberStore
}

func New(
	repository numberStore,
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

	numbers, hasNext, err := uc.repository.GetFeedNumbers(ctx, repository.GetNumbersParams(params))
	if err != nil {
		return Result{}, err
	}

	if !hasNext {
		return Result{Numbers: numbers}, nil
	}

	last := numbers[len(numbers)-1]

	return Result{
		Numbers: numbers,
		NextCursor: &data.FeedCursor{
			RefreshedAt: last.RefreshedAt,
			UpdatedAt:   last.UpdatedAt,
			ID:          last.ID,
		},
	}, nil
}
