package postgres

import (
	"context"
	"core/internal/domain"
	"fmt"
	"time"
)

type OfferRepository struct {
	postgres *Postgres
}

func NewOfferRepository(postgres *Postgres) *OfferRepository {
	return &OfferRepository{postgres: postgres}
}

func (o *OfferRepository) UpdateOrCreate(ctx context.Context, offer *domain.Offer) (*domain.Offer, error) {
	row := o.postgres.pool.QueryRow(ctx,
		`INSERT INTO offers (number_id, provider, external_id, price, status, posted_at, url, raw)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				ON CONFLICT (provider, external_id)
				DO UPDATE SET
					number_id = EXCLUDED.number_id,
					price = EXCLUDED.price,
					status = EXCLUDED.status,
					url = EXCLUDED.url,
					raw = EXCLUDED.raw,
					updated_at = CURRENT_TIMESTAMP
				RETURNING id, number_id, provider, external_id, price, status, posted_at, url, raw, created_at, updated_at;`,
		offer.NumberId,
		offer.Provider,
		offer.ExternalId,
		offer.Price,
		offer.Status,
		offer.PostedAt,
		offer.Url,
		offer.Raw,
	)

	var (
		id         int64
		numberId   int64
		provider   string
		externalId string
		price      float32
		status     string
		postedAt   time.Time
		url        string
		raw        string
		createdAt  time.Time
		updatedAt  time.Time
	)

	if err := row.Scan(&id, &numberId, &provider, &externalId, &price, &status, &postedAt, &url, &raw, &createdAt, &updatedAt); err != nil {
		return nil, fmt.Errorf("offer create failed: %v", err)
	}

	dto, err := domain.NewOffer(&id, &numberId, provider, externalId, price, status, &postedAt, url, raw, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("create domain object: %w", err)
	}

	return dto, nil
}
