package postgres

import (
	"context"
	"errors"
	"fmt"
	"plate-service/internal/domain"
	"plate-service/internal/domain/data"
	"plate-service/internal/repository"
	"strings"
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
WITH upserted_plate AS (
	INSERT INTO plates (id, number, type)
	VALUES ($1, $2, $3)
	ON CONFLICT (number, type) DO UPDATE SET
		updated_at = CURRENT_TIMESTAMP
	RETURNING id, number, type, created_at, updated_at
),
upserted_offer AS (
	INSERT INTO offers (id, plate_id, provider, external_id, price, status, posted_at, refreshed_at, url, raw)
	SELECT $4::uuid, upserted_plate.id, $5, $6, $7, $8, $9, $10, $11, $12
	FROM upserted_plate
	ON CONFLICT (provider, external_id) DO UPDATE SET
		plate_id     = EXCLUDED.plate_id,
		price        = EXCLUDED.price,
		status       = EXCLUDED.status,
		posted_at    = LEAST(offers.posted_at, EXCLUDED.posted_at),
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
	upserted_plate.id, upserted_plate.number, upserted_plate.type,
	upserted_plate.created_at, upserted_plate.updated_at
FROM upserted_offer, upserted_plate;`

// upsertOfferWithDetailQuery - вставка номера и предложения вместе с деталкой одним запросом
const upsertOfferWithDetailQuery = `
WITH upserted_plate AS (
	INSERT INTO plates (id, number, type)
	VALUES ($1, $2, $3)
	ON CONFLICT (number, type) DO UPDATE SET
		updated_at = CURRENT_TIMESTAMP
	RETURNING id, number, type, created_at, updated_at
),
upserted_offer AS (
	INSERT INTO offers (id, plate_id, provider, external_id, price, status, whereabouts, reissue_included,
		view_count, posted_at, refreshed_at, url, raw, raw_detail, comment)
	SELECT $4::uuid, upserted_plate.id, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
	FROM upserted_plate
	ON CONFLICT (provider, external_id) DO UPDATE SET
		plate_id         = EXCLUDED.plate_id,
		price            = EXCLUDED.price,
		status           = EXCLUDED.status,
		whereabouts      = EXCLUDED.whereabouts,
		reissue_included = EXCLUDED.reissue_included,
		view_count       = EXCLUDED.view_count,
		posted_at        = LEAST(offers.posted_at, EXCLUDED.posted_at),
		refreshed_at     = EXCLUDED.refreshed_at,
		url              = EXCLUDED.url,
		raw              = EXCLUDED.raw,
		raw_detail       = EXCLUDED.raw_detail,
		comment          = EXCLUDED.comment,
		updated_at       = CURRENT_TIMESTAMP
	RETURNING id, provider, external_id, price, status, whereabouts, reissue_included, view_count,
		posted_at, refreshed_at, url, raw, raw_detail, comment, created_at, updated_at
)
SELECT upserted_offer.id, upserted_offer.provider, upserted_offer.external_id, upserted_offer.price,
	upserted_offer.status, upserted_offer.whereabouts, upserted_offer.reissue_included, upserted_offer.view_count,
	upserted_offer.posted_at, upserted_offer.refreshed_at, upserted_offer.url, upserted_offer.raw,
	upserted_offer.raw_detail, upserted_offer.comment, upserted_offer.created_at, upserted_offer.updated_at,
	upserted_plate.id, upserted_plate.number, upserted_plate.type,
	upserted_plate.created_at, upserted_plate.updated_at
FROM upserted_offer, upserted_plate;`

const getOfferByIdQuery = `
SELECT 
    offers.id, provider, external_id, price, status, whereabouts, reissue_included, view_count, posted_at,
    refreshed_at, url, raw, raw_detail, comment, offers.created_at, offers.updated_at,
    plates.id, number, type, plates.created_at, plates.updated_at
FROM offers
JOIN plates ON offers.plate_id = plates.id
WHERE offers.id = $1
;`

const getOfferByExternalIdQuery = `
SELECT
    offers.id, provider, external_id, price, status, whereabouts, reissue_included, view_count, posted_at,
    refreshed_at, url, raw, raw_detail, comment, offers.created_at, offers.updated_at,
    plates.id, number, type, plates.created_at, plates.updated_at
FROM offers
JOIN plates ON offers.plate_id = plates.id
WHERE offers.provider = $1 AND offers.external_id = $2
;`

const getOffersQuery = `
SELECT
    offers.id, provider, external_id, price, status, whereabouts, reissue_included, view_count, posted_at,
    refreshed_at, url, raw, raw_detail, comment, offers.created_at, offers.updated_at,
    plates.id, number, type, plates.created_at, plates.updated_at
FROM offers
JOIN plates ON offers.plate_id = plates.id
WHERE offers.provider = $1 AND offers.status = $2
;`

const getOfferIdsQuery = `
SELECT offers.id
FROM offers
WHERE offers.provider = $1 AND offers.status = $2
ORDER BY posted_at
;`

const upsertPriceHistoryQuery = `
INSERT INTO price_history (id, offer_id, plate_id, price)
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

const getFeedPlates = `
SELECT 
	plates.id,
	number,
	regions.id AS region_id,
	regions.name,
	region_codes.code,
	MIN(price),
	type,
	count(*) as count,
	MAX(refreshed_at::date) as refreshed_at,
	MAX(offers.updated_at) as updated_at,
	CASE
		WHEN COUNT(CASE WHEN reissue_included = true THEN 1 END) > 0 THEN true
		WHEN COUNT(CASE WHEN reissue_included = false THEN 1 END) > 0 THEN false
		ELSE null
	END as reissue_included
FROM public.plates
JOIN offers ON offers.plate_id = plates.id
LEFT JOIN region_codes ON plates.region_code = region_codes.code
LEFT JOIN regions ON regions.id = region_codes.region_id
WHERE status = $1
GROUP BY plates.id, number, regions.id, regions.name, region_codes.code, type
HAVING ($2::date IS NULL OR (MAX(refreshed_at::date), MAX(offers.updated_at), plates.id) < ($2::date, $3::timestamptz, $4::uuid))
AND ($5::TEXT IS NULL OR number LIKE $5::TEXT)
AND ($6::BIGINT IS NULL OR regions.id = $6::BIGINT)
AND ($7::FLOAT IS NULL OR MIN(price) >= $7::FLOAT)
AND ($8::FLOAT IS NULL OR MIN(price) <= $8::FLOAT)
AND ($9::BOOLEAN IS NULL OR 
	CASE WHEN COUNT(CASE WHEN reissue_included = true THEN 1 END) > 0 THEN true 
		WHEN COUNT(CASE WHEN reissue_included = false THEN 1 END) > 0 THEN false
		ELSE null
	END = $9::BOOLEAN)
ORDER BY refreshed_at DESC, updated_at DESC, id DESC
LIMIT $10
`

const getPlateWithOffersByID = `
SELECT 
	plates.id as id,
	number,
	regions.id AS region_id,
	regions.name,
	region_codes.code,
	type,
	offers.id as offer_id,
	provider,
	price,
	status,
	reissue_included,
	whereabouts,
	view_count,
	comment,
	posted_at,
	refreshed_at,
	url
FROM public.plates
LEFT JOIN offers ON offers.plate_id = plates.id
LEFT JOIN region_codes ON plates.region_code = region_codes.code
LEFT JOIN regions ON regions.id = region_codes.region_id
WHERE plates.id = $1
`

// SaveBatch - Сохранение батча в одной транзакции и за один поход в базу
func (r *OfferRepository) SaveBatch(ctx context.Context, items []domain.OfferWithPlate) ([]domain.OfferWithPlate, error) {
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
			item.Plate.ID,
			item.Plate.Number,
			item.Plate.Type,
			item.Offer.ID,
			item.Offer.Provider,
			item.Offer.ExternalID,
			item.Offer.Price,
			item.Offer.Status,
			item.Offer.PostedAt,
			item.Offer.RefreshedAt,
			item.Offer.URL,
			item.Offer.Raw,
		)
	}

	results := tx.SendBatch(ctx, batch)

	saved := make([]domain.OfferWithPlate, 0, len(items))

	for _, item := range items {
		stored, err := scanOfferWithPlate(results.QueryRow())
		if err != nil {
			results.Close()
			return nil, fmt.Errorf("save offer %s/%s: %w", item.Offer.Provider, item.Offer.ExternalID, err)
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

// UpdateOrCreate - Сохранение одного предложения вместе с номером и деталкой
func (r *OfferRepository) UpdateOrCreate(ctx context.Context, item domain.OfferWithPlate) (domain.OfferWithPlate, error) {
	if item.Plate == nil || item.Offer == nil {
		return domain.OfferWithPlate{}, fmt.Errorf("update or create offer: empty offer or plate")
	}

	tx, err := r.postgres.pool.Begin(ctx)
	if err != nil {
		return domain.OfferWithPlate{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, upsertOfferWithDetailQuery,
		item.Plate.ID,
		item.Plate.Number,
		item.Plate.Type,
		item.Offer.ID,
		item.Offer.Provider,
		item.Offer.ExternalID,
		item.Offer.Price,
		item.Offer.Status,
		item.Offer.Whereabouts,
		item.Offer.ReissueIncluded,
		item.Offer.ViewCount,
		item.Offer.PostedAt,
		item.Offer.RefreshedAt,
		item.Offer.URL,
		item.Offer.Raw,
		item.Offer.RawDetailed,
		item.Offer.Comment,
	)

	saved, err := scanOfferWithPlate(row)
	if err != nil {
		return domain.OfferWithPlate{}, fmt.Errorf("save offer %s/%s: %w", item.Offer.Provider, item.Offer.ExternalID, err)
	}

	if err := r.upsertPriceHistory(ctx, tx, []domain.OfferWithPlate{saved}); err != nil {
		return domain.OfferWithPlate{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.OfferWithPlate{}, fmt.Errorf("commit transaction: %w", err)
	}

	return saved, nil
}

// upsertPriceHistory - Запись наблюдений цены, единственное место записи в price_history
func (r *OfferRepository) upsertPriceHistory(ctx context.Context, tx pgx.Tx, items []domain.OfferWithPlate) error {
	if len(items) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, item := range items {
		history, err := domain.NewPriceHistoryFromOffer(item.Offer)
		if err != nil {
			return fmt.Errorf("build price history for offer %s: %w", item.Offer.ID, err)
		}

		batch.Queue(upsertPriceHistoryQuery, history.ID, history.OfferID, history.PlateID, history.Price)
	}

	results := tx.SendBatch(ctx, batch)

	for _, item := range items {
		if _, err := results.Exec(); err != nil {
			results.Close()
			return fmt.Errorf("track price for offer %s: %w", item.Offer.ID, err)
		}
	}

	if err := results.Close(); err != nil {
		return fmt.Errorf("close price history batch: %w", err)
	}

	return nil
}

func (r *OfferRepository) GetOfferByID(ctx context.Context, id uuid.UUID) (domain.OfferWithPlate, error) {
	row := r.postgres.pool.QueryRow(ctx, getOfferByIdQuery, id)

	item, err := scanOfferWithPlate(row)
	if err != nil {
		return domain.OfferWithPlate{}, fmt.Errorf("get offer by id %s: %w", id, err)
	}

	return item, nil
}

// GetOfferByExternalID - Предложение по идентификатору у провайдера, domain.ErrOfferNotFound если его ещё нет
func (r *OfferRepository) GetOfferByExternalID(ctx context.Context, provider domain.Provider, externalID string) (domain.OfferWithPlate, error) {
	row := r.postgres.pool.QueryRow(ctx, getOfferByExternalIdQuery, provider, externalID)

	item, err := scanOfferWithPlate(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.OfferWithPlate{}, domain.ErrOfferNotFound
	}
	if err != nil {
		return domain.OfferWithPlate{}, fmt.Errorf("get offer by external id %s/%s: %w", provider, externalID, err)
	}

	return item, nil
}

// GetOffers - Предложения выбранного провайдера в заданном статусе вместе с их номерами
func (r *OfferRepository) GetOffers(ctx context.Context, status domain.OfferStatus, provider domain.Provider) ([]domain.OfferWithPlate, error) {
	rows, err := r.postgres.pool.Query(ctx, getOffersQuery, provider, status)
	if err != nil {
		return nil, fmt.Errorf("get offers (provider=%s, status=%s): %w", provider, status, err)
	}
	defer rows.Close()

	var offers []domain.OfferWithPlate
	for rows.Next() {
		item, err := scanOfferWithPlate(rows)
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

// GetOfferIdsByProviderAndStatus - возвращает идентификаторы предложений по провайдеру и статусу
func (r *OfferRepository) GetOfferIdsByProviderAndStatus(ctx context.Context, provider domain.Provider, status domain.OfferStatus) ([]uuid.UUID, error) {
	rows, err := r.postgres.pool.Query(ctx, getOfferIdsQuery, provider, status)
	if err != nil {
		return nil, fmt.Errorf("get offer ids (provider=%s, status=%s): %w", provider, status, err)
	}
	defer rows.Close()

	var offersIds []uuid.UUID
	for rows.Next() {
		var offerID uuid.UUID

		if err := rows.Scan(&offerID); err != nil {
			return nil, fmt.Errorf("get offer ids (provider=%s, status=%s): %w", provider, status, err)
		}

		offersIds = append(offersIds, offerID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get offer ids (provider=%s, status=%s): %w", provider, status, err)
	}

	return offersIds, nil
}

func (r *OfferRepository) UpdateOffer(ctx context.Context, offer *domain.Offer) error {
	tx, err := r.postgres.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, updateOfferQuery, offer.ID, offer.Price, offer.Status, offer.Whereabouts,
		offer.ReissueIncluded, offer.ViewCount, offer.PostedAt, offer.RefreshedAt,
		offer.RawDetailed, offer.Comment,
	)
	if err != nil {
		return fmt.Errorf("update offer %s: %w", offer.ID, err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("update offer %s: offer not found", offer.ID)
	}

	if err := r.upsertPriceHistory(ctx, tx, []domain.OfferWithPlate{{Offer: offer}}); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// GetPlates - Возвращает список номеров
func (r *OfferRepository) GetPlates(ctx context.Context, params repository.GetPlatesParams) ([]data.FeedPlate, bool, error) {

	var query *string
	if params.Query != nil && len(*params.Query) > 0 {
		str := strings.ReplaceAll(*params.Query, "*", "_") + "%"
		query = &str
	}

	var (
		afterRefreshedAt *time.Time
		afterUpdatedAt   *time.Time
		afterID          *uuid.UUID
	)
	if params.Cursor != nil {
		afterRefreshedAt, afterUpdatedAt, afterID = &params.Cursor.RefreshedAt, &params.Cursor.UpdatedAt, &params.Cursor.ID
	}

	rows, err := r.postgres.pool.Query(
		ctx,
		getFeedPlates,
		string(domain.OfferStatusActive),
		afterRefreshedAt,
		afterUpdatedAt,
		afterID,
		query,
		params.RegionId,
		params.PriceFrom,
		params.PriceTo,
		params.ReissueIncluded,
		params.Limit+1,
	)
	if err != nil {
		return nil, false, fmt.Errorf("get feed plates (limit=%d, refreshed_at:%s, updated_at:%s, id:%s): %w",
			params.Limit,
			afterRefreshedAt,
			afterUpdatedAt,
			afterID,
			err,
		)
	}
	defer rows.Close()

	plates := make([]data.FeedPlate, 0, params.Limit+1)
	for rows.Next() {
		var (
			id              uuid.UUID
			number          string
			regionID        *int
			regionName      *string
			regionCode      *string
			price           *float64
			plateType       string
			count           int
			refreshedAt     time.Time
			updatedAt       time.Time
			reissueIncluded *bool
		)

		if err := rows.Scan(&id, &number, &regionID, &regionName, &regionCode, &price, &plateType, &count, &refreshedAt, &updatedAt, &reissueIncluded); err != nil {
			return nil, false, fmt.Errorf("get feed plates (limit=%d, refreshed_at:%s, updated_at:%s, id:%s): %w",
				params.Limit,
				afterRefreshedAt,
				afterUpdatedAt,
				afterID,
				err,
			)
		}

		plates = append(plates, data.FeedPlate{
			ID:              id,
			Number:          number,
			RegionID:        regionID,
			RegionName:      regionName,
			RegionCode:      regionCode,
			Price:           price,
			Type:            domain.PlateType(plateType),
			Count:           count,
			RefreshedAt:     refreshedAt,
			UpdatedAt:       updatedAt,
			ReissueIncluded: reissueIncluded,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("get feed plates (limit=%d, refreshed_at:%s, updated_at:%s, id:%s): %w",
			params.Limit,
			afterRefreshedAt,
			afterUpdatedAt,
			afterID,
			err,
		)
	}

	hasNext := len(plates) > params.Limit
	if hasNext {
		plates = plates[:params.Limit]
	}

	return plates, hasNext, nil
}

func (r *OfferRepository) GetPlateWithOffersByID(ctx context.Context, id uuid.UUID) (*data.Plate, error) {
	rows, err := r.postgres.pool.Query(ctx, getPlateWithOffersByID, id)
	if err != nil {
		return nil, fmt.Errorf("get plate with offers (id=%s): %w", id, err)
	}
	defer rows.Close()

	offers := make([]data.Offer, 0)
	hasResult := false
	var (
		plateID    uuid.UUID
		number     string
		regionID   *int
		regionName *string
		regionCode *string
		plateType  string
	)

	for rows.Next() {
		var (
			offerID         *uuid.UUID
			provider        *string
			price           *float64
			status          *string
			reissueIncluded *bool
			whereabouts     *string
			viewCount       *int
			comment         *string
			postedAt        *time.Time
			refreshedAt     *time.Time
			url             *string
		)

		if err = rows.Scan(&plateID, &number, &regionID, &regionName, &regionCode, &plateType, &offerID, &provider, &price, &status,
			&reissueIncluded, &whereabouts, &viewCount, &comment, &postedAt, &refreshedAt, &url); err != nil {
			return nil, fmt.Errorf("get plate with offers (id=%s): %w", id, err)
		}

		hasResult = true

		var whereaboutsVo *domain.OfferWhereabouts
		if whereabouts != nil {
			w := domain.OfferWhereabouts(*whereabouts)
			whereaboutsVo = &w
		}
		if viewCount != nil && *viewCount == 0 {
			viewCount = nil
		}

		if offerID != nil {
			offers = append(offers, data.Offer{
				ID:              *offerID,
				Provider:        domain.Provider(*provider),
				Price:           price,
				Status:          domain.OfferStatus(*status),
				ReissueIncluded: reissueIncluded,
				Whereabouts:     whereaboutsVo,
				ViewCount:       viewCount,
				Comment:         comment,
				PostedAt:        *postedAt,
				RefreshedAt:     *refreshedAt,
				URL:             *url,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get plate with offers (id=%s): %w", id, err)
	}

	if !hasResult {
		return nil, domain.ErrPlateNotFound
	}

	return &data.Plate{
		ID:         plateID,
		Number:     number,
		RegionID:   regionID,
		RegionName: regionName,
		RegionCode: regionCode,
		Offers:     offers,
	}, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

// scanOfferWithPlate - разбирает одну строку выборки offers JOIN plates в домен
func scanOfferWithPlate(row rowScanner) (domain.OfferWithPlate, error) {
	var (
		offerID         uuid.UUID
		provider        string
		externalID      string
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
		plateID         uuid.UUID
		number          string
		vehicleType     string
		plateCreatedAt  *time.Time
		plateUpdatedAt  *time.Time
	)

	err := row.Scan(&offerID, &provider, &externalID, &price, &status, &whereabouts, &reissueIncluded, &viewCount, &postedAt,
		&refreshedAt, &url, &raw, &rawDetailed, &comment, &offerCreatedAt, &offerUpdatedAt, &plateID, &number, &vehicleType, &plateCreatedAt, &plateUpdatedAt)
	if err != nil {
		return domain.OfferWithPlate{}, err
	}

	plate, err := domain.RestorePlate(plateID, number, domain.PlateType(vehicleType), plateCreatedAt, plateUpdatedAt)
	if err != nil {
		return domain.OfferWithPlate{}, fmt.Errorf("restore plate %s: %w", plateID, err)
	}

	var whereaboutsVO *domain.OfferWhereabouts
	if whereabouts != nil {
		vo := domain.OfferWhereabouts(*whereabouts)
		whereaboutsVO = &vo
	}

	offer, err := domain.RestoreOffer(
		offerID, plateID, domain.Provider(provider), externalID, price, domain.OfferStatus(status), whereaboutsVO, reissueIncluded,
		viewCount, postedAt, refreshedAt, url, raw, rawDetailed, comment, offerCreatedAt, offerUpdatedAt,
	)
	if err != nil {
		return domain.OfferWithPlate{}, fmt.Errorf("restore offer %s: %w", offerID, err)
	}

	return domain.OfferWithPlate{Plate: plate, Offer: offer}, nil
}
