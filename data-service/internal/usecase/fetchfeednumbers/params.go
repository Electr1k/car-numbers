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

	Offset int
}

func (p Params) validate() error {
	switch {
	case p.Limit < minLimit:
		return fmt.Errorf("%w: limit must be at least %d, got %d", domain.ErrInvalidArgument, minLimit, p.Limit)
	case p.Limit > maxLimit:
		return fmt.Errorf("%w: limit must not exceed %d, got %d", domain.ErrInvalidArgument, maxLimit, p.Limit)
	case p.Offset < 0:
		return fmt.Errorf("%w: offset must not be negative, got %d", domain.ErrInvalidArgument, p.Offset)
	}

	return nil
}
