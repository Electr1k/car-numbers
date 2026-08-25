package importgosnomeru

import (
	"fmt"
	"time"
)

const (
	// defaultMaxPages - потолок страниц, когда он не задан явно
	defaultMaxPages = 10000
)

// Params - входные параметры импорта
type Params struct {
	// StartPage - с какой страницы начинать
	StartPage int

	// MaxPages - потолок страниц за прогон
	MaxPages int

	// StopAfter - не забирать публикации старше этого возраста
	StopAfter time.Duration
}

func (p Params) validate() error {
	switch {
	case p.StartPage < 0:
		return fmt.Errorf("start page must not be negative, got %d", p.StartPage)
	case p.MaxPages < 0:
		return fmt.Errorf("max pages must not be negative, got %d", p.MaxPages)
	case p.StopAfter < 0:
		return fmt.Errorf("stop after must not be negative, got %s", p.StopAfter)
	}

	return nil
}

// pageLimit - сколько страниц берём за прогон, считая от startPage
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
