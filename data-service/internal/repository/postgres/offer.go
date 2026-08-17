package postgres

import (
	"context"
	"data-service/internal/domain"
	"fmt"
	"time"

	"github.com/google/uuid"
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
	RETURNING id, number, type, created_at, updated_at
),
upserted_offer AS (
	INSERT INTO offers (id, number_id, provider, external_id, price, status, posted_at, refreshed_at, url, raw)
	SELECT $4::uuid, upserted_number.id, $5, $6, $7, $8, $9, $10, $11, $12
	FROM upserted_number
	ON CONFLICT (provider, external_id) DO UPDATE SET
		number_id    = EXCLUDED.number_id,
		price        = EXCLUDED.price,
		status       = EXCLUDED.status,
		refreshed_at = EXCLUDED.refreshed_at,
		url          = EXCLUDED.url,
		raw          = EXCLUDED.raw,
		updated_at   = CURRENT_TIMESTAMP
	RETURNING id, provider, external_id, price, status, whereabouts, reissue_included, view_count,
		posted_at, refreshed_at, url, raw, raw_detail, comment, created_at, updated_at
)
SELECT upserted_offer.id, upserted_offer.provider, upserted_offer.external_id, upserted_offer.price,
	upserted_offer.status, upserted_offer.whereabouts, upserted_offer.reissue_included, upserted_offer.view_count,
	upserted_offer.posted_at, upserted_offer.refreshed_at, upserted_offer.url, upserted_offer.raw,
	upserted_offer.raw_detail, upserted_offer.comment, upserted_offer.created_at, upserted_offer.updated_at,
	upserted_number.id, upserted_number.number, upserted_number.type,
	upserted_number.created_at, upserted_number.updated_at
FROM upserted_offer, upserted_number;`

const getOfferByIdQuery = `
SELECT 
    offers.id, provider, external_id, price, status, whereabouts, reissue_included, view_count, posted_at,
    refreshed_at, url, raw, raw_detail, comment, offers.created_at, offers.updated_at,
    numbers.id, number, type, numbers.created_at, numbers.updated_at
FROM offers
JOIN numbers ON offers.number_id = numbers.id
WHERE offers.id = $1
;`

const getOffersQuery = `
SELECT
    offers.id, provider, external_id, price, status, whereabouts, reissue_included, view_count, posted_at,
    refreshed_at, url, raw, raw_detail, comment, offers.created_at, offers.updated_at,
    numbers.id, number, type, numbers.created_at, numbers.updated_at
