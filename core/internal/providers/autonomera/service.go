package autonomera

import (
	"context"
	"core/internal/domain"
	"core/pkg/integration/autonomera"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Service struct {
	client autonomera.Client
	mapper Mapper
}

func NewService(client autonomera.Client, mapper Mapper) *Service {
	return &Service{
		client: client,
		mapper: mapper,
	}
}

type Offer struct {
	Number *domain.Number
	Offer  *domain.Offer
}

func (s *Service) FetchOffers(ctx context.Context, start int) ([]Offer, error) {
	response, err := s.client.FetchNumbersHTML(ctx, start)
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
		slog.Log(ctx, slog.LevelError, "Bad offers", "errors", errors.Join(parseErrors...))
	}

	return offers, nil
}
