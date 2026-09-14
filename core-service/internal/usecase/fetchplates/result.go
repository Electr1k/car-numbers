package fetchplates

import (
	"time"

	"github.com/google/uuid"
)

// Result - результат выборки свежих номеров
type Result struct {
	Items      []Plate
	NextCursor *string
}

type Plate struct {
	ID              uuid.UUID
	Number          string
	Region          *Region
	Price           *float64
	Type            string
	Count           int
	RefreshedAt     time.Time
	UpdatedAt       time.Time
	ReissueIncluded *bool
}

type Region struct {
	Code string
	Name string
}
