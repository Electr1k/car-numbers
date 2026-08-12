package autonomera

import (
	"core/internal/domain"
	"core/internal/providers"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

const (
	// offerRowSelector - строка таблицы с одним предложением
	offerRowSelector = "a.table__tr--td"

	// dateSelector - ячейка с датой публикации
	dateSelector = ".table-date span"

	// priceSelector - ячейка с ценой
	priceSelector = ".table-price"

	// externalIDPrefix - префикс атрибута id, за ним идёт идентификатор провайдера
	externalIDPrefix = "item_number_id"

	// dateLayout - формат даты публикации
	dateLayout = "02.01.2006"
)

type Mapper struct {
	baseURL string
}

func NewMapper(baseURL string) *Mapper {
	return &Mapper{baseURL: strings.TrimRight(baseURL, "/")}
}

// MapToDomain - Маппит строку выдачи в домен
func (m *Mapper) MapToDomain(sel *goquery.Selection, status domain.OfferStatus) (domain.OfferWithNumber, error) {
	var empty domain.OfferWithNumber

	raw, err := sel.Html()
	if err != nil {
		return empty, fmt.Errorf("%w: read row html: %w", providers.ErrBrokenOffer, err)
	}

	idAttr, err := requiredAttr(sel, "id")
	if err != nil {
		return empty, err
	}

	externalID, err := parseExternalID(idAttr)
	if err != nil {
		return empty, err
	}

	href, err := requiredAttr(sel, "href")
	if err != nil {
		return empty, err
	}

	numberType, err := parseNumberType(href)
	if err != nil {
		return empty, err
	}

	title, err := requiredAttr(sel, "title")
	if err != nil {
		return empty, err
	}

	postedAt, err := parsePostedAt(sel)
	if err != nil {
		return empty, err
	}

	price, err := parsePrice(sel.Find(priceSelector).Text())
	if err != nil {
		return empty, err
	}

	number, err := domain.NewNumber(title, numberType)
	if err != nil {
		return empty, fmt.Errorf("%w: invalid number %q: %w", providers.ErrRowSkipped, title, err)
	}

	offer, err := domain.NewOffer(
		number.Id,
		domain.ProviderAutonomera,
		externalID,
		price,
		status,
		&postedAt,
		m.baseURL+href,
		raw,
	)
	if err != nil {
		return empty, fmt.Errorf("%w: invalid offer %q: %w", providers.ErrRowSkipped, externalID, err)
	}

	return domain.OfferWithNumber{Number: number, Offer: offer}, nil
}

// requiredAttr - обязательный атрибут строки
func requiredAttr(sel *goquery.Selection, name string) (string, error) {
	value, exists := sel.Attr(name)
	if !exists || value == "" {
		return "", fmt.Errorf("%w: missing %s attribute", providers.ErrBrokenOffer, name)
	}

	return value, nil
}

// parseExternalID - идентификатор провайдера из атрибута вида item_number_id12345
func parseExternalID(idAttr string) (string, error) {
	externalID := strings.TrimPrefix(idAttr, externalIDPrefix)

	if externalID == idAttr {
		return "", fmt.Errorf("%w: id attribute %q has no %q prefix",
			providers.ErrBrokenOffer, idAttr, externalIDPrefix)
	}

	if externalID == "" {
		return "", fmt.Errorf("%w: empty external id in %q", providers.ErrBrokenOffer, idAttr)
	}

	return externalID, nil
}

// parseNumberType - тип ТС из первого сегмента href вида /standart/а123аа77
func parseNumberType(href string) (domain.NumberType, error) {
	segments := strings.Split(strings.TrimPrefix(href, "/"), "/")
	if len(segments) == 0 || segments[0] == "" {
		return "", fmt.Errorf("%w: no vehicle type in href %q", providers.ErrBrokenOffer, href)
	}

	switch segments[0] {
	case "standart":
		return domain.NumberTypeCar, nil
	case "moto":
		return domain.NumberTypeMoto, nil
	case "trailer":
		return domain.NumberTypeTrailer, nil
	default:
		return "", fmt.Errorf("%w: unknown vehicle type %q in href %q",
			providers.ErrBrokenOffer, segments[0], href)
	}
}

// parsePostedAt - дата публикации
func parsePostedAt(sel *goquery.Selection) (time.Time, error) {
	date := strings.TrimSpace(sel.Find(dateSelector).Text())
	if date == "" {
		return time.Time{}, fmt.Errorf("%w: empty date cell", providers.ErrBrokenOffer)
	}

	postedAt, err := time.Parse(dateLayout, date)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: parse date %q: %w", providers.ErrBrokenOffer, date, err)
	}

	return postedAt, nil
}

// parsePrice - цена в рублях
func parsePrice(priceText string) (float64, error) {
	// Выкидывает любой пробельный символ
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}

		return r
	}, priceText)

	cleaned = strings.ReplaceAll(cleaned, "₽", "")

	if cleaned == "" {
		return 0, fmt.Errorf("%w: empty price cell", providers.ErrBrokenOffer)
	}

	price, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: price %q is not a number",
			providers.ErrRowSkipped, strings.TrimSpace(priceText))
	}

	return price, nil
}
