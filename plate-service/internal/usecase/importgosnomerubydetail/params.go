package importgosnomerubydetail

import (
	"fmt"
)

type Params struct {
	// StartID - стартовый id поставщика
	StartID int

	// EndID - конечный id поставщика
	EndID int
}

func (p Params) validate() error {
	switch {
	case p.StartID < 0:
		return fmt.Errorf("start id must not be negative, got %d", p.StartID)
	case p.EndID < 0:
		return fmt.Errorf("end id must not be negative, got %d", p.EndID)
	}

	return nil
}
