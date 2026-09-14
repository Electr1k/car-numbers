package fetchfeednumbers

import (
	"data-service/internal/domain"
	"data-service/internal/domain/data"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	minLimit    = 1
	maxLimit    = 25
	maxQueryLen = 9
)

// Params - входные параметры выборки свежих номеров
type Params struct {
	Query           *string
	RegionId        *int
	PriceFrom       *float64
	PriceTo         *float64
	ReissueIncluded *bool
	CategoryIds     []int
	Limit           int
	Cursor          *data.FeedCursor
}

func (p *Params) validate() error {
	if p.Query != nil {
		str := strings.ToUpper(strings.TrimSpace(*p.Query))
		p.Query = &str
	}

	switch {
	case p.Limit < minLimit:
		return fmt.Errorf("%w: limit must be at least %d, got %d", domain.ErrInvalidArgument, minLimit, p.Limit)
	case p.Limit > maxLimit:
		return fmt.Errorf("%w: limit must not exceed %d, got %d", domain.ErrInvalidArgument, maxLimit, p.Limit)
	case p.Query != nil && utf8.RuneCountInString(*p.Query) > maxQueryLen:
		return fmt.Errorf("%w: query must not exceed 9, got %s", domain.ErrInvalidArgument, *p.Query)
	case p.Query != nil && !regexp.MustCompile(`^[АВЕКМНОРСТУХ0-9*]{1,9}$`).MatchString(*p.Query):
		return fmt.Errorf("%w: unknown symbol passed: %s", domain.ErrInvalidArgument, *p.Query)
	case p.Query != nil && utf8.RuneCountInString(*p.Query) >= 8 && !regexp.MustCompile(`^[АВЕКМНОРСТУХ*][0-9*]{3}[АВЕКМНОРСТУХ*]{2}[0-9*]{2,3}$`).MatchString(*p.Query):
		return fmt.Errorf("%w: invalid query: %s", domain.ErrInvalidArgument, *p.Query)
	case p.PriceFrom != nil && p.PriceTo != nil && *p.PriceFrom > *p.PriceTo:
		return fmt.Errorf("%w: price_from cannot be greater than price_to", domain.ErrInvalidArgument)
	}

	return nil
}
