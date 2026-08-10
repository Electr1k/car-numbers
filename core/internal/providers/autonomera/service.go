package autonomera

import (
	"context"
	"core/internal/domain"
	"core/internal/providers"
	"core/pkg/integration/autonomera"
	"errors"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Service struct {
	client autonomera.Client
	mapper Mapper
}

type Offer struct {
	Number *domain.Number
	Offer  *domain.Offer
}

func NewService(client autonomera.Client, mapper Mapper) *Service {
	return &Service{
		client: client,
		mapper: mapper,
	}
}

func (s *Service) Parse(ctx context.Context) ([]Offer, error) {
	response, err := s.client.FetchNumbersHTML(ctx, 0)
	if err != nil {
		return nil, err
	}

	html, err := goquery.NewDocumentFromReader(strings.NewReader(string(response)))
	if err != nil {
		return nil, fmt.Errorf("parse html document: %w", err)
	}

	var offers []Offer
	var parseErrors []error

	html.Find("a.table__tr--td").Each(func(i int, html *goquery.Selection) {
		number, offer, err := s.mapper.MapToDomain(html)

		if err != nil {
			parseErrors = append(parseErrors, fmt.Errorf("row %d: %w", i, err))
			return
		}

		offers = append(offers, Offer{number, offer})
	})

	if len(parseErrors) > 0 && len(offers) == 0 {
		return nil, fmt.Errorf("%w: all %d rows failed: %w",
			providers.ErrParsingFailed, len(parseErrors), errors.Join(parseErrors...))
	}

	return offers, nil
}
