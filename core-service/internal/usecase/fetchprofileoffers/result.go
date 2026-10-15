package fetchprofileoffers

import (
	"time"

	"github.com/google/uuid"
)

// Result - результат выборки предложений продавца
type Result struct {
	Items      []Offer
	NextCursor *string
}

type Offer struct {
	ID          uuid.UUID
	PlateID     uuid.UUID
	Number      string
	Price       *float64
	Status      string
	PostedAt    time.Time
	RefreshedAt time.Time
	URL         string
}
