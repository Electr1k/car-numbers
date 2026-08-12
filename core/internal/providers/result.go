package providers

import (
	"core/internal/domain"
	"errors"
	"fmt"
)

// RowError - ошибка разбора оффера
type RowError struct {
	Index int
	Err   error
}

// FetchResult - результат получения страницы офферов
type FetchResult struct {
	Offers    []domain.OfferWithNumber
	RowsFound int
	RowErrors []RowError
}

// SchemaDriftCount - сколько строк упало из-за изменившейся разметки
func (r FetchResult) SchemaDriftCount() int {
	count := 0
	for _, rowErr := range r.RowErrors {
		if errors.Is(rowErr.Err, ErrSchemaDrift) {
			count++
		}
	}

	return count
}

func (r FetchResult) Reasons() []string {
	reasons := make([]string, 0, len(r.RowErrors))
	for _, rowErr := range r.RowErrors {
		reasons = append(reasons, fmt.Sprintf("row %d: %s", rowErr.Index, rowErr.Err))
	}

	return reasons
}
