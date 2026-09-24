package domain

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	ID           uuid.UUID  `validate:"required"` // ID - Идентификатор
	UserID       *uuid.UUID // UserID - Идентификатор пользователя
	Provider     Provider   `validate:"required,oneof=autonomera gosnomeru anomera"` // Provider - Провайдер, в котором найден профиль
	ExternalID   string     `validate:"required"`                                    // ExternalID - Идентификатор у провайдера
	URL          string     `validate:"required"`                                    // URL - Ссылка на профиль
	Login        *string    // Login - Логин
	Name         string     `validate:"required"` // Name - Имя
	Phone        *string    // Phone - Телефон
	Email        *string    // Email - Почта
	Badge        *string    // Badge - Статус у провайдера
	Rating       *int       // Rating - Рейтинг у провайдера
	RegisteredAt *time.Time `validate:"required"` // RegisteredAt - Дата регистрации у провайдера
	LastVisitAt  *time.Time `validate:"required"` // LastVisitAt - Дата последней активности
	Raw          string     `validate:"required"` // Raw - Сырой объект поставщика
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}

// NewProfile - Создание профиля, которого ещё не существовало
func NewProfile(
	userID *uuid.UUID,
	provider Provider,
	externalID string,
	url string,
	login *string,
	name string,
	phone *string,
	email *string,
	badge *string,
	rating *int,
	registeredAt *time.Time,
	lastVisitAt *time.Time,
	raw string,
) (*Profile, error) {
	id, err := newID()
	if err != nil {
		return nil, err
	}

	return newProfile(
		id,
		userID,
		provider,
		externalID,
		url,
		login,
		name,
		phone,
		email,
		badge,
		rating,
		registeredAt,
		lastVisitAt,
		raw,
		nil,
		nil,
	)
}

// RestoreProfile - Восстановление существующего профиля из хранилища
func RestoreProfile(
	id uuid.UUID,
	userID *uuid.UUID,
	provider Provider,
	externalID string,
	url string,
	login *string,
	name string,
	phone *string,
	email *string,
	badge *string,
	rating *int,
	registeredAt *time.Time,
	lastVisitAt *time.Time,
	raw string,
	createdAt *time.Time,
	updatedAt *time.Time,
) (*Profile, error) {
	return newProfile(
		id,
		userID,
		provider,
		externalID,
		url,
		login,
		name,
		phone,
		email,
		badge,
		rating,
		registeredAt,
		lastVisitAt,
		raw,
		createdAt,
		updatedAt,
	)
}

func newProfile(
	id uuid.UUID,
	userID *uuid.UUID,
	provider Provider,
	externalID string,
	url string,
	login *string,
	name string,
	phone *string,
	email *string,
	badge *string,
	rating *int,
	registeredAt *time.Time,
	lastVisitAt *time.Time,
	raw string,
	createdAt *time.Time,
	updatedAt *time.Time,
) (*Profile, error) {
	profile := &Profile{
		ID:           id,
		UserID:       userID,
		Provider:     provider,
		ExternalID:   externalID,
		URL:          url,
		Login:        login,
		Name:         name,
		Phone:        phone,
		Email:        email,
		Badge:        badge,
		Rating:       rating,
		RegisteredAt: registeredAt,
		LastVisitAt:  lastVisitAt,
		Raw:          raw,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}

	if err := validate.Struct(profile); err != nil {
		return nil, err
	}

	return profile, nil
}
