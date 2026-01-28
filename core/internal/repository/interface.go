package repository

import "core/internal/domain"

type CarNumberRepository interface {
	UpdateOrCreate(carNumber *domain.CarNumber) (*domain.CarNumber, error)
}
