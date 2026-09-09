package data

import (
	"data-service/internal/domain"
	"time"

	"github.com/google/uuid"
)

// Number - номер для выдачи
type Number struct {
	Id         uuid.UUID `validate:"required"`             // Id - Идентификатор
	Number     string    `validate:"required,min=8,max=9"` // Number - Номер
	RegionName *string   // RegionName - Название региона
	RegionCode *string   // RegionCode - Код региона
	Offers     []Offer   // Offers - номера в номере
}

// Offer - оффер в номере
type Offer struct {
	Id              uuid.UUID          `validate:"required"`
	Provider        domain.Provider    `validate:"required"`
	Price           *float64           `validate:"required"`
	Status          domain.OfferStatus `validate:"required"`
	ReissueIncluded *bool
	Whereabouts     *domain.OfferWhereabouts
	ViewCount       *int
	Comment         *string
	PostedAt        *time.Time
	RefreshedAt     *time.Time
	Url             string
}
