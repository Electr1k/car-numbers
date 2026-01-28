package car_number

import "core/internal/domain"

type NumberProvider interface {
	FetchNumbers() ([]domain.CarNumber, error)
}
