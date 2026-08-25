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
	CategoryNotCar  = 11
	StatusPublished = "published"
	StatusDeleted   = "deleted"
	dateLayout      = "2006-01-02"
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

	url := "https://gosnomeru.com/plate/" + externalOffer.ID + "-" + externalOffer.Slug + ".html"

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

//
//func (m *Mapper) MapOfferDetailToDomain(sel *goquery.Selection, offer *domain.OfferWithNumber) (*domain.OfferWithNumber, error) {
//	raw, err := sel.Html()
//	if err != nil {
//		return nil, fmt.Errorf("%w: read row html: %w", provider.ErrBrokenOffer, err)
//	}
//
//	number, err := parseNumberFromDetail(sel)
//	if err != nil {
//		return nil, err
//	}
//
//	if offer.Number.Type == domain.NumberTypeMoto {
//		runes := []rune(number)
//		// Для мото буквы и цифры при парсинге склиеваются не в том порядке - меняем местами
//		letters := runes[0:2]
//		digits := runes[2:6]
//		reg := runes[6:]
//
//		number = string(digits) + string(letters) + string(reg)
//	}
//
//	if number != offer.Number.Number {
//		return nil, fmt.Errorf("%w: offer with number %q does not match its number %q", provider.ErrMapOffer, offer.Number.Number, number)
//	}
//
//	whereAbouts := parseWhereaboutsFromDetail(sel)
//	reissueIncluded := parseReissueFromDetailed(sel)
//
//	var (
//		price       *float64
//		views       *int
//		postedAt    *time.Time
//		refreshedAt *time.Time
//		parseErr    error
//	)
//
//	table := sel.Find(".article__table.user-data-table")
//	if table.Length() == 0 {
//		return nil, fmt.Errorf("%w: no user-data table", provider.ErrBrokenOffer)
//	}
//	table.Find("div.user-data-table__tr").Each(func(i int, s *goquery.Selection) {
//		title := strings.TrimSpace(s.Find(".user-data-table__th").Text())
//		value := s.Find(".user-data-table__td")
//		switch title {
//
//		case "Цена авто с номером":
//			price, parseErr = parsePrice(value.Text())
//		case "Просмотров":
//			var count int
//			count, parseErr = strconv.Atoi(strings.TrimSpace(value.Text()))
//			if parseErr == nil {
//				views = &count
//			}
//		case "Дата размещения":
//			postedAt, parseErr = parseDateFromDetail(value.Text())
//		case "Дата поднятия":
//			refreshedAt, parseErr = parseDateFromDetail(value.Text())
//		}
//	})
//
//	if parseErr != nil {
//		return nil, parseErr
//	}
//
//	commentText := strings.TrimSpace(sel.Find(".article-comment__content").Text())
//	var comment *string = nil
//	if len(commentText) != 0 {
//		comment = &commentText
//	}
//
//	_, err = offer.Offer.ApplyDetail(
//		offer.Offer.Status,
//		price,
//		whereAbouts,
//		reissueIncluded,
//		views,
//		postedAt,
//		refreshedAt,
//		raw,
//		comment,
//	)
//	if err != nil {
//		return nil, fmt.Errorf("%w: read row html: %w", provider.ErrRowSkipped, err)
//	}
//
//	return offer, nil
//}
//
//func parseNumberFromDetail(sel *goquery.Selection) (string, error) {
//	var parts []string
//
//	// Символы номера лежат в value каждого input.filter-plate-number__input, в порядке DOM
//	sel.Find(offerDetailNumber).Each(func(i int, s *goquery.Selection) {
//		if val, exists := s.Attr("value"); exists {
//			parts = append(parts, val)
//		}
//	})
//
//	region, exists := sel.Find(offerDetailRegion).Attr("value")
//	if !exists || len(region) > 3 || len(parts) < 2 {
//		return "", fmt.Errorf("%w: invalid offer detail number %q region %q",
//			provider.ErrBrokenOffer, strings.Join(parts, ""), region)
//	}
//
//	return strings.Join(parts, "") + region, nil
//}
//
//func parseWhereaboutsFromDetail(sel *goquery.Selection) *domain.OfferWhereabouts {
//	onCar := sel.Find(offerDetailOnCar).Nodes
//	if onCar != nil {
//		whereAbouts := domain.OfferWhereaboutsOnCar
//		return &whereAbouts
//	}
//
//	onStorage := sel.Find(offerDetailOnStorage).Nodes
//	if onStorage != nil {
//		whereAbouts := domain.OfferWhereaboutsOnStorage
//		return &whereAbouts
//	}
//
//	return nil
//}
//
//func parseReissueFromDetailed(sel *goquery.Selection) *bool {
//	reissueInclude := sel.Find(offerDetailReissueInclude).Nodes
//	if reissueInclude != nil {
//		t := true
//		return &t
//	}
//
//	return nil
//}
//
//func parseDateFromDetail(strDate string) (*time.Time, error) {
//	months := map[string]string{
//		"января":   "January",
//		"февраля":  "February",
//		"марта":    "March",
//		"апреля":   "April",
//		"мая":      "May",
//		"июня":     "June",
//		"июля":     "July",
//		"августа":  "August",
//		"сентября": "September",
//		"октября":  "October",
//		"ноября":   "November",
//		"декабря":  "December",
//	}
//
//	parts := strings.Fields(strDate)
//	if len(parts) != 3 {
//		return nil, fmt.Errorf("%w: invalid date format", provider.ErrBrokenOffer)
//	}
//
//	englishMonth := months[parts[1]]
//	if englishMonth == "" {
//		return nil, fmt.Errorf("%w: invalid date format", provider.ErrBrokenOffer)
//	}
//
//	englishDate := fmt.Sprintf("%s %s %s", parts[0], englishMonth, parts[2])
//
//	parsedTime, err := time.Parse("02 January 2006", englishDate)
//	if err != nil {
//		return nil, fmt.Errorf("%w: invalid date format", provider.ErrBrokenOffer)
//	}
//
//	return &parsedTime, nil
//}
