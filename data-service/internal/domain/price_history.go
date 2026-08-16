package domain

import (
	"time"

	"github.com/google/uuid"
)

// PriceHistory - История цены продажи предложения
type PriceHistory struct {
	Id        uuid.UUID `validate:"required"`       // Id - Идентификатор
	OfferId   uuid.UUID `validate:"required"`       // OfferId - Идентификатор предложения
	NumberId  uuid.UUID `validate:"required"`       // NumberId - Идентификатор номера
	Price     *float64  `validate:"omitempty,gt=0"` // Price - Цена
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

func NewPriceHistory(offerId uuid.UUID, numberId uuid.UUID, price *float64) (*PriceHistory, error) {
	id, err := newID()
	if err != nil {
		return nil, err
	}

	return newPriceHistory(id, offerId, numberId, price, nil, nil)
}

func NewPriceHistoryFromOffer(offer *Offer) (*PriceHistory, error) {
	return NewPriceHistory(offer.Id, offer.NumberId, offer.Price)
}

func RestorePriceHistory(
	id uuid.UUID,
	offerId uuid.UUID,
	numberId uuid.UUID,
	price *float64,
	createdAt *time.Time,
	updatedAt *time.Time,
) (*PriceHistory, error) {
	return newPriceHistory(id, offerId, numberId, price, createdAt, updatedAt)
}

func newPriceHistory(
	id uuid.UUID,
	offerId uuid.UUID,
	numberId uuid.UUID,
	price *float64,
	createdAt *time.Time,
	updatedAt *time.Time,
) (*PriceHistory, error) {
	p := &PriceHistory{
		Id:        id,
		OfferId:   offerId,
		NumberId:  numberId,
		Price:     price,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	if err := validate.Struct(p); err != nil {
		return nil, err
	}

	return p, nil
}
