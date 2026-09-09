package domain

import (
	"time"

	"github.com/google/uuid"
)

// PriceHistory - История цены продажи предложения
type PriceHistory struct {
	ID        uuid.UUID `validate:"required"`       // ID - Идентификатор
	OfferID   uuid.UUID `validate:"required"`       // OfferID - Идентификатор предложения
	NumberID  uuid.UUID `validate:"required"`       // NumberID - Идентификатор номера
	Price     *float64  `validate:"omitempty,gt=0"` // Price - Цена
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

func NewPriceHistory(offerID uuid.UUID, numberID uuid.UUID, price *float64) (*PriceHistory, error) {
	id, err := newID()
	if err != nil {
		return nil, err
	}

	return newPriceHistory(id, offerID, numberID, price, nil, nil)
}

func NewPriceHistoryFromOffer(offer *Offer) (*PriceHistory, error) {
	return NewPriceHistory(offer.ID, offer.NumberID, offer.Price)
}

func RestorePriceHistory(
	id uuid.UUID,
	offerID uuid.UUID,
	numberID uuid.UUID,
	price *float64,
	createdAt *time.Time,
	updatedAt *time.Time,
) (*PriceHistory, error) {
	return newPriceHistory(id, offerID, numberID, price, createdAt, updatedAt)
}

func newPriceHistory(
	id uuid.UUID,
	offerID uuid.UUID,
	numberID uuid.UUID,
	price *float64,
	createdAt *time.Time,
	updatedAt *time.Time,
) (*PriceHistory, error) {
	p := &PriceHistory{
		ID:        id,
		OfferID:   offerID,
		NumberID:  numberID,
		Price:     price,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	if err := validate.Struct(p); err != nil {
		return nil, err
	}

	return p, nil
}
