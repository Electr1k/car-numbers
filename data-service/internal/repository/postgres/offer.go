package postgres

import (
	"context"
	"data-service/internal/domain"
	"fmt"

	"github.com/jackc/pgx/v4"
)

// OfferRepository - хранение предложений и номеров, к которым они относятся
type OfferRepository struct {
	postgres *Postgres
}

func NewOfferRepository(postgres *Postgres) *OfferRepository {
	return &OfferRepository{postgres: postgres}
}

// upsertOfferQuery - вставка номера и предложения одним запросом
const upsertOfferQuery = `
WITH upserted_number AS (
	INSERT INTO numbers (id, number, type)
	VALUES ($1, $2, $3)
	ON CONFLICT (number, type) DO UPDATE SET
		updated_at = CURRENT_TIMESTAMP
	RETURNING id
)
INSERT INTO offers (id, number_id, provider, external_id, price, status, posted_at, url, raw)
SELECT $4::uuid, upserted_number.id, $5, $6, $7, $8, $9, $10, $11
FROM upserted_number
ON CONFLICT (provider, external_id) DO UPDATE SET
	number_id  = EXCLUDED.number_id,
	price      = EXCLUDED.price,
	status     = EXCLUDED.status,
	url        = EXCLUDED.url,
	raw        = EXCLUDED.raw,
	updated_at = CURRENT_TIMESTAMP;`

// SaveBatch - Сохранение батча в одной транзакции и за один поход в базу
// Возвращает количество записанных строк: вставленных и обновлённых
func (r *OfferRepository) SaveBatch(ctx context.Context, items []domain.OfferWithNumber) (int, error) {
	if len(items) == 0 {
		return 0, nil
	}

	tx, err := r.postgres.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}
	for _, item := range items {
		// Порядок аргументов - $1..$11 запроса выше
		batch.Queue(upsertOfferQuery,
			item.Number.Id,
			item.Number.Number,
			item.Number.Type,
			item.Offer.Id,
			item.Offer.Provider,
			item.Offer.ExternalId,
			item.Offer.Price,
			item.Offer.Status,
			item.Offer.PostedAt,
			item.Offer.Url,
			item.Offer.Raw,
		)
	}

	results := tx.SendBatch(ctx, batch)

	var saved int64
	for _, item := range items {
		tag, err := results.Exec()
		if err != nil {
			results.Close()
			return 0, fmt.Errorf("save offer %s/%s: %w", item.Offer.Provider, item.Offer.ExternalId, err)
		}

		saved += tag.RowsAffected()
	}

	if err := results.Close(); err != nil {
		return 0, fmt.Errorf("close batch: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit transaction: %w", err)
	}

	return int(saved), nil
}
