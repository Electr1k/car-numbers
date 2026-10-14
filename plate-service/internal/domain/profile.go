package domain

import (
	"errors"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

// ErrProfileNotFound - Профиль не найден
var ErrProfileNotFound = errors.New("profile not found")

// Profile - Профиль продавца у провайдера
type Profile struct {
	ID           uuid.UUID  `validate:"required"`                                    // ID - Идентификатор
	Provider     Provider   `validate:"required,oneof=autonomera gosnomeru anomera"` // Provider - Провайдер, в котором найден профиль
	ExternalID   string     `validate:"required"`                                    // ExternalID - Идентификатор у провайдера
	URL          string     `validate:"required,max=255"`                            // URL - Ссылка на профиль у провайдера
	Login        string     `validate:"required"`                                    // Login - Логин
	Name         *string    `validate:"omitempty"`                                   // Name - Имя
	Phone        *string    `validate:"omitempty,len=11,numeric,startswith=7"`       // Phone - Номер в виде 79999999999
	Email        *string    `validate:"omitempty"`                                   // Email - Email
	RegisteredAt *time.Time `validate:"omitempty"`                                   // RegisteredAt - Дата регистрации
	LastVisitAt  *time.Time `validate:"omitempty"`                                   // LastVisitAt - Дата последнего визита
	Raw          string     `validate:"required"`                                    // Raw - Сырой объект поставщика
	Description  *string    `validate:"omitempty"`                                   // Description - Описание профиля
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}

// NormalizePhone - Нормализация номера
func NormalizePhone(raw string) *string {
	digits := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, raw)

	switch {
	case len(digits) == 11 && (digits[0] == '7' || digits[0] == '8'):
		digits = "7" + digits[1:]
	case len(digits) == 10 && digits[0] == '9':
		digits = "7" + digits
	default:
		return nil
	}

	return &digits
}

// NewProfile - Создание профиля, которого ещё не существовало
func NewProfile(
	provider Provider,
	externalID string,
	url string,
	login string,
	name *string,
	phone *string,
	email *string,
	registeredAt *time.Time,
	lastVisitAt *time.Time,
	raw string,
	description *string,
) (*Profile, error) {
	id, err := newID()
	if err != nil {
		return nil, err
	}

	return newProfile(
		id,
		provider,
		externalID,
		url,
		login,
		name,
		phone,
		email,
		registeredAt,
		lastVisitAt,
		raw,
		description,
		nil,
		nil,
	)
}

// RestoreProfile - Восстановление существующего профиля из хранилища
func RestoreProfile(
	id uuid.UUID,
	provider Provider,
	externalID string,
	url string,
	login string,
	name *string,
	phone *string,
	email *string,
	registeredAt *time.Time,
	lastVisitAt *time.Time,
	raw string,
	description *string,
	createdAt *time.Time,
	updatedAt *time.Time,
) (*Profile, error) {
	return newProfile(
		id,
		provider,
		externalID,
		url,
		login,
		name,
		phone,
		email,
		registeredAt,
		lastVisitAt,
		raw,
		description,
		createdAt,
		updatedAt,
	)
}

func newProfile(
	id uuid.UUID,
	provider Provider,
	externalID string,
	url string,
	login string,
	name *string,
	phone *string,
	email *string,
	registeredAt *time.Time,
	lastVisitAt *time.Time,
	raw string,
	description *string,
	createdAt *time.Time,
	updatedAt *time.Time,
) (*Profile, error) {
	profile := &Profile{
		ID:           id,
		Provider:     provider,
		ExternalID:   externalID,
		URL:          url,
		Login:        login,
		Name:         name,
		Phone:        phone,
		Email:        email,
		RegisteredAt: registeredAt,
		LastVisitAt:  lastVisitAt,
		Raw:          raw,
		Description:  description,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}

	if err := validate.Struct(profile); err != nil {
		return nil, err
	}

	return profile, nil
}
