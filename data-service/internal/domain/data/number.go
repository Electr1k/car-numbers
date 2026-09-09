package data

import (
	"data-service/internal/domain"
	"time"

	"github.com/google/uuid"
)

// Number - номер для выдачи
type Number struct {
	ID         uuid.UUID // ID - Идентификатор
	Number     string    // Number - Номер
	RegionName *string   // RegionName - Название региона
	RegionCode *string   // RegionCode - Код региона
	Offers     []Offer   // Offers - предложения по номеру
}

// Offer - оффер в номере
type Offer struct {
	ID              uuid.UUID
	Provider        domain.Provider
	Price           *float64
	Status          domain.OfferStatus
	ReissueIncluded *bool
	Whereabouts     *domain.OfferWhereabouts
	ViewCount       *int
	Comment         *string
	PostedAt        time.Time
	RefreshedAt     time.Time
	URL             string
}
