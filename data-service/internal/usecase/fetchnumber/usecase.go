package fetchnumber

import (
	"context"
	"data-service/internal/domain/data"

	"github.com/google/uuid"
)

type numberStore interface {
	GetNumberWithOffersById(ctx context.Context, id uuid.UUID) (*data.Number, error)
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
func (uc *UseCase) Handle(ctx context.Context, id uuid.UUID) (*data.Number, error) {

	return uc.repository.GetNumberWithOffersById(ctx, id)
}
