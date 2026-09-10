package fetchfeednumbers

import (
	"data-service/internal/domain"
	"fmt"
)

const (
	minLimit = 1
	maxLimit = 25
)

// Params - входные параметры выборки свежих номеров
type Params struct {
	Limit int

	Cursor *string
}

func (p Params) validate() error {
	switch {
	case p.Limit < minLimit:
		return fmt.Errorf("%w: limit must be at least %d, got %d", domain.ErrInvalidArgument, minLimit, p.Limit)
	case p.Limit > maxLimit:
		return fmt.Errorf("%w: limit must not exceed %d, got %d", domain.ErrInvalidArgument, maxLimit, p.Limit)
	}

	return nil
}
