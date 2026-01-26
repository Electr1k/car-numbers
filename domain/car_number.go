package domain

import "time"

type CarNumber struct {
	Number   string    `validate:"required,min=8,max=9"`
	Price    float32   `validate:"required,min=0"`
	PostedAt time.Time `validate:"required"`
}
