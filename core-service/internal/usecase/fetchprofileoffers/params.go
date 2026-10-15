package fetchprofileoffers

import (
	"core-service/internal/service"
	"core-service/internal/service/plate"
	"fmt"

	"github.com/google/uuid"
)

const (
	minLimit = 1
	maxLimit = 100
)

// Params - входные параметры выборки предложений продавца
type Params struct {
	ProfileID uuid.UUID
	Statuses  []string // Statuses - пустой список - все статусы, значения проверяет plate-service
	Limit     int
	Cursor    string
}

func (p Params) validate() error {
	switch {
	case p.Limit < minLimit:
		return fmt.Errorf("%w: limit must be at least %d, got %d", service.ErrBadRequest, minLimit, p.Limit)
	case p.Limit > maxLimit:
		return fmt.Errorf("%w: limit must not exceed %d, got %d", service.ErrBadRequest, maxLimit, p.Limit)
	}

	return nil
}

func (p Params) toClientParams() plate.FetchProfileOffersParams {
	return plate.FetchProfileOffersParams{
		ProfileID: p.ProfileID,
		Statuses:  p.Statuses,
		Limit:     p.Limit,
		Cursor:    p.Cursor,
	}
}
