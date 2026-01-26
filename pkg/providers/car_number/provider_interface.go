package car_number

import "car-numers/internal/domain"

type NumberProvider interface {
	FetchNumbers() ([]domain.CarNumber, error)
}
