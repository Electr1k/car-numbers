package fetchfeednumbers

import (
	"context"
	"data-service/internal/domain"
	"data-service/internal/domain/data"
	"fmt"
)

type numberStore interface {
	GetFeedNumbers(ctx context.Context, cursor data.FeedCursor, limit int) ([]data.FeedNumber, bool, error)
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

	var cursor data.FeedCursor
	if params.Cursor == nil {
		cursor = data.GetEmptyFeedCursor()
	} else {
		c, err := data.DecodeFeedCursor(*params.Cursor)
		if err != nil {
			return Result{}, fmt.Errorf("%w: err: %s", domain.ErrInvalidArgument, err.Error())
		}
		cursor = *c
	}

	numbers, hasNext, err := uc.repository.GetFeedNumbers(ctx, cursor, params.Limit)

	if err != nil {
		return Result{}, err
	}

	if !hasNext {
		return Result{Numbers: numbers, Cursor: nil}, err
	}

	lastNumber := numbers[params.Limit-1]
	newCursor := data.EncodeFeedCursor(lastNumber.RefreshedAt, lastNumber.UpdatedAt, lastNumber.ID)

	return Result{Numbers: numbers, Cursor: &newCursor}, nil
}
