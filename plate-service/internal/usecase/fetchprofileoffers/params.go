package fetchprofileoffers

import (
	"fmt"
	"plate-service/internal/domain"
	"plate-service/internal/domain/data"

	"github.com/google/uuid"
)

const (
	minLimit = 1
	maxLimit = 100
)

// Params - входные параметры выборки предложений продавца
type Params struct {
	ProfileID uuid.UUID
	Statuses  []domain.OfferStatus
	Limit     int
	Cursor    *data.ProfileOffersCursor
}

func (p *Params) validate() error {
	for _, status := range p.Statuses {
		if !status.Valid() {
			return fmt.Errorf("%w: unknown status %q", domain.ErrInvalidArgument, status)
		}
	}

	switch {
	case p.Limit < minLimit:
		return fmt.Errorf("%w: limit must be at least %d, got %d", domain.ErrInvalidArgument, minLimit, p.Limit)
	case p.Limit > maxLimit:
		return fmt.Errorf("%w: limit must not exceed %d, got %d", domain.ErrInvalidArgument, maxLimit, p.Limit)
	}

	return nil
}
