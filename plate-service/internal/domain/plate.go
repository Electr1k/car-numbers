package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// PlateType - тип ТС, которому принадлежит номер
type PlateType string

const (
	// PlateTypeCar - Авто
	PlateTypeCar PlateType = "car"

	// PlateTypeMoto - Мото
	PlateTypeMoto PlateType = "moto"

	// PlateTypeTrailer - Прицеп
	PlateTypeTrailer PlateType = "trailer"
)

var ErrPlateNotFound = errors.New("plate not found")

// Plate - Номер
type Plate struct {
	ID        uuid.UUID `validate:"required"`                        // ID - Идентификатор
	Number    string    `validate:"required,min=8,max=9"`            // Number - Номер
	Type      PlateType `validate:"required,oneof=car moto trailer"` // Type - Тип номера
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

// Categories - Классифицирует номер по категорями
func (p *Plate) Categories() []CategoryID {
	return classifyPlate(p.Number, p.Type)
}

// NewPlate - Создание номера, которого ещё не существовало
func NewPlate(number string, plateType PlateType) (*Plate, error) {
	id, err := newID()
	if err != nil {
		return nil, err
	}

	return newPlate(id, number, plateType, nil, nil)
}

// RestorePlate - Восстановление существующего номера из хранилища
func RestorePlate(
	id uuid.UUID,
	number string,
	plateType PlateType,
	createdAt *time.Time,
	updatedAt *time.Time,
) (*Plate, error) {
	return newPlate(id, number, plateType, createdAt, updatedAt)
}

func newPlate(
	id uuid.UUID,
	number string,
	plateType PlateType,
	createdAt *time.Time,
	updatedAt *time.Time,
) (*Plate, error) {
	n := &Plate{
		ID:        id,
		Number:    number,
		Type:      plateType,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	if err := validate.Struct(n); err != nil {
		return nil, err
	}

	return n, nil
}
