package domain

import (
	"time"

	"github.com/google/uuid"
)

// NumberType - тип ТС, которому принадлежит номер
type NumberType string

const (
	// NumberTypeCar - Авто
	NumberTypeCar NumberType = "car"

	// NumberTypeMoto - Мото
	NumberTypeMoto NumberType = "moto"

	// NumberTypeTrailer - Прицеп
	NumberTypeTrailer NumberType = "trailer"
)

// Number - Номер
type Number struct {
	Id        uuid.UUID  `validate:"required"`                        // Id - Идентификатор
	Number    string     `validate:"required,min=8,max=9"`            // Number - Номер
	Type      NumberType `validate:"required,oneof=car moto trailer"` // Type - Тип номера
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

// NewNumber - Создание номера, которого ещё не существовало
func NewNumber(number string, numberType NumberType) (*Number, error) {
	id, err := newID()
	if err != nil {
		return nil, err
	}

	return newNumber(id, number, numberType, nil, nil)
}

// RestoreNumber - Восстановление существующего номера из хранилища
func RestoreNumber(
	id uuid.UUID,
	number string,
	numberType NumberType,
	createdAt *time.Time,
	updatedAt *time.Time,
) (*Number, error) {
	return newNumber(id, number, numberType, createdAt, updatedAt)
}

func newNumber(
	id uuid.UUID,
	number string,
	numberType NumberType,
	createdAt *time.Time,
	updatedAt *time.Time,
) (*Number, error) {
	n := &Number{
		Id:        id,
		Number:    number,
		Type:      numberType,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	if err := validate.Struct(n); err != nil {
		return nil, err
	}

	return n, nil
}
