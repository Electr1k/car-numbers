package postgres

import (
	"car-numers/internal/domain"
	"context"
	"time"
)

func (p *Postgres) UpdateOrCreate(number *domain.CarNumber) (*domain.CarNumber, error) {

	row := p.pool.QueryRow(context.Background(),
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

	if err := row.Scan(
		&carNumber,
		&price,
		&postedAt,
	); err != nil {
		return nil, err
	}

	dto, err := domain.NewCarNumber(
		carNumber,
		price,
		postedAt,
	)

	if err != nil {
		return nil, err
	}

	return dto, nil
}
