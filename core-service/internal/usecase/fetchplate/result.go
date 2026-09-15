package fetchplate

import (
	"time"

	"github.com/google/uuid"
)

// Result - деталка номера
type Result struct {
	ID            uuid.UUID
	Number        string
	Region        *Region
	ActiveOffers  []Offer
	ArchiveOffers []Offer
}

type Offer struct {
	ID              uuid.UUID
	Provider        string
	Price           *float64
	Status          string
	ReissueIncluded *bool
	Whereabouts     *string
	ViewCount       *int
	Comment         *string
	RefreshedAt     time.Time
	PostedAt        time.Time
	URL             string
}

type Region struct {
	Code string
	Name string
}
