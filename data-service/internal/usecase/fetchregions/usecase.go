package fetchregions

import (
	"context"
	"data-service/internal/domain"
)

type regionStore interface {
	GetRegions(ctx context.Context) ([]domain.RegionWithCodes, error)
}

// UseCase - возвращает список регионов
type UseCase struct {
	repository regionStore
}

func New(
	repository regionStore,
) *UseCase {
	return &UseCase{
		repository: repository,
	}
}

// Handle - возвращает список регионов и их кодов
func (uc *UseCase) Handle(ctx context.Context) ([]domain.RegionWithCodes, error) {
	return uc.repository.GetRegions(ctx)
}
