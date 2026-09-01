package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrOfferNotFound - Предложение не найдено
var ErrOfferNotFound = errors.New("offer not found")

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

type OfferWhereabouts string

const (
	// OfferWhereaboutsOnCar - на машине
	OfferWhereaboutsOnCar OfferWhereabouts = "on-car"

	// OfferWhereaboutsOnStorage - на хранении
	OfferWhereaboutsOnStorage OfferWhereabouts = "on-storage"
)

// Offer - Предложение
type Offer struct {
	Id              uuid.UUID         `validate:"required"`                                    // Id - Идентификатор
	NumberId        uuid.UUID         `validate:"required"`                                    // NumberId - Идентификатор номера
	Provider        Provider          `validate:"required,oneof=autonomera gosnomeru anomera"` // Provider - Провайдер, в котором найдено предложение
	ExternalId      string            `validate:"required"`                                    // ExternalId - Идентификатор у провайдера
	Price           *float64          `validate:"omitempty,gt=0"`                              // Price - Цена
	Status          OfferStatus       `validate:"required,oneof=active sold inactive"`         // Status - Статус предложения
	Whereabouts     *OfferWhereabouts `validate:"omitempty,oneof=on-car on-storage"`           // Whereabouts - Где находится номер (на авто/хранении)
	ReissueIncluded *bool             `validate:"omitempty"`                                   // ReissueIncluded - Переоформление включено в стоимость
	ViewCount       *int              `validate:"omitempty,gte=0"`                             // ViewCount - Число просмотров у провайдера
	PostedAt        *time.Time        `validate:"required"`                                    // PostedAt - Дата создания предложения у провайдера
	RefreshedAt     *time.Time        `validate:"required"`                                    // RefreshedAt - Дата поднятия/обновления предложения у провайдера
	Url             string            `validate:"required"`                                    // Url - Ссылка на предложение
	Raw             string            `validate:"required"`                                    // Raw - Сырой объект поставщика
	RawDetailed     *string           `validate:"omitempty"`                                   // RawDetailed - Сырой объект поставщика с детальной информацией
	Comment         *string           `validate:"omitempty"`                                   // Comment - Комментарий
	CreatedAt       *time.Time
	UpdatedAt       *time.Time
}

// ApplyDetail - Добавление детальной информации в предложение
func (o *Offer) ApplyDetail(
	status OfferStatus,
	price *float64,
	whereabouts *OfferWhereabouts,
	reissueIncluded *bool,
	viewCount *int,
	postedAt *time.Time,
	refreshAt *time.Time,
	rawDetailed string,
	comment *string,
) (*Offer, error) {
	o.Status = status
	o.Price = price
	o.Whereabouts = whereabouts
	o.ReissueIncluded = reissueIncluded
	o.ViewCount = viewCount
	o.RawDetailed = &rawDetailed
	if postedAt != nil {
		o.PostedAt = postedAt
	}
	if refreshAt != nil {
		o.RefreshedAt = refreshAt
	}
	o.Comment = comment

	if err := validate.Struct(o); err != nil {
		return nil, err
	}

	return o, nil
}

// NewOffer - Создание предложения, которого ещё не существовало
func NewOffer(
	numberId uuid.UUID,
	provider Provider,
	externalId string,
	price *float64,
	status OfferStatus,
	whereabouts *OfferWhereabouts,
	reissueIncluded *bool,
	viewCount *int,
	postedAt *time.Time,
	refreshedAt *time.Time,
	url string,
	raw string,
	rawDetailed *string,
	comment *string,
) (*Offer, error) {
	id, err := newID()
	if err != nil {
		return nil, err
	}

	return newOffer(
		id,
		numberId,
		provider,
		externalId,
		price,
		status,
		whereabouts,
		reissueIncluded,
		viewCount,
		postedAt,
		refreshedAt,
		url,
		raw,
		rawDetailed,
		comment,
		nil,
		nil,
	)
}

// RestoreOffer - Восстановление существующего предложения из хранилища
func RestoreOffer(
	id uuid.UUID,
	numberId uuid.UUID,
	provider Provider,
	externalId string,
	price *float64,
	status OfferStatus,
	whereabouts *OfferWhereabouts,
	reissueIncluded *bool,
	viewCount *int,
	postedAt *time.Time,
	refreshedAt *time.Time,
	url string,
	raw string,
	rawDetailed *string,
	comment *string,
	createdAt *time.Time,
	updatedAt *time.Time,
) (*Offer, error) {
	return newOffer(
		id,
		numberId,
		provider,
		externalId,
		price,
		status,
		whereabouts,
		reissueIncluded,
		viewCount,
		postedAt,
		refreshedAt,
		url,
		raw,
		rawDetailed,
		comment,
		createdAt,
		updatedAt,
	)
}

func newOffer(
	id uuid.UUID,
	numberId uuid.UUID,
	provider Provider,
	externalId string,
	price *float64,
	status OfferStatus,
	whereabouts *OfferWhereabouts,
	reissueIncluded *bool,
	viewCount *int,
	postedAt *time.Time,
	refreshedAt *time.Time,
	url string,
	raw string,
	rawDetailed *string,
	comment *string,
	createdAt *time.Time,
	updatedAt *time.Time,
) (*Offer, error) {
	offer := &Offer{
		Id:              id,
		NumberId:        numberId,
		Provider:        provider,
		ExternalId:      externalId,
		Price:           price,
		Status:          status,
		Whereabouts:     whereabouts,
		ReissueIncluded: reissueIncluded,
		ViewCount:       viewCount,
		PostedAt:        postedAt,
		RefreshedAt:     refreshedAt,
		Url:             url,
		Raw:             raw,
		RawDetailed:     rawDetailed,
		Comment:         comment,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
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
