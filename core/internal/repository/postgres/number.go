package postgres

import (
	"context"
	"core/internal/domain"
	"fmt"
	"time"
)

type NumberRepository struct {
	postgres *Postgres
}

func NewNumberRepository(postgres *Postgres) *NumberRepository {
	return &NumberRepository{postgres: postgres}
}

func (n *NumberRepository) UpdateOrCreate(ctx context.Context, number *domain.Number) (*domain.Number, error) {
	row := n.postgres.pool.QueryRow(ctx,
		`INSERT INTO numbers (number, type)
				VALUES ($1, $2)
				ON CONFLICT (number, type)
				DO UPDATE SET
					updated_at = CURRENT_TIMESTAMP
				RETURNING id, number, type, created_at, updated_at;`,
		number.Number,
		number.Type,
	)

	var (
		id            int64
		vehicleNumber string
		vehicleType   string
		createdAt     time.Time
		updatedAt     time.Time
	)

	if err := row.Scan(&id, &vehicleNumber, &vehicleType, &createdAt, &updatedAt); err != nil {
		return nil, fmt.Errorf("number create failed: %v", err)
	}

	dto, err := domain.NewNumber(&id, vehicleNumber, vehicleType, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("create domain object: %w", err)
	}

	return dto, nil
}
