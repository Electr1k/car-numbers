package gosnomeru

import (
	"context"
	"data-service/internal/domain"
	"data-service/internal/provider"
	"errors"
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

func (s *Service) FetchOfferDetail(ctx context.Context, offer domain.OfferWithNumber) (domain.OfferWithNumber, error) {
	response, err := s.client.FetchOfferDetail(ctx, offer.Offer.ExternalId)
	if errors.Is(err, provider.ErrNotFound) {
		offer.Offer.Status = domain.OfferStatusInactive
		return offer, nil
	}
	if err != nil {
		return domain.OfferWithNumber{}, err
	}

	return s.mapper.ApplyOfferDetailToDomain(*response, offer)
}

func (s *Service) FetchOfferDetailByExternalId(ctx context.Context, externalId string) (domain.OfferWithNumber, error) {
	var emptyOffer domain.OfferWithNumber

	response, err := s.client.FetchOfferDetail(ctx, externalId)
	if errors.Is(err, provider.ErrNotFound) {
		return emptyOffer, provider.ErrNotFound
	}
	if err != nil {
		return emptyOffer, err
	}

	offer, err := s.mapper.MapOfferDetailToDomain(*response)
	if err != nil {
		return emptyOffer, err
	}

	return offer, nil
}

func (s *Service) FetchLatestOffers(ctx context.Context) (provider.FetchResult, error) {
	response, err := s.client.FetchLatestOffers(ctx)
	if err != nil {
		return provider.FetchResult{}, err
	}

	result := provider.FetchResult{TotalPages: 0}
	for index, number := range response.Items {
		result.RowsFound++
		offer, err := s.mapper.MapOfferToDomain(number)
		if err != nil {
			result.RowErrors = append(result.RowErrors, provider.RowError{Index: index, Err: err})
			continue
		}

		result.Offers = append(result.Offers, offer)
	}

	return result, err
}
