package importautonomera

import (
	"data-service/internal/provider/autonomera"
	"fmt"
	"time"
)

// defaultMaxPages - потолок страниц, когда он не задан явно
const defaultMaxPages = 30000

// Params - входные параметры импорта
type Params struct {
	// Section - раздел выдачи, из которого забираем предложения
	Section autonomera.Section

	// StartOffset - с какой позиции выдачи начинать
	StartOffset int

	// MaxPages - потолок страниц за прогон
	MaxPages int

	// StopAfter - не забирать публикации старше этого возраста. 0 - без ограничений
	StopAfter time.Duration
}

func (p Params) validate() error {
	switch {
	case p.StartOffset < 0:
		return fmt.Errorf("start offset must not be negative, got %d", p.StartOffset)
	case p.MaxPages < 0:
		return fmt.Errorf("max pages must not be negative, got %d", p.MaxPages)
	case p.StopAfter < 0:
		return fmt.Errorf("stop after must not be negative, got %s", p.StopAfter)
	}

	return nil
}

// pageLimit - потолок страниц с подставленным умолчанием
func (p Params) pageLimit() int {
	if p.MaxPages == 0 {
		return defaultMaxPages
	}

	return p.MaxPages
}

// stopDate - нижняя граница импорта по дате публикации
func (p Params) stopDate() time.Time {
	if p.StopAfter == 0 {
		return time.Time{}
	}

	return time.Now().Add(-p.StopAfter)
}
