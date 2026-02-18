package price

import (
	"core/internal/domain"
	"fmt"
	"time"
)

const (
	batchSize     = 20
	timeSleep     = 1 * time.Second
	startPosition = 0
)

var (
	ErrFetchFailed = fmt.Errorf("failed to fetch car numbers")
	ErrSaveFailed  = fmt.Errorf("failed to save car number")
)

type CarNumberRepository interface {
	UpdateOrCreate(carNumber *domain.CarNumber) (*domain.CarNumber, error)
}

type NumberProvider interface {
	FetchNumbers(start int) ([]domain.CarNumber, error)
}

type FetchCarNumberUseCase struct {
	provider   NumberProvider
	repository CarNumberRepository
}

func NewFetchPricesUseCase(
	provider NumberProvider,
	repository CarNumberRepository,
) *FetchCarNumberUseCase {
	return &FetchCarNumberUseCase{
		provider:   provider,
		repository: repository,
	}
}

func (uc *FetchCarNumberUseCase) Handle() error {
	currentStart := startPosition
	lastProcessedTime := time.Now()

	for {
		data, err := uc.provider.FetchNumbers(currentStart)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrFetchFailed, err)
		}

		for _, number := range data {
			_, err := uc.repository.UpdateOrCreate(&number)
			if err != nil {
				return fmt.Errorf("%w: %v", ErrSaveFailed, err)
			}
		}

		fmt.Printf("\nFetched %d car numbers; ", len(data))
		if len(data) > 0 {
			fmt.Printf("Last number: %s %f %s\n", data[len(data)-1].Number, data[len(data)-1].Price, data[len(data)-1].PostedAt.Format(time.DateOnly))
			lastProcessedTime = data[len(data)-1].PostedAt

			if lastProcessedTime.Before(time.Now().AddDate(0, 0, -3)) {
				fmt.Printf("Last processed time: %s\n", lastProcessedTime.Format(time.DateOnly))
				break
			}
		}

		currentStart += batchSize

		time.Sleep(timeSleep)
	}

	return nil
}
