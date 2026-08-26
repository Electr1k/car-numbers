package gosnomeru

import (
	"data-service/internal/domain"
	"data-service/internal/provider"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"
)

const (
	CategoryNotCar      = 11
	StatusPublished     = "published"
	StatusDeleted       = "deleted"
	AvailabilityRemoved = "removed"
	dateLayout          = "2006-01-02"
)

type Mapper struct {
	baseURL string
}

func NewMapper(baseURL string) *Mapper {
	return &Mapper{baseURL: strings.TrimRight(baseURL, "/")}
}

// MapOfferToDomain - Маппит строку выдачи в домен
func (m *Mapper) MapOfferToDomain(externalOffer OffersItem) (domain.OfferWithNumber, error) {
	var empty domain.OfferWithNumber

	raw, err := json.Marshal(externalOffer)
	if err != nil {
		return empty, fmt.Errorf("%w: read row json: %w", provider.ErrBrokenOffer, err)
	}

	url := m.baseURL + "/plate/" + externalOffer.ID + "-" + externalOffer.Slug + ".html"

	numberStr := externalOffer.Number.Letters + externalOffer.Number.Region
	numberType := domain.NumberTypeCar
	if slices.Contains(externalOffer.Categories, CategoryNotCar) {
		numberType = domain.NumberTypeMoto
	}

	number, err := domain.NewNumber(numberStr, numberType)
	if err != nil {
		return empty, fmt.Errorf("%w: invalid number %q: %w", provider.ErrRowSkipped, numberStr, err)
	}

	status, err := mapStatus(externalOffer.Status)
	if err != nil {
		return empty, fmt.Errorf("%w: invalid status %q: %w", provider.ErrRowSkipped, status, err)
	}

	postedAt, err := mapPostedAt(externalOffer.Date)
	if err != nil {
		return empty, fmt.Errorf("%w: invalid date %q: %w", provider.ErrRowSkipped, externalOffer.Date, err)
	}

	var price *float64
	if externalOffer.Price > 0 {
		price = &externalOffer.Price
	}

	offer, err := domain.NewOffer(
		number.Id,
		domain.ProviderGosnomeru,
		externalOffer.ID,
		price,
		status,
		nil,
		nil,
		nil,
		&postedAt,
		&postedAt,
		url,
		string(raw),
		nil,
		nil,
	)
	if err != nil {
		return empty, fmt.Errorf("%w: invalid offer %q: %w", provider.ErrRowSkipped, externalOffer.ID, err)
	}

	return domain.OfferWithNumber{Number: number, Offer: offer}, nil
}

// mapStatus - маппинг статуса
func mapStatus(status string) (domain.OfferStatus, error) {
	switch status {
	case StatusPublished:
		return domain.OfferStatusActive, nil
	case StatusDeleted:
		return domain.OfferStatusInactive, nil
	default:
		return "", fmt.Errorf("%w: unknown status %q",
			provider.ErrBrokenOffer, status)
	}
}

// parsePostedAt - дата публикации
func mapPostedAt(dateStr string) (time.Time, error) {
	date := strings.TrimSpace(dateStr)
	if date == "" {
		return time.Time{}, fmt.Errorf("%w: empty date cell", provider.ErrBrokenOffer)
	}

	postedAt, err := time.Parse(dateLayout, date)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: parse date %q: %w", provider.ErrBrokenOffer, date, err)
	}

	return postedAt, nil
}

func (m *Mapper) MapOfferDetailToDomain(response OfferDetail, offer *domain.OfferWithNumber) (*domain.OfferWithNumber, error) {
	raw, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("%w: read row json: %w", provider.ErrBrokenOffer, err)
	}

	offerNumber := response.Number.Letters + response.Number.Region
	if offerNumber != offer.Number.Number {
		return nil, fmt.Errorf("%w: offer with number %q does not match its number %q", provider.ErrMapOffer, offer.Number.Number, offerNumber)
	}

	status := domain.OfferStatusActive
	if response.Availability != nil && *response.Availability == AvailabilityRemoved || response.Status == StatusDeleted {
		status = domain.OfferStatusInactive
	}

	var price *float64 = nil
	if response.Price > 0 {
		price = &response.Price
	}

	posted, err := time.Parse(dateLayout, response.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid date %q: %w", provider.ErrBrokenOffer, response.CreatedAt, err)
	}
	postedAt := &posted

	refreshed, err := time.Parse(dateLayout, response.Date)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid date %q: %w", provider.ErrBrokenOffer, response.Date, err)
	}
	refreshedAt := &refreshed

	var comment *string = nil
	if len(strings.TrimSpace(response.Description)) > 0 {
		comment = &response.Description
	}

	_, err = offer.Offer.ApplyDetail(
		status,
		price,
		nil,
		&response.DealIncluded,
		&response.ViewCount,
		postedAt,
		refreshedAt,
		string(raw),
		comment,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: read row json: %w", provider.ErrRowSkipped, err)
	}

	return offer, nil
}
