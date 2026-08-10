package domain

import (
	"time"
)

// Offer - Предложение
type Offer struct {
	Id         *int       `validate:"omitempty"`     // Id - Идентификатор
	NumberId   *int       `validate:"omitempty"`     // NumberId - Идентификатор номера
	Provider   string     `validate:"required"`      // Provider - Провайдер, в котором найдено предложение
	ExternalId string     `validate:"required"`      // ExternalId - Идентификатор у провайдера
	Price      float32    `validate:"required,gt=0"` // Price - Цена
	Status     string     `validate:"required"`      // Status - Стутас предложения (active, sold, unactive)
	PostedAt   *time.Time `validate:"required"`      // PostedAt - Дата публикации
	Url        string     `validate:"required"`      // Url - Ссылка на предложение
	Raw        string     `validate:"required"`      // Raw - Сырой объект поставщика
	CreatedAt  *time.Time `validate:""`
	UpdateAt   *time.Time `validate:""`
}

func NewOffer(
	id *int,
	numberId *int,
	provider string,
	externalId string,
	price float32,
	status string,
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
		UpdateAt:   updatedAt,
	}

	if err := offer.Validate(); err != nil {
		return nil, err
	}

	return offer, nil
}

func (o *Offer) Validate() error {
	return validate.Struct(o)
}
