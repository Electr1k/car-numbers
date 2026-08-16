package postgres

import (
	"context"
	"data-service/internal/domain"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

// FeatureRepository - хранение фичей
type FeatureRepository struct {
	postgres *Postgres
}

func NewFeatureRepository(postgres *Postgres) *FeatureRepository {
	return &FeatureRepository{postgres: postgres}
}

const getFeatureByKeyQuery = `
SELECT id, key, name, active, created_at, updated_at
FROM features
WHERE key = $1
LIMIT 1
`

func (r *FeatureRepository) GetFeatureByKey(ctx context.Context, key domain.FeatureKey) (*domain.Feature, error) {
	row := r.postgres.pool.QueryRow(ctx, getFeatureByKeyQuery, string(key))

	var (
		id         uuid.UUID
		keyFeature string
		name       string
		active     bool
		createdAt  *time.Time
		updatedAt  *time.Time
	)

	err := row.Scan(&id, &keyFeature, &name, &active, &createdAt, &updatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrFeatureNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("error on get feature: %w", err)
	}

	feature, err := domain.RestoreFeature(
		id,
		domain.FeatureKey(keyFeature),
		name,
		active,
		createdAt,
		updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error on restore feature: %w", err)
	}

	return feature, nil
}
