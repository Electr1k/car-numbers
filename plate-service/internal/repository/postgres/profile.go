package postgres

import (
	"context"
	"fmt"
	"plate-service/internal/domain"
	"time"

	"github.com/google/uuid"
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
INSERT INTO profiles (id, provider, external_id, url, login, name, phone, email,
	registered_at, last_visit_at, raw, description)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
ON CONFLICT (provider, external_id) DO UPDATE SET
	url           = EXCLUDED.url,
	login         = EXCLUDED.login,
	name          = EXCLUDED.name,
	phone         = EXCLUDED.phone,
	email         = EXCLUDED.email,
	registered_at = EXCLUDED.registered_at,
	last_visit_at = EXCLUDED.last_visit_at,
	raw           = EXCLUDED.raw,
	description   = EXCLUDED.description,
	updated_at    = CURRENT_TIMESTAMP
RETURNING id, provider, external_id, url, login, name, phone, email,
	registered_at, last_visit_at, raw, description, created_at, updated_at`

// UpdateOrCreateProfile - Сохранение профиля из свежей деталки поверх сохранённого
func (r *ProfileRepository) UpdateOrCreateProfile(ctx context.Context, profile *domain.Profile) (*domain.Profile, error) {
	row := r.postgres.pool.QueryRow(ctx, upsertProfileQuery,
		profile.ID, profile.Provider, profile.ExternalID, profile.URL, profile.Login, profile.Name, profile.Phone,
		profile.Email, profile.RegisteredAt, profile.LastVisitAt, profile.Raw, profile.Description,
	)

	var (
		id           uuid.UUID
		provider     string
		externalID   string
		url          string
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

	err := row.Scan(&id, &provider, &externalID, &url, &login, &name, &phone, &email,
		&registeredAt, &lastVisitAt, &raw, &description, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("save profile %s/%s: %w", profile.Provider, profile.ExternalID, err)
	}

	stored, err := domain.RestoreProfile(
		id, domain.Provider(provider), externalID, url, login, name, phone, email,
		registeredAt, lastVisitAt, raw, description, createdAt, updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("restore profile %s: %w", id, err)
	}

	return stored, nil
}
