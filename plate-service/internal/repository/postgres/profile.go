package postgres

import (
	"context"
	"errors"
	"fmt"
	"plate-service/internal/domain"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

// ProfileRepository - хранение профилей провайдеров и доменных пользователей
type ProfileRepository struct {
	postgres *Postgres
}

func NewProfileRepository(postgres *Postgres) *ProfileRepository {
	return &ProfileRepository{postgres: postgres}
}

// getUserByContactsQuery - пользователь по телефону или email
const getUserByContactsQuery = `
SELECT id, name, phone, email, created_at, updated_at
FROM users
WHERE phone = $1 OR email = $2
ORDER BY (phone = $1) IS TRUE DESC, created_at
LIMIT 1;`

const insertUserQuery = `
INSERT INTO users (id, name, phone, email)
VALUES ($1, $2, $3, $4)
RETURNING id, name, phone, email, created_at, updated_at;`

const upsertProfileQuery = `
INSERT INTO profiles (id, user_id, provider, external_id, url, login, name, phone, email, badge, rating,
	registered_at, last_visit_at, raw)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
ON CONFLICT (provider, external_id) DO UPDATE SET
	user_id       = EXCLUDED.user_id,
	url           = EXCLUDED.url,
	login         = EXCLUDED.login,
	name          = EXCLUDED.name,
	phone         = EXCLUDED.phone,
	email         = EXCLUDED.email,
	badge         = EXCLUDED.badge,
	rating        = EXCLUDED.rating,
	registered_at = EXCLUDED.registered_at,
	last_visit_at = EXCLUDED.last_visit_at,
	raw           = EXCLUDED.raw,
	updated_at    = CURRENT_TIMESTAMP
RETURNING id, user_id, provider, external_id, url, login, name, phone, email, badge, rating,
	registered_at, last_visit_at, raw, created_at, updated_at;`

const getExistingProfileExternalIDsQuery = `
SELECT external_id
FROM profiles
WHERE provider = $1 AND external_id = ANY($2);`

// GetExistingProfileExternalIDs - Внешние идентификаторы профилей, которые уже есть в базе
func (r *ProfileRepository) GetExistingProfileExternalIDs(
	ctx context.Context,
	provider domain.Provider,
	externalIDs []string,
) ([]string, error) {
	rows, err := r.postgres.pool.Query(ctx, getExistingProfileExternalIDsQuery, provider, externalIDs)
	if err != nil {
		return nil, fmt.Errorf("get existing profiles: %w", err)
	}
	defer rows.Close()

	existing := make([]string, 0, len(externalIDs))
	for rows.Next() {
		var externalID string
		if err := rows.Scan(&externalID); err != nil {
			return nil, fmt.Errorf("scan existing profile: %w", err)
		}

		existing = append(existing, externalID)
	}

	return existing, rows.Err()
}

// GetUserByContacts - Пользователь по телефону или email
func (r *ProfileRepository) GetUserByContacts(ctx context.Context, phone *string, email *string) (*domain.User, error) {
	if phone == nil && email == nil {
		return nil, domain.ErrUserNotFound
	}

	user, err := scanUser(r.postgres.pool.QueryRow(ctx, getUserByContactsQuery, phone, email))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by contacts: %w", err)
	}

	return user, nil
}

// CreateUser - Создание пользователя
func (r *ProfileRepository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	saved, err := scanUser(r.postgres.pool.QueryRow(ctx, insertUserQuery, user.ID, user.Name, user.Phone, user.Email))
	if err != nil {
		return nil, fmt.Errorf("create user %s: %w", user.ID, err)
	}

	return saved, nil
}

// UpsertProfile - Сохранение профиля по (provider, external_id)
func (r *ProfileRepository) UpsertProfile(ctx context.Context, profile *domain.Profile) (*domain.Profile, error) {
	row := r.postgres.pool.QueryRow(ctx, upsertProfileQuery,
		profile.ID,
		profile.UserID,
		profile.Provider,
		profile.ExternalID,
		profile.URL,
		profile.Login,
		profile.Name,
		profile.Phone,
		profile.Email,
		profile.Badge,
		profile.Rating,
		profile.RegisteredAt,
		profile.LastVisitAt,
		profile.Raw,
	)

	saved, err := scanProfile(row)
	if err != nil {
		return nil, fmt.Errorf("save profile %s/%s: %w", profile.Provider, profile.ExternalID, err)
	}

	return saved, nil
}

// scanUser - разбирает одну строку users в домен
func scanUser(row rowScanner) (*domain.User, error) {
	var (
		id        uuid.UUID
		name      string
		phone     *string
		email     *string
		createdAt *time.Time
		updatedAt *time.Time
	)

	if err := row.Scan(&id, &name, &phone, &email, &createdAt, &updatedAt); err != nil {
		return nil, err
	}

	user, err := domain.RestoreUser(id, name, phone, email, createdAt, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("restore user %s: %w", id, err)
	}

	return user, nil
}

// scanProfile - разбирает одну строку profiles в домен
func scanProfile(row rowScanner) (*domain.Profile, error) {
	var (
		id           uuid.UUID
		userID       *uuid.UUID
		provider     string
		externalID   string
		url          string
		login        *string
		name         string
		phone        *string
		email        *string
		badge        *string
		rating       *int
		registeredAt *time.Time
		lastVisitAt  *time.Time
		raw          string
		createdAt    *time.Time
		updatedAt    *time.Time
	)

	err := row.Scan(&id, &userID, &provider, &externalID, &url, &login, &name, &phone, &email, &badge, &rating,
		&registeredAt, &lastVisitAt, &raw, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	profile, err := domain.RestoreProfile(
		id, userID, domain.Provider(provider), externalID, url, login, name, phone, email, badge, rating,
		registeredAt, lastVisitAt, raw, createdAt, updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("restore profile %s: %w", id, err)
	}

	return profile, nil
}
