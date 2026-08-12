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

// BrokenOfferCount - количество офферов, которые не смогли разобрать
func (r FetchResult) BrokenOfferCount() int {
	count := 0
	for _, rowErr := range r.RowErrors {
		if errors.Is(rowErr.Err, ErrBrokenOffer) {
			count++
		}
	}

	return count
}

// ErrorMessages - тексты ошибок разбора, по одному на строку
// Именно тексты, а не []error: значения нужны только для лога
func (r FetchResult) ErrorMessages() []string {
	errs := make([]string, 0, len(r.RowErrors))
	for _, rowErr := range r.RowErrors {
		errs = append(errs, fmt.Sprintf("row %d: %s", rowErr.Index, rowErr.Err))
	}

	return errs
}
