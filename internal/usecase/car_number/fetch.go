package price

import (
	"car-numers/internal/repository"
	proveders "car-numers/pkg/providers/car_number"
	"time"
)

type FetchCarNumberUseCase struct {
	provider   proveders.NumberProvider
	repository repository.CarNumberRepository
}

func NewFetchPricesUseCase(
	provider proveders.NumberProvider,
	repository repository.CarNumberRepository,
) *FetchCarNumberUseCase {
	return &FetchCarNumberUseCase{
		provider:   provider,
		repository: repository,
	}
}

func (uc *FetchCarNumberUseCase) Handle() error {

	for {
		data, err := uc.provider.FetchNumbers()

		if err != nil {
			panic(err)
		}

		for _, number := range data {
			_, err := uc.repository.UpdateOrCreate(&number)
			if err != nil {
				return err
			}
		}

		if len(data) == 0 {
			return nil
		}

		time.Sleep(1 * time.Second)
	}
}
