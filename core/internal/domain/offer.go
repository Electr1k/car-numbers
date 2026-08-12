package domain

import (
	"time"

	"github.com/google/uuid"
)

// OfferStatus - состояние предложения у провайдера
type OfferStatus string

const (
	// OfferStatusActive - Активно
	OfferStatusActive OfferStatus = "active"

	// OfferStatusSold - Продано
	OfferStatusSold OfferStatus = "sold"

	// OfferStatusInactive - Неактивно
	OfferStatusInactive OfferStatus = "inactive"
)

// Offer - Предложение
type Offer struct {
	Id         uuid.UUID   `validate:"required"`                            // Id - Идентификатор
	NumberId   uuid.UUID   `validate:"required"`                            // NumberId - Идентификатор номера
	Provider   Provider    `validate:"required,oneof=autonomera"`           // Provider - Провайдер, в котором найдено предложение
	ExternalId string      `validate:"required"`                            // ExternalId - Идентификатор у провайдера
	Price      float32     `validate:"required,gt=0"`                       // Price - Цена
	Status     OfferStatus `validate:"required,oneof=active sold inactive"` // Status - Статус предложения
	PostedAt   *time.Time  `validate:"required"`                            // PostedAt - Дата публикации
	Url        string      `validate:"required"`                            // Url - Ссылка на предложение
	Raw        string      `validate:"required"`                            // Raw - Сырой объект поставщика
	CreatedAt  *time.Time
	UpdatedAt  *time.Time
}

// NewOffer - Создание предложения, которого ещё не существовало
func NewOffer(
	numberId uuid.UUID,
	provider Provider,
	externalId string,
	price float32,
	status OfferStatus,
	postedAt *time.Time,
	url string,
	raw string,
) (*Offer, error) {
	id, err := newID()
	if err != nil {
		return nil, err
	}

	return newOffer(id, numberId, provider, externalId, price, status, postedAt, url, raw, nil, nil)
}

// RestoreOffer - Восстановление существующего предложения из хранилища
func RestoreOffer(
	id uuid.UUID,
	numberId uuid.UUID,
	provider Provider,
	externalId string,
	price float32,
	status OfferStatus,
	postedAt *time.Time,
	url string,
	raw string,
	createdAt *time.Time,
	updatedAt *time.Time,
) (*Offer, error) {
	return newOffer(id, numberId, provider, externalId, price, status, postedAt, url, raw, createdAt, updatedAt)
}

func newOffer(
	id uuid.UUID,
	numberId uuid.UUID,
	provider Provider,
	externalId string,
	price float32,
	status OfferStatus,
	postedAt *time.Time,
	url string,
	raw string,
	createdAt *time.Time,
	updatedAt *time.Time,
) (*Offer, error) {
	offer := &Offer{
		Id:         id,
		NumberId:   numberId,
		Provider:   provider,
		ExternalId: externalId,
		Price:      price,
		Status:     status,
		PostedAt:   postedAt,
		Url:        url,
		Raw:        raw,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}

	if err := validate.Struct(offer); err != nil {
		return nil, err
	}

	return offer, nil
}

// OfferWithNumber - Предложение вместе с номером, к которому оно относится
type OfferWithNumber struct {
	Number *Number
	Offer  *Offer
}
