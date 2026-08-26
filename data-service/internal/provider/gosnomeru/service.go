package gosnomeru

import (
	"context"
	"data-service/internal/domain"
	"data-service/internal/provider"
	"errors"
	"time"
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
func (s *Service) FetchOffers(ctx context.Context, page int) (provider.FetchResult, error) {

	response, err := s.client.FetchOffers(ctx, page)
	if err != nil {
		return provider.FetchResult{}, err
	}

	result := provider.FetchResult{TotalPages: response.TotalPages}
	for index, number := range response.Items {
		result.RowsFound++
		offer, err := s.mapper.MapOfferToDomain(number)
		if err != nil {
			result.RowErrors = append(result.RowErrors, provider.RowError{Index: index, Err: err})
			continue
		}

		result.Offers = append(result.Offers, offer)
	}

	return result, nil
}

func (s *Service) FetchOfferDetail(ctx context.Context, offer *domain.OfferWithNumber) (*domain.OfferWithNumber, error) {
	response, err := s.client.FetchOfferDetail(ctx, offer.Offer.ExternalId)
	if errors.Is(err, provider.ErrNotFound) {
		offer.Offer.Status = domain.OfferStatusInactive
		return offer, nil
	}
	if err != nil {
		return &domain.OfferWithNumber{}, err
	}

	time.Sleep(time.Second)
	return s.mapper.MapOfferDetailToDomain(*response, offer)
}
