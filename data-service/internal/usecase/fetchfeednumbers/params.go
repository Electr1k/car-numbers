package fetchfeednumbers

import (
	"fmt"
)

const maxLimit = 25

// Params - входные параметры импорта
type Params struct {
	Limit int

	Offset int
}

func (p Params) validate() error {
	switch {
	case p.Limit < 0:
		return fmt.Errorf("limit must not be negative, got %d", p.Limit)
	case p.Limit > maxLimit:
		return fmt.Errorf("limit must not be biggest %d, got %d", maxLimit, p.Limit)
	case p.Offset < 0:
		return fmt.Errorf("offset must not be negative, got %d", p.Offset)
	}

	return nil
}
