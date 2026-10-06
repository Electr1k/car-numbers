package autonomera

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"plate-service/internal/domain"
	"plate-service/internal/provider"

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
func (s *Service) FetchOffers(ctx context.Context, section Section, offset int) (provider.FetchResult, error) {
	status, err := statusForSection(section)
	if err != nil {
		return provider.FetchResult{}, err
	}

	response, err := s.client.FetchOffersHTML(ctx, section, offset)
	if err != nil {
		return provider.FetchResult{}, err
	}

	document, err := goquery.NewDocumentFromReader(bytes.NewReader(response))
	if err != nil {
		return provider.FetchResult{}, fmt.Errorf("parse html document: %w", err)
	}

	var result provider.FetchResult

	document.Find(offerRowSelector).Each(func(index int, row *goquery.Selection) {
		result.RowsFound++

		offer, err := s.mapper.MapOfferToDomain(row, status)
		if err != nil {
			result.RowErrors = append(result.RowErrors, provider.RowError{Index: index, Err: err})
			return
		}

		result.Offers = append(result.Offers, offer)
	})

	return result, nil
}

func (s *Service) FetchOfferDetail(ctx context.Context, offer domain.OfferWithPlate) (domain.OfferWithPlate, error) {
	response, err := s.client.FetchOfferDetailHTML(ctx, offer.Offer.URL)
	if errors.Is(err, provider.ErrNotFound) {
		offer.Offer.Status = domain.OfferStatusInactive
		return offer, nil
	}
	if err != nil {
		return domain.OfferWithPlate{}, err
	}

	document, err := goquery.NewDocumentFromReader(bytes.NewReader(response))
	if err != nil {
		return domain.OfferWithPlate{}, fmt.Errorf("parse html document: %w", err)
	}

	return s.mapper.MapOfferDetailToDomain(document.Find(offerDetailSelector), offer)
}

func (s *Service) FetchProfileExternalIds(ctx context.Context, offset int) ([]string, error) {
	var result = make([]string, 0)

	response, err := s.client.FetchUsersHTML(ctx, offset)
	if err != nil {
		return result, err
	}

	document, err := goquery.NewDocumentFromReader(bytes.NewReader(response))
	if err != nil {
		return result, fmt.Errorf("parse html document: %w", err)
	}

	document.Find(userRowSelector).Each(func(index int, row *goquery.Selection) {
		id, mapError := s.mapper.MapUserToExternalProfileId(row)
		if mapError != nil {
			err = mapError
			return
		}

		result = append(result, id)
	})

	if err != nil {
		return result, err
	}

	return result, nil
}

func (s *Service) FetchProfile(ctx context.Context, id string) (domain.Profile, error) {
	response, err := s.client.FetchUserHTML(ctx, id)
	if err != nil {
		return domain.Profile{}, err
	}

	document, err := goquery.NewDocumentFromReader(bytes.NewReader(response))
	if err != nil {
		return domain.Profile{}, fmt.Errorf("parse html document: %w", err)
	}

	return s.mapper.MapUserDetailToProfile(document.Find(userSelector), id)
}