FROM offers
JOIN numbers ON offers.number_id = numbers.id
WHERE offers.status = $1 AND offers.provider = $2
;`

const upsertPriceHistoryQuery = `
INSERT INTO price_history (id, offer_id, number_id, price)
SELECT $1::uuid, $2::uuid, $3::uuid, $4::numeric(12,2)
WHERE $4::numeric(12,2) IS DISTINCT FROM (
	SELECT price
	FROM price_history
	WHERE offer_id = $2::uuid
	ORDER BY created_at DESC, id DESC
	LIMIT 1
)`

const updateOfferQuery = `
UPDATE offers SET
price = $2,
status = $3,
whereabouts = $4,
reissue_included = $5,
view_count = $6,
posted_at = $7,
refreshed_at = $8,
raw_detail = $9,
comment = $10,
updated_at = CURRENT_TIMESTAMP
WHERE id = $1`

// SaveBatch - Сохранение батча в одной транзакции и за один поход в базу
func (r *OfferRepository) SaveBatch(ctx context.Context, items []domain.OfferWithNumber) ([]domain.OfferWithNumber, error) {
	if len(items) == 0 {
		return nil, nil
	}

	tx, err := r.postgres.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}
	for _, item := range items {
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
			item.Offer.RefreshedAt,
			item.Offer.Url,
			item.Offer.Raw,
		)
	}

	results := tx.SendBatch(ctx, batch)

	saved := make([]domain.OfferWithNumber, 0, len(items))

	for _, item := range items {
		stored, err := scanOfferWithNumber(results.QueryRow())
		if err != nil {
			results.Close()
			return nil, fmt.Errorf("save offer %s/%s: %w", item.Offer.Provider, item.Offer.ExternalId, err)
		}

		saved = append(saved, stored)
	}

	if err := results.Close(); err != nil {
		return nil, fmt.Errorf("close batch: %w", err)
	}

	if err := r.upsertPriceHistory(ctx, tx, saved); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return saved, nil
}

// upsertPriceHistory - Запись наблюдений цены, единственное место записи в price_history
func (r *OfferRepository) upsertPriceHistory(ctx context.Context, tx pgx.Tx, items []domain.OfferWithNumber) error {
	if len(items) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, item := range items {
		history, err := domain.NewPriceHistoryFromOffer(item.Offer)
		if err != nil {
			return fmt.Errorf("build price history for offer %s: %w", item.Offer.Id, err)
		}

		batch.Queue(upsertPriceHistoryQuery, history.Id, history.OfferId, history.NumberId, history.Price)
	}

	results := tx.SendBatch(ctx, batch)

	for _, item := range items {
		if _, err := results.Exec(); err != nil {
			results.Close()
			return fmt.Errorf("track price for offer %s: %w", item.Offer.Id, err)
		}
	}

	if err := results.Close(); err != nil {
		return fmt.Errorf("close price history batch: %w", err)
	}

	return nil
}

func (r *OfferRepository) GetOfferById(ctx context.Context, id uuid.UUID) (domain.OfferWithNumber, error) {
	row := r.postgres.pool.QueryRow(ctx, getOfferByIdQuery, id)

	item, err := scanOfferWithNumber(row)
	if err != nil {
		return domain.OfferWithNumber{}, fmt.Errorf("get offer by id %s: %w", id, err)
	}

	return item, nil
}

// GetOffers - Предложения выбранного провайдера в заданном статусе вместе с их номерами
func (r *OfferRepository) GetOffers(ctx context.Context, status domain.OfferStatus, provider domain.Provider) ([]domain.OfferWithNumber, error) {
	rows, err := r.postgres.pool.Query(ctx, getOffersQuery, status, provider)
	if err != nil {
		return nil, fmt.Errorf("get offers (provider=%s, status=%s): %w", provider, status, err)
	}
	defer rows.Close()

	var offers []domain.OfferWithNumber
	for rows.Next() {
		item, err := scanOfferWithNumber(rows)
		if err != nil {
			return nil, fmt.Errorf("get offers (provider=%s, status=%s): %w", provider, status, err)
		}

		offers = append(offers, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get offers (provider=%s, status=%s): %w", provider, status, err)
	}

	return offers, nil
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

// scanOfferWithNumber - разбирает одну строку выборки offers JOIN numbers в домен
func scanOfferWithNumber(row rowScanner) (domain.OfferWithNumber, error) {
	var (
		offerId         uuid.UUID
		provider        string
		externalId      string
		price           *float64
		status          string
		whereabouts     *string
		reissueIncluded *bool
		viewCount       *int
		postedAt        *time.Time
		refreshedAt     *time.Time
		url             string
		raw             string
		rawDetailed     *string
		comment         *string
		offerCreatedAt  *time.Time
		offerUpdatedAt  *time.Time
		numberId        uuid.UUID
		number          string
		vehicleType     string
		numberCreatedAt *time.Time
		numberUpdatedAt *time.Time
	)

	err := row.Scan(&offerId, &provider, &externalId, &price, &status, &whereabouts, &reissueIncluded, &viewCount, &postedAt,
		&refreshedAt, &url, &raw, &rawDetailed, &comment, &offerCreatedAt, &offerUpdatedAt, &numberId, &number, &vehicleType, &numberCreatedAt, &numberUpdatedAt)
	if err != nil {
		return domain.OfferWithNumber{}, err
	}

	n, err := domain.RestoreNumber(numberId, number, domain.NumberType(vehicleType), numberCreatedAt, numberUpdatedAt)
	if err != nil {
		return domain.OfferWithNumber{}, fmt.Errorf("restore number %s: %w", numberId, err)
	}

	var whereaboutsVO *domain.OfferWhereabouts
	if whereabouts != nil {
		vo := domain.OfferWhereabouts(*whereabouts)
		whereaboutsVO = &vo
	}

	offer, err := domain.RestoreOffer(
		offerId, numberId, domain.Provider(provider), externalId, price, domain.OfferStatus(status), whereaboutsVO, reissueIncluded,
		viewCount, postedAt, refreshedAt, url, raw, rawDetailed, comment, offerCreatedAt, offerUpdatedAt,
	)
	if err != nil {
		return domain.OfferWithNumber{}, fmt.Errorf("restore offer %s: %w", offerId, err)
	}

	return domain.OfferWithNumber{Number: n, Offer: offer}, nil
}

func (r *OfferRepository) UpdateOffer(ctx context.Context, offer *domain.Offer) error {
	tx, err := r.postgres.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, updateOfferQuery, offer.Id, offer.Price, offer.Status, offer.Whereabouts,
		offer.ReissueIncluded, offer.ViewCount, offer.PostedAt, offer.RefreshedAt,
		offer.RawDetailed, offer.Comment,
	)
	if err != nil {
		return fmt.Errorf("update offer %s: %w", offer.Id, err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("update offer %s: offer not found", offer.Id)
	}

	if err := r.upsertPriceHistory(ctx, tx, []domain.OfferWithNumber{{Offer: offer}}); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
