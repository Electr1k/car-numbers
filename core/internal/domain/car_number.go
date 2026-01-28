package domain

import (
	"time"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

type CarNumber struct {
	Number   string    `validate:"required,min=8,max=9"`
	Price    float32   `validate:"required,min=0"`
	PostedAt time.Time `validate:"required"`
}

func NewCarNumber(
	number string,
	price float32,
	postedAt time.Time,
) (*CarNumber, error) {
	n := &CarNumber{
		Number:   number,
		Price:    price,
		PostedAt: postedAt,
	}

	if err := n.Validate(); err != nil {
		return nil, err
	}

	return n, nil
}

func (p *CarNumber) Validate() error {
	return validate.Struct(p)
}
