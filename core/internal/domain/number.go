package domain

import (
	"time"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

// Number - Номер
type Number struct {
	Id        *int64     `validate:"omitempty"`            // Id - Идентификатор
	Number    string     `validate:"required,min=8,max=9"` // Number - Номер
	Type      string     `validate:"required"`             // Type - Тип номера (авто, мото, прицеп)
	CreatedAt *time.Time `validate:""`
	UpdateAt  *time.Time `validate:""`
}

func NewNumber(
	id *int64,
	number string,
	numberType string,
	createdAt *time.Time,
	updatedAt *time.Time,
) (*Number, error) {
	n := &Number{
		Id:        id,
		Number:    number,
		Type:      numberType,
		CreatedAt: createdAt,
		UpdateAt:  updatedAt,
	}

	if err := n.Validate(); err != nil {
		return nil, err
	}

	return n, nil
}

func (p *Number) Validate() error {
	return validate.Struct(p)
}
