package provider

import (
	"data-service/internal/domain"
	"errors"
	"fmt"
	"time"
)

// RowError - ошибка разбора оффера
type RowError struct {
	Index int
	Err   error
}

// FetchResult - результат получения страницы офферов
type FetchResult struct {
	Offers     []domain.OfferWithNumber
	RowsFound  int
	RowErrors  []RowError
	TotalPages int
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
func (r FetchResult) ErrorMessages() []string {
	errs := make([]string, 0, len(r.RowErrors))
	for _, rowErr := range r.RowErrors {
		errs = append(errs, fmt.Sprintf("row %d: %s", rowErr.Index, rowErr.Err))
	}

	return errs
}

// OldestRefreshedAt - самая ранняя дата публикации в батче
func (r FetchResult) OldestRefreshedAt() time.Time {
	var oldest time.Time

	for _, item := range r.Offers {
		if item.Offer.RefreshedAt == nil {
			continue
		}

		if oldest.IsZero() || item.Offer.RefreshedAt.Before(oldest) {
			oldest = *item.Offer.RefreshedAt
		}
	}

	return oldest
}
