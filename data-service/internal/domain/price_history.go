package domain

import (
	"time"

	"github.com/google/uuid"
)

// PriceHistory - История цены продажи предложения
type PriceHistory struct {
	ID        uuid.UUID `validate:"required"`       // ID - Идентификатор
	OfferID   uuid.UUID `validate:"required"`       // OfferID - Идентификатор предложения
	PlateID   uuid.UUID `validate:"required"`       // PlateID - Идентификатор номера
	Price     *float64  `validate:"omitempty,gt=0"` // Price - Цена
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

func NewPriceHistory(offerID uuid.UUID, plateID uuid.UUID, price *float64) (*PriceHistory, error) {
	id, err := newID()
	if err != nil {
		return nil, err
	}

	return newPriceHistory(id, offerID, plateID, price, nil, nil)
}

func NewPriceHistoryFromOffer(offer *Offer) (*PriceHistory, error) {
	return NewPriceHistory(offer.ID, offer.PlateID, offer.Price)
}

func RestorePriceHistory(
	id uuid.UUID,
	offerID uuid.UUID,
	plateID uuid.UUID,
	price *float64,
	createdAt *time.Time,
	updatedAt *time.Time,
) (*PriceHistory, error) {
	return newPriceHistory(id, offerID, plateID, price, createdAt, updatedAt)
}

func newPriceHistory(
	id uuid.UUID,
	offerID uuid.UUID,
	plateID uuid.UUID,
	price *float64,
	createdAt *time.Time,
	updatedAt *time.Time,
) (*PriceHistory, error) {
	p := &PriceHistory{
		ID:        id,
		OfferID:   offerID,
		PlateID:   plateID,
		Price:     price,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	if err := validate.Struct(p); err != nil {
		return nil, err
	}

	return p, nil
}
