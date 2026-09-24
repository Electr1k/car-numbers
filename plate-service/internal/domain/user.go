package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrUserNotFound - Пользователь не найден
var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID        uuid.UUID `validate:"required"` // ID - Идентификатор
	Name      string    `validate:"required"` // Name - Имя
	Phone     *string   // Phone - Номер
	Email     *string   // Email - E-mail
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

// NewUser - Создание пользователя, которого ещё не существовало
func NewUser(
	name string,
	phone *string,
	email *string,
) (*User, error) {
	id, err := newID()
	if err != nil {
		return nil, err
	}

	return newUser(
		id,
		name,
		phone,
		email,
		nil,
		nil,
	)
}

// RestoreUser - Восстановление существующего пользователя из хранилища
func RestoreUser(
	id uuid.UUID,
	name string,
	phone *string,
	email *string,
	createdAt *time.Time,
	updatedAt *time.Time,
) (*User, error) {
	return newUser(
		id,
		name,
		phone,
		email,
		createdAt,
		updatedAt,
	)
}

func newUser(
	id uuid.UUID,
	name string,
	phone *string,
	email *string,
	createdAt *time.Time,
	updatedAt *time.Time,
) (*User, error) {
	user := &User{
		ID:        id,
		Name:      name,
		Phone:     phone,
		Email:     email,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	if err := validate.Struct(user); err != nil {
		return nil, err
	}

	return user, nil
}
