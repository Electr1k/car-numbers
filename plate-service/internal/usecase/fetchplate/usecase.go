package fetchplate

import (
	"context"
	"plate-service/internal/domain/data"

	"github.com/google/uuid"
)

type plateStore interface {
	GetPlateWithOffersByID(ctx context.Context, id uuid.UUID) (*data.Plate, error)
}

// UseCase - возвращает номер с предложениями
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

// Handle - возвращает номер с предложениями по идентификатору
func (uc *UseCase) Handle(ctx context.Context, id uuid.UUID) (*data.Plate, error) {
	return uc.repository.GetPlateWithOffersByID(ctx, id)
}
