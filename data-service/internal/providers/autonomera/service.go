package autonomera

import (
	"bytes"
	"context"
	"data-service/internal/providers"
	"fmt"

	"github.com/PuerkitoBio/goquery"
)

type Service struct {
	client *Client
	mapper *Mapper
}

func NewService(client *Client, mapper *Mapper) *Service {
	return &Service{
		client: client,
		mapper: mapper,
	}
}

// FetchOffers забирает одну страницу раздела и маппит её в домен
func (s *Service) FetchOffers(ctx context.Context, section Section, offset int) (providers.FetchResult, error) {
	status, err := statusForSection(section)
	if err != nil {
		return providers.FetchResult{}, err
	}

	response, err := s.client.FetchOffersHTML(ctx, section, offset)
	if err != nil {
		return providers.FetchResult{}, err
	}

	document, err := goquery.NewDocumentFromReader(bytes.NewReader(response))
	if err != nil {
		return providers.FetchResult{}, fmt.Errorf("parse html document: %w", err)
	}

	var result providers.FetchResult

	document.Find(offerRowSelector).Each(func(index int, row *goquery.Selection) {
		result.RowsFound++

		offer, err := s.mapper.MapToDomain(row, status)
		if err != nil {
			result.RowErrors = append(result.RowErrors, providers.RowError{Index: index, Err: err})
			return
		}

		result.Offers = append(result.Offers, offer)
	})

	return result, nil
}
