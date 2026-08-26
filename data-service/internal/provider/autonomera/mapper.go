package autonomera

import (
	"data-service/internal/domain"
	"data-service/internal/provider"
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

	offerDetailSelector = "div.wrap-table"

	offerDetailNumber = "input.filter-plate-number__input"

	offerDetailRegion = "input.filter-plate-region-code__input"

	offerDetailOnCar = "div.func__item--on-car"

	offerDetailOnStorage = "div.func__item--in-storage"

	offerDetailReissueInclude = "div.func__item--key"

	priceNegotiable = "Договорная"
)

type Mapper struct {
	baseURL string
}

func NewMapper(baseURL string) *Mapper {
	return &Mapper{baseURL: strings.TrimRight(baseURL, "/")}
}

// MapOfferToDomain - Маппит строку выдачи в домен
func (m *Mapper) MapOfferToDomain(sel *goquery.Selection, status domain.OfferStatus) (domain.OfferWithNumber, error) {
	var empty domain.OfferWithNumber

	raw, err := sel.Html()
	if err != nil {
		return empty, fmt.Errorf("%w: read row html: %w", provider.ErrBrokenOffer, err)
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
		return empty, fmt.Errorf("%w: invalid number %q: %w", provider.ErrRowSkipped, title, err)
	}

	offer, err := domain.NewOffer(
		number.Id,
		domain.ProviderAutonomera,
		externalID,
		price,
		status,
		nil,
		nil,
		nil,
		&postedAt,
		&postedAt,
		m.baseURL+href,
		raw,
		nil,
		nil,
	)
	if err != nil {
		return empty, fmt.Errorf("%w: invalid offer %q: %w", provider.ErrRowSkipped, externalID, err)
	}

	return domain.OfferWithNumber{Number: number, Offer: offer}, nil
}

// requiredAttr - обязательный атрибут строки
func requiredAttr(sel *goquery.Selection, name string) (string, error) {
	value, exists := sel.Attr(name)
	if !exists || value == "" {
		return "", fmt.Errorf("%w: missing %s attribute", provider.ErrBrokenOffer, name)
	}

	return value, nil
}

// parseExternalID - идентификатор провайдера из атрибута вида item_number_id12345
func parseExternalID(idAttr string) (string, error) {
	externalID := strings.TrimPrefix(idAttr, externalIDPrefix)

	if externalID == idAttr {
		return "", fmt.Errorf("%w: id attribute %q has no %q prefix",
			provider.ErrBrokenOffer, idAttr, externalIDPrefix)
	}

	if externalID == "" {
		return "", fmt.Errorf("%w: empty external id in %q", provider.ErrBrokenOffer, idAttr)
	}

	return externalID, nil
}

// parseNumberType - тип ТС из первого сегмента href вида /standart/а123аа77
func parseNumberType(href string) (domain.NumberType, error) {
	segments := strings.Split(strings.TrimPrefix(href, "/"), "/")
	if len(segments) == 0 || segments[0] == "" {
		return "", fmt.Errorf("%w: no vehicle type in href %q", provider.ErrBrokenOffer, href)
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
			provider.ErrBrokenOffer, segments[0], href)
	}
}

// parsePostedAt - дата публикации
func parsePostedAt(sel *goquery.Selection) (time.Time, error) {
	date := strings.TrimSpace(sel.Find(dateSelector).Text())
	if date == "" {
		return time.Time{}, fmt.Errorf("%w: empty date cell", provider.ErrBrokenOffer)
	}

	postedAt, err := time.Parse(dateLayout, date)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: parse date %q: %w", provider.ErrBrokenOffer, date, err)
	}

	return postedAt, nil
}

// parsePrice - цена в рублях
func parsePrice(priceText string) (*float64, error) {
	// Выкидывает любой пробельный символ
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}

		return r
	}, priceText)

	cleaned = strings.ReplaceAll(cleaned, "₽", "")
	if cleaned == priceNegotiable {
		return nil, nil
	}

	if cleaned == "" {
		return nil, fmt.Errorf("%w: empty price cell", provider.ErrBrokenOffer)
	}

	price, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: price %q is not a number",
			provider.ErrRowSkipped, strings.TrimSpace(priceText))
	}

	return &price, nil
}

