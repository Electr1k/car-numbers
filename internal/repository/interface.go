package repository

import "car-numers/internal/domain"

type CarNumberRepository interface {
	UpdateOrCreate(carNumber *domain.CarNumber) (*domain.CarNumber, error)
}
