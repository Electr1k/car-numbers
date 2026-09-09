package fetchfeednumbers

import (
	"context"
	"data-service/internal/domain/data"
)

type numberStore interface {
	GetFeedNumbers(ctx context.Context, limit int, offset int) ([]data.FeedNumber, error)
}

// UseCase - возвращает свежих номеров
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
func (uc *UseCase) Handle(ctx context.Context, params Params) ([]data.FeedNumber, error) {
	err := params.validate()
	if err != nil {
		return nil, err
	}

	return uc.repository.GetFeedNumbers(ctx, params.Limit, params.Offset)
}
