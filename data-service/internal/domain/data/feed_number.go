package data

import (
	"data-service/internal/domain"
	"time"

	"github.com/google/uuid"
)

// FeedNumber - номер для выдачи в свежих номерах
type FeedNumber struct {
	Id              uuid.UUID         `validate:"required"`                        // Id - Идентификатор
	Number          string            `validate:"required,min=8,max=9"`            // Number - Номер
	RegionName      *string           `validate:"omitempty"`                       // RegionName - Название региона
	RegionCode      *string           `validate:"omitempty"`                       // RegionCode - Код региона
	Price           *float64          `validate:"omitempty"`                       // Price - Минимальная цена
	Type            domain.NumberType `validate:"required,oneof=car moto trailer"` // Type - Тип номера
	Count           int               `validate:"required"`                        // Count - количество офферов
	RefreshedAt     time.Time         `validate:"required"`                        // RefreshedAt - максимальная дата обновления оффера у провайдера
	UpdatedAt       time.Time         `validate:"required"`                        // UpdatedAt - максимальная дата обновления обновления оффера в системе
	ReissueIncluded *bool             `validate:"omitempty"`                       // ReissueIncluded - включено ли переоформления
}
