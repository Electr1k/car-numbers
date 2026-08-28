package importgosnomerubydetail

import (
	"fmt"
)

type Params struct {
	// StartId - стартовый id поставщика
	StartId int

	// EndId - конечный id поставщика
	EndId int
}

func (p Params) validate() error {
	switch {
	case p.StartId < 0:
		return fmt.Errorf("start id must not be negative, got %d", p.StartId)
	case p.EndId < 0:
		return fmt.Errorf("end id must not be negative, got %d", p.EndId)
	}

	return nil
}
