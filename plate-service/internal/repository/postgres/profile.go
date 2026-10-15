package postgres

import (
	"context"
	"errors"
	"fmt"
	"plate-service/internal/domain"
	"plate-service/internal/domain/data"
	"plate-service/internal/repository"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

// ProfileRepository - хранение профилей продавцов
type ProfileRepository struct {
	postgres *Postgres
}

func NewProfileRepository(postgres *Postgres) *ProfileRepository {
	return &ProfileRepository{postgres: postgres}
}

// upsertProfileQuery - вставка профиля или обновление по (provider, external_id), id существующего профиля сохраняется
const upsertProfileQuery = `
INSERT INTO profiles (id, provider, external_id, url, icon_url, login, name, phone, email,
	registered_at, last_visit_at, raw, description)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
ON CONFLICT (provider, external_id) DO UPDATE SET
	url           = EXCLUDED.url,
	icon_url      = EXCLUDED.icon_url,
	login         = EXCLUDED.login,
	name          = EXCLUDED.name,
	phone         = EXCLUDED.phone,
	email         = EXCLUDED.email,
	registered_at = EXCLUDED.registered_at,
	last_visit_at = EXCLUDED.last_visit_at,
	raw           = EXCLUDED.raw,
	description   = EXCLUDED.description,
	updated_at    = CURRENT_TIMESTAMP
RETURNING id, provider, external_id, url, icon_url, login, name, phone, email,
	registered_at, last_visit_at, raw, description, created_at, updated_at`

// UpdateOrCreateProfile - Сохранение профиля из свежей деталки поверх сохранённого
func (r *ProfileRepository) UpdateOrCreateProfile(ctx context.Context, profile *domain.Profile) (*domain.Profile, error) {
	row := r.postgres.pool.QueryRow(ctx, upsertProfileQuery,
		profile.ID, profile.Provider, profile.ExternalID, profile.URL, profile.IconURL, profile.Login, profile.Name, profile.Phone,
		profile.Email, profile.RegisteredAt, profile.LastVisitAt, profile.Raw, profile.Description,
	)

	var (
		id           uuid.UUID
		provider     string
		externalID   string
		url          string
		iconURL      *string
		login        string
		name         *string
		phone        *string
		email        *string
		registeredAt *time.Time
		lastVisitAt  *time.Time
		raw          string
		description  *string
		createdAt    *time.Time
		updatedAt    *time.Time
	)

	err := row.Scan(&id, &provider, &externalID, &url, &iconURL, &login, &name, &phone, &email,
		&registeredAt, &lastVisitAt, &raw, &description, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("save profile %s/%s: %w", profile.Provider, profile.ExternalID, err)
	}

	stored, err := domain.RestoreProfile(
		id, domain.Provider(provider), externalID, url, iconURL, login, name, phone, email,
		registeredAt, lastVisitAt, raw, description, createdAt, updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("restore profile %s: %w", id, err)
	}

	return stored, nil
}

// relatedProfilesLimit - сколько связанных аккаунтов отдаём в профиле
const relatedProfilesLimit = 50

// getProfileDetailQuery - профиль продавца с количеством его предложений
const getProfileDetailQuery = `
SELECT p.id, p.provider, p.url, p.icon_url, p.login, p.name, p.phone, p.email,
	count(o.id) FILTER (WHERE o.status = $2),
	count(o.id)
FROM profiles p
LEFT JOIN offers o ON o.profile_id = p.id
WHERE p.id = $1
GROUP BY p.id`

// getRelatedProfilesQuery - другие профили с тем же телефоном, email или логином
const getRelatedProfilesQuery = `
SELECT p.id, p.provider, p.url, p.icon_url, p.login, p.name,
        coalesce(p.phone = me.phone, false),
        coalesce(lower(p.email) = lower(me.email), false),
        lower(p.login) = lower(me.login),
        count(o.id) FILTER (WHERE o.status = $2) as active,
        count(o.id) as total
FROM profiles me
JOIN profiles p ON p.id <> me.id
        AND (p.phone = me.phone OR lower(p.email) = lower(me.email) OR lower(p.login) = lower(me.login))
LEFT JOIN offers o ON o.profile_id = p.id
WHERE me.id = $1
GROUP BY p.id, me.id
ORDER BY active DESC, total DESC, p.provider, p.login`

// GetProfileDetail - Профиль продавца
func (r *ProfileRepository) GetProfileDetail(ctx context.Context, id uuid.UUID) (*data.ProfileDetail, error) {
	var (
		profile  data.ProfileDetail
		provider string
	)

	err := r.postgres.pool.QueryRow(ctx, getProfileDetailQuery, id, domain.OfferStatusActive).Scan(
		&profile.ID, &provider, &profile.URL, &profile.IconURL, &profile.Login, &profile.Name, &profile.Phone, &profile.Email,
		&profile.Offers.Active, &profile.Offers.Total,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrProfileNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get profile %s: %w", id, err)
	}

	profile.Provider = domain.Provider(provider)

	return &profile, nil
}

// GetRelatedProfiles - Вероятные другие профили продавца
func (r *ProfileRepository) GetRelatedProfiles(ctx context.Context, id uuid.UUID) ([]data.RelatedProfile, error) {
	rows, err := r.postgres.pool.Query(ctx, getRelatedProfilesQuery, id, domain.OfferStatusActive)
	if err != nil {
		return nil, fmt.Errorf("get related profiles %s: %w", id, err)
	}
	defer rows.Close()

	related := make([]data.RelatedProfile, 0)
	for rows.Next() {
		var (
			profile                            data.RelatedProfile
			provider                           string
			phoneMatch, emailMatch, loginMatch bool
		)

		if err := rows.Scan(&profile.ID, &provider, &profile.URL, &profile.IconURL, &profile.Login, &profile.Name,
			&phoneMatch, &emailMatch, &loginMatch, &profile.Offers.Active, &profile.Offers.Total); err != nil {
			return nil, fmt.Errorf("get related profiles %s: %w", id, err)
		}

		profile.Provider = domain.Provider(provider)
		if phoneMatch {
			profile.Matches = append(profile.Matches, data.ProfileMatchPhone)
		}
		if emailMatch {
			profile.Matches = append(profile.Matches, data.ProfileMatchEmail)
		}
		if loginMatch {
			profile.Matches = append(profile.Matches, data.ProfileMatchLogin)
		}

		related = append(related, profile)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get related profiles %s: %w", id, err)
	}

	return related, nil
}

// getProfileOffersQuery - предложения продавца, свежие сверху; $2 - статусы, NULL - все
const getProfileOffersQuery = `
SELECT o.id, o.plate_id, pl.number, o.price, o.status, o.posted_at, o.refreshed_at, o.url
FROM offers o
JOIN plates pl ON pl.id = o.plate_id
WHERE o.profile_id = $1
	AND ($2::text[] IS NULL OR o.status = ANY($2))
	AND ($4::uuid IS NULL OR (o.refreshed_at, o.id) < ($5::timestamptz, $4::uuid))
ORDER BY o.refreshed_at DESC, o.id DESC
LIMIT $3`

// GetProfileOffers - Предложения продавца с заданными статусами, после курсора
func (r *ProfileRepository) GetProfileOffers(ctx context.Context, params repository.GetProfileOffersParams) ([]data.ProfileOffer, bool, error) {
	var statuses any
	if len(params.Statuses) > 0 {
		values := make([]string, 0, len(params.Statuses))
		for _, status := range params.Statuses {
			values = append(values, string(status))
		}
		statuses = values
	}

	var cursorID, cursorRefreshedAt any
	if params.Cursor != nil {
		cursorID, cursorRefreshedAt = params.Cursor.ID, params.Cursor.RefreshedAt
	}

	// Лишняя строка показывает, есть ли следующая страница
	rows, err := r.postgres.pool.Query(ctx, getProfileOffersQuery,
		params.ProfileID, statuses, params.Limit+1, cursorID, cursorRefreshedAt)
	if err != nil {
		return nil, false, fmt.Errorf("get profile offers %s: %w", params.ProfileID, err)
	}
	defer rows.Close()

	offers := make([]data.ProfileOffer, 0, params.Limit+1)
	for rows.Next() {
		var (
			offer  data.ProfileOffer
			status string
		)

		if err := rows.Scan(&offer.ID, &offer.PlateID, &offer.Number, &offer.Price, &status,
			&offer.PostedAt, &offer.RefreshedAt, &offer.URL); err != nil {
			return nil, false, fmt.Errorf("get profile offers %s: %w", params.ProfileID, err)
		}

		offer.Status = domain.OfferStatus(status)
		offers = append(offers, offer)
	}

	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("get profile offers %s: %w", params.ProfileID, err)
	}

	hasNext := len(offers) > params.Limit
	if hasNext {
		offers = offers[:params.Limit]
	}

	return offers, hasNext, nil
}
