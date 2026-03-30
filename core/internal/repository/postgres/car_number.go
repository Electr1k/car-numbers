package postgres

import (
	"context"
	"core/internal/domain"
	"core/internal/repository"
	"fmt"
	"time"
)

func (p *Postgres) UpdateOrCreate(ctx context.Context, number *domain.CarNumber) (*domain.CarNumber, error) {
	row := p.pool.QueryRow(ctx,
		`
			INSERT INTO car_numbers (number, price, posted_at) 
			VALUES ($1, $2, $3)
			ON CONFLICT (number) 
			DO UPDATE SET 
				price = EXCLUDED.price,
				posted_at = EXCLUDED.posted_at,
				updated_at = CURRENT_TIMESTAMP
			RETURNING number, price, posted_at;
		`, number.Number, number.Price, number.PostedAt)

	var (
		carNumber string
		price     float32
		postedAt  time.Time
	)

	if err := row.Scan(&carNumber, &price, &postedAt); err != nil {
		return nil, fmt.Errorf("%w: %v", repository.ErrQueryFailed, err)
	}

	dto, err := domain.NewCarNumber(carNumber, price, postedAt)
	if err != nil {
		return nil, fmt.Errorf("create domain object: %w", err)
	}

	return dto, nil
}
