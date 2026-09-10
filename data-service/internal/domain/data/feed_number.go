package data

import (
	"data-service/internal/domain"
	"time"

	"github.com/google/uuid"
)

// FeedNumber - номер для выдачи в свежих номерах
type FeedNumber struct {
	ID              uuid.UUID         // ID - Идентификатор
	Number          string            // Number - Номер
	RegionName      *string           // RegionName - Название региона
	RegionCode      *string           // RegionCode - Код региона
	Price           *float64          // Price - Минимальная цена
	Type            domain.NumberType // Type - Тип номера
	Count           int               // Count - количество офферов
	RefreshedAt     time.Time         // RefreshedAt - максимальная дата обновления оффера у провайдера
	UpdatedAt       time.Time         // UpdatedAt - максимальная дата обновления оффера в системе
	ReissueIncluded *bool             // ReissueIncluded - включено ли переоформление
}

type FeedCursor struct {
	RefreshedAt time.Time `json:"refreshed_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ID          uuid.UUID `json:"id"`
}