func (m *Mapper) MapOfferDetailToDomain(sel *goquery.Selection, offer *domain.OfferWithNumber) (*domain.OfferWithNumber, error) {
	raw, err := sel.Html()
	if err != nil {
		return nil, fmt.Errorf("%w: read row html: %w", provider.ErrBrokenOffer, err)
	}

	number, err := parseNumberFromDetail(sel)
	if err != nil {
		return nil, err
	}

	if offer.Number.Type == domain.NumberTypeMoto {
		runes := []rune(number)
		// Для мото буквы и цифры при парсинге склиеваются не в том порядке - меняем местами
		letters := runes[0:2]
		digits := runes[2:6]
		reg := runes[6:]

		number = string(digits) + string(letters) + string(reg)
	}

	if number != offer.Number.Number {
		return nil, fmt.Errorf("%w: offer with number %q does not match its number %q", provider.ErrMapOffer, offer.Number.Number, number)
	}

	whereAbouts := parseWhereaboutsFromDetail(sel)
	reissueIncluded := parseReissueFromDetailed(sel)

	var (
		price       *float64
		views       *int
		postedAt    *time.Time
		refreshedAt *time.Time
		parseErr    error
	)

	table := sel.Find(".article__table.user-data-table")
	if table.Length() == 0 {
		return nil, fmt.Errorf("%w: no user-data table", provider.ErrBrokenOffer)
	}
	table.Find("div.user-data-table__tr").Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Find(".user-data-table__th").Text())
		value := s.Find(".user-data-table__td")
		switch title {

		case "Цена авто с номером":
			price, parseErr = parsePrice(value.Text())
			if *price <= 0 {
				price = nil
			}
		case "Просмотров":
			var count int
			count, parseErr = strconv.Atoi(strings.TrimSpace(value.Text()))
			if parseErr == nil {
				views = &count
			}
		case "Дата размещения":
			postedAt, parseErr = parseDateFromDetail(value.Text())
		case "Дата поднятия":
			refreshedAt, parseErr = parseDateFromDetail(value.Text())
		}
	})

	if parseErr != nil {
		return nil, parseErr
	}

	commentText := strings.TrimSpace(sel.Find(".article-comment__content").Text())
	var comment *string = nil
	if len(commentText) != 0 {
		comment = &commentText
	}

	_, err = offer.Offer.ApplyDetail(
		offer.Offer.Status,
		price,
		whereAbouts,
		reissueIncluded,
		views,
		postedAt,
		refreshedAt,
		raw,
		comment,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: read row html: %w", provider.ErrRowSkipped, err)
	}

	return offer, nil
}

func parseNumberFromDetail(sel *goquery.Selection) (string, error) {
	var parts []string

	// Символы номера лежат в value каждого input.filter-plate-number__input, в порядке DOM
	sel.Find(offerDetailNumber).Each(func(i int, s *goquery.Selection) {
		if val, exists := s.Attr("value"); exists {
			parts = append(parts, val)
		}
	})

	region, exists := sel.Find(offerDetailRegion).Attr("value")
	if !exists || len(region) > 3 || len(parts) < 2 {
		return "", fmt.Errorf("%w: invalid offer detail number %q region %q",
			provider.ErrBrokenOffer, strings.Join(parts, ""), region)
	}

	return strings.Join(parts, "") + region, nil
}

func parseWhereaboutsFromDetail(sel *goquery.Selection) *domain.OfferWhereabouts {
	onCar := sel.Find(offerDetailOnCar).Nodes
	if onCar != nil {
		whereAbouts := domain.OfferWhereaboutsOnCar
		return &whereAbouts
	}

	onStorage := sel.Find(offerDetailOnStorage).Nodes
	if onStorage != nil {
		whereAbouts := domain.OfferWhereaboutsOnStorage
		return &whereAbouts
	}

	return nil
}

func parseReissueFromDetailed(sel *goquery.Selection) *bool {
	reissueInclude := sel.Find(offerDetailReissueInclude).Nodes
	if reissueInclude != nil {
		t := true
		return &t
	}

	return nil
}

func parseDateFromDetail(strDate string) (*time.Time, error) {
	months := map[string]string{
		"января":   "January",
		"февраля":  "February",
		"марта":    "March",
		"апреля":   "April",
		"мая":      "May",
		"июня":     "June",
		"июля":     "July",
		"августа":  "August",
		"сентября": "September",
		"октября":  "October",
		"ноября":   "November",
		"декабря":  "December",
	}

	parts := strings.Fields(strDate)
	if len(parts) != 3 {
		return nil, fmt.Errorf("%w: invalid date format", provider.ErrBrokenOffer)
	}

	englishMonth := months[parts[1]]
	if englishMonth == "" {
		return nil, fmt.Errorf("%w: invalid date format", provider.ErrBrokenOffer)
	}

	englishDate := fmt.Sprintf("%s %s %s", parts[0], englishMonth, parts[2])

	parsedTime, err := time.Parse("02 January 2006", englishDate)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid date format", provider.ErrBrokenOffer)
	}

	return &parsedTime, nil
}
