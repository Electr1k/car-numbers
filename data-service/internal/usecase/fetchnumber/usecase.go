package fetchnumber

import (
	"context"
	"data-service/internal/domain/data"

	"github.com/google/uuid"
)

type numberStore interface {
	GetNumberWithOffersByID(ctx context.Context, id uuid.UUID) (*data.Number, error)
}

// UseCase - возвращает номер с предложениями
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

// Handle - возвращает номер с предложениями по идентификатору
func (uc *UseCase) Handle(ctx context.Context, id uuid.UUID) (*data.Number, error) {
	return uc.repository.GetNumberWithOffersByID(ctx, id)
}
