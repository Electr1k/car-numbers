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
func (m *Mapper) MapOfferToDomain(externalOffer OffersItem) (domain.OfferWithPlate, error) {
	var emptyOffer domain.OfferWithPlate

	raw, err := json.Marshal(externalOffer)
	if err != nil {
		return emptyOffer, fmt.Errorf("%w: read row json: %w", provider.ErrBrokenOffer, err)
	}

	url := m.getOfferURL(externalOffer.ID, externalOffer.Slug)

	numberStr := externalOffer.Number.Letters + externalOffer.Number.Region
	plateType := domain.PlateTypeCar
	if slices.Contains(externalOffer.Categories, CategoryNotCar) {
		plateType = domain.PlateTypeMoto
	}

	plate, err := domain.NewPlate(numberStr, plateType)
	if err != nil {
		return emptyOffer, fmt.Errorf("%w: invalid number %q: %w", provider.ErrRowSkipped, numberStr, err)
	}

	status, err := mapStatus(externalOffer.Status)
	if err != nil {
		return emptyOffer, fmt.Errorf("%w: invalid status %q: %w", provider.ErrRowSkipped, status, err)
	}

	postedAt, err := mapPostedAt(externalOffer.Date)
	if err != nil {
		return emptyOffer, fmt.Errorf("%w: invalid date %q: %w", provider.ErrRowSkipped, externalOffer.Date, err)
	}

	var price *float64
	if externalOffer.Price > 0 {
		price = &externalOffer.Price
	}

	offer, err := domain.NewOffer(
		plate.ID,
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
		return emptyOffer, fmt.Errorf("%w: invalid offer %q: %w", provider.ErrRowSkipped, externalOffer.ID, err)
	}

	return domain.OfferWithPlate{Plate: plate, Offer: offer}, nil
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

func (m *Mapper) ApplyOfferDetailToDomain(response OfferDetail, offer domain.OfferWithPlate) (domain.OfferWithPlate, error) {
	var emptyOffer domain.OfferWithPlate

	newOffer, err := m.MapOfferDetailToDomain(response)
	if err != nil {
		return emptyOffer, err
	}

	if newOffer.Plate.Number != offer.Plate.Number {
		return emptyOffer, fmt.Errorf("%w: offer with number %q does not match its number %q", provider.ErrMapOffer, offer.Plate.Number, newOffer.Plate.Number)
	}

	viewCount := offer.Offer.ViewCount
	if viewCount == nil || newOffer.Offer.ViewCount != nil && *viewCount < *newOffer.Offer.ViewCount {
		viewCount = newOffer.Offer.ViewCount
	}
	_, err = offer.Offer.ApplyDetail(
		newOffer.Offer.Status,
		newOffer.Offer.Price,
		newOffer.Offer.Whereabouts,
		newOffer.Offer.ReissueIncluded,
		viewCount,
		newOffer.Offer.PostedAt,
		newOffer.Offer.RefreshedAt,
		*newOffer.Offer.RawDetailed,
		newOffer.Offer.Comment,
	)
	if err != nil {
		return emptyOffer, fmt.Errorf("%w: read row json: %w", provider.ErrRowSkipped, err)
	}

	return offer, nil
}

func (m *Mapper) MapOfferDetailToDomain(response OfferDetail) (domain.OfferWithPlate, error) {
	var emptyOffer domain.OfferWithPlate

	raw, err := json.Marshal(response)
	if err != nil {
		return emptyOffer, fmt.Errorf("%w: read row json: %w", provider.ErrBrokenOffer, err)
	}
	rawStr := string(raw)

	url := m.getOfferURL(response.ID, response.Slug)

	numberStr := response.Number.Letters + response.Number.Region
	plateType := domain.PlateTypeCar
	if slices.Contains(response.Categories, CategoryNotCar) {
		plateType = domain.PlateTypeMoto
	}

	plate, err := domain.NewPlate(numberStr, plateType)
	if err != nil {
		return emptyOffer, fmt.Errorf("%w: invalid number %q: %w", provider.ErrRowSkipped, numberStr, err)
	}

	status := domain.OfferStatusActive
	if response.Availability != nil && *response.Availability == AvailabilityRemoved || response.Status == StatusDeleted {
		status = domain.OfferStatusInactive
	}

	var price *float64
	if response.Price > 0 {
		price = &response.Price
	}

	posted, err := time.Parse(dateLayout, response.CreatedAt)
	if err != nil {
		return emptyOffer, fmt.Errorf("%w: invalid date %q: %w", provider.ErrBrokenOffer, response.CreatedAt, err)
	}
	postedAt := &posted

	refreshed, err := time.Parse(dateLayout, response.Date)
	if err != nil {
		return emptyOffer, fmt.Errorf("%w: invalid date %q: %w", provider.ErrBrokenOffer, response.Date, err)
	}
	refreshedAt := &refreshed

	var comment *string
	if len(strings.TrimSpace(response.Description)) > 0 {
		comment = &response.Description
	}

	offer, err := domain.NewOffer(
		plate.ID,
		domain.ProviderGosnomeru,
		response.ID,
		price,
		status,
		nil,
		&response.DealIncluded,
		&response.ViewCount,
		postedAt,
		refreshedAt,
		url,
		rawStr,
		&rawStr,
		comment,
	)
	if err != nil {
		return emptyOffer, fmt.Errorf("%w: invalid offer %q: %w", provider.ErrRowSkipped, response.ID, err)
	}

	return domain.OfferWithPlate{Plate: plate, Offer: offer}, nil
}

func (m *Mapper) getOfferURL(externalID string, slug string) string {
	return m.baseURL + "/plate/" + externalID + "-" + slug + ".html"
}
