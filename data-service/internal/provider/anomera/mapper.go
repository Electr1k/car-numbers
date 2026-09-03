package anomera

import (
	"data-service/internal/domain"
	"data-service/internal/provider"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

const (
	// offerRowSelector - строка таблицы с одним предложением
	offerRowSelector = "table.main-table tbody tr:has(td.car-img)"

	// numberSelector - мета с номером вида "А123АА 77"
	numberSelector = `meta[itemprop="name"]`

	// offerLinkSelector - относительная ссылка на страницу предложения
	offerLinkSelector = "td.car-img a"

	// listPriceSelector - мета с ценой в рублях
	listPriceSelector = `meta[itemprop="price"]`

	// listPriceCellSelector - ячейка цены: у договорных предложений вместо микроразметки в ней текст
	listPriceCellSelector = "td.car-price"

	// availabilitySelector - ссылка на статус наличия
	availabilitySelector = `link[itemprop="availability"]`

	// dateSelector - ячейка с датой публикации
	dateSelector = ".date__text"

	dateLayout = "02.01.06"

	// Для закрепленных объявлений дата не указывается - подставляем текущую
	dateInvisible = "закреплено"

	// availability - статусы наличия
	availabilityInStock      = "https://schema.org/InStock"
	availabilityOutOfStock   = "https://schema.org/OutOfStock"
	availabilitySoldOut      = "https://schema.org/SoldOut"
	availabilityDiscontinued = "https://schema.org/Discontinued"

	plateLength = 6

	numberMaskChar = '*'

	// productSelector - корневой блок карточки предложения
	productSelector = `div.product`

	// offerScopeSelector - вложенная schema.org-область с ценой и наличием
	offerScopeSelector = `[itemprop="offers"]`

	// detailPriceSelector - цена в карточке
	detailPriceSelector = ".product-info__item-price"

	// canonicalURLSelector - канонический адрес карточки
	canonicalURLSelector = `meta[property="og:url"]`

	// negotiablePrice - цену продавец не указал
	negotiablePrice = "договорная"

	// plateImageSelector - картинка номера: её alt содержит номер целиком, вместе с маской
	plateImageSelector = "img.img-view-grz"

	// commentSelector - описание предложения от продавца
	commentSelector = `[itemprop="description"]`

	// viewCountSelector - счётчик просмотров карточки
	viewCountSelector = ".views-count"

	// infoItemSelector - строка таблицы характеристик карточки
	infoItemSelector = ".product-info__item"

	// infoQuestionSelector - название характеристики
	infoQuestionSelector = "p.question"

	// infoAnswerSelector - значение характеристики
	infoAnswerSelector = "p.answer"

	// infoQuestion* - названия характеристик карточки
	infoQuestionPostedAt    = "Размещен"
	infoQuestionReissue     = "Переоформление"
	infoQuestionWhereabouts = "Наличие"

	// whereabouts* - значения характеристики "Наличие"
	whereaboutsOnCar     = "Номер на автомобиле"
	whereaboutsOnStorage = "Номер на хранении"
	whereaboutsUnknown   = "Не указано"

	// reissue* - значения характеристики "Переоформление"
	reissueIncludedText = "Включен в стоимость"
	reissueSeparateText = "Оплачивается отдельно"

	// motoURLPrefix, trailerURLPrefix - префиксы слага деталки: у мото и прицепов свои разделы и своя нумерация
	motoURLPrefix    = "moto-"
	trailerURLPrefix = "pricep-"
)

// offerSlugPattern - слаг деталки
var offerSlugPattern = regexp.MustCompile(`^(` + motoURLPrefix + `|` + trailerURLPrefix + `)?(\d+)-([a-z0-9]+)\.html$`)

// sectionNumberTypes - тип ТС по префиксу раздела в слаге
var sectionNumberTypes = map[string]domain.NumberType{
	motoURLPrefix:    domain.NumberTypeMoto,
	trailerURLPrefix: domain.NumberTypeTrailer,
}

// plateCaptionPattern - номер с регионом в конце подписи картинки вида "Красивый номер на авто О999У* 126"
var plateCaptionPattern = regexp.MustCompile(`([А-Я\d*]{6})\s*([\d*]{2,3})$`)

// plateImageNumberAttrs - подписи картинки номера: в них номер есть целиком, вместе с маской
var plateImageNumberAttrs = []string{"title", "alt"}

// plateLetters - обратная транслитерация: в адресах провайдер пишет номер латиницей по этой карте
var plateLetters = map[rune]rune{
	'a': 'А', 'b': 'В', 'e': 'Е', 'k': 'К', 'm': 'М', 'h': 'Н',
	'o': 'О', 'p': 'Р', 'c': 'С', 't': 'Т', 'y': 'У', 'x': 'Х',
}

type Mapper struct {
	baseURL string
}

// offerRef - всё, что известно о предложении из его адреса
type offerRef struct {
	url string

	externalID string

	number string

	numberType domain.NumberType
}

// parseOfferRef - адрес, идентификатор и номер предложения из ссылки на карточку
func (m *Mapper) parseOfferRef(href string) (offerRef, error) {
	parsed, err := url.Parse(href)
	if err != nil || !strings.HasPrefix(parsed.Path, "/") {
		return offerRef{}, fmt.Errorf("%w: unexpected offer href %q", provider.ErrBrokenOffer, href)
	}

	groups := offerSlugPattern.FindStringSubmatch(path.Base(parsed.Path))
	if groups == nil {
		return offerRef{}, fmt.Errorf("%w: unexpected offer slug in %q", provider.ErrBrokenOffer, href)
	}

	section, externalID, plate := groups[1], groups[2], groups[3]

	return offerRef{
		url:        m.baseURL + parsed.Path,
		externalID: section + externalID,
		number:     transliteratePlate(plate),
		numberType: sectionNumberTypes[section],
	}, nil
}

// transliteratePlate - номер кириллицей из слага
func transliteratePlate(plate string) string {
	var restored strings.Builder

	for _, symbol := range plate {
		if letter, ok := plateLetters[symbol]; ok {
			restored.WriteRune(letter)
			continue
		}

		restored.WriteRune(symbol)
	}

	return restored.String()
}

func NewMapper(baseURL string) *Mapper {
	return &Mapper{baseURL: strings.TrimRight(baseURL, "/")}
}

// MapOfferToDomain - Маппит строку выдачи в домен
func (m *Mapper) MapOfferToDomain(sel *goquery.Selection) (domain.OfferWithNumber, error) {
	var empty domain.OfferWithNumber

	raw, err := goquery.OuterHtml(sel)
	if err != nil {
		return empty, fmt.Errorf("%w: read row html: %w", provider.ErrBrokenOffer, err)
	}

	href, err := requiredAttr(sel, offerLinkSelector, "href")
	if err != nil {
		return empty, err
	}

	ref, err := m.parseOfferRef(href)
	if err != nil {
		return empty, err
	}

	negotiable, err := isNegotiable(sel, listPriceSelector, listPriceCellSelector)
	if err != nil {
		return empty, err
	}

	numberStr, numberType, err := parseRowNumber(sel, ref, negotiable)
	if err != nil {
		return empty, err
	}

	postedAt, err := parsePostedAt(sel)
	if err != nil {
		return empty, err
	}

	price, status, err := parsePriceStatus(sel, negotiable)
	if err != nil {
		return empty, err
	}

	number, err := domain.NewNumber(numberStr, numberType)
	if err != nil {
		return empty, fmt.Errorf("%w: invalid number %q: %w", provider.ErrRowSkipped, numberStr, err)
	}

	offer, err := domain.NewOffer(
		number.Id,
		domain.ProviderAnomera,
		ref.externalID,
		price,
		status,
		nil,
		nil,
		nil,
		&postedAt,
		&postedAt,
		ref.url,
		raw,
		nil,
		nil,
	)
	if err != nil {
		return empty, fmt.Errorf("%w: invalid offer %q: %w", provider.ErrRowSkipped, ref.externalID, err)
	}

	return domain.OfferWithNumber{Number: number, Offer: offer}, nil
}

// ApplyOfferDetailToDomain - Дополняет предложение из выдачи данными его карточки
func (m *Mapper) ApplyOfferDetailToDomain(sel *goquery.Selection, offer domain.OfferWithNumber) (domain.OfferWithNumber, error) {
	var emptyOffer domain.OfferWithNumber

	newOffer, err := m.MapOfferDetailToDomain(sel)
	if err != nil {
		return emptyOffer, err
	}

	if newOffer.Number.Number != offer.Number.Number {
		return emptyOffer, fmt.Errorf("%w: offer with number %q does not match its number %q", provider.ErrMapOffer, offer.Number.Number, newOffer.Number.Number)
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

// requiredAttr - обязательный атрибут первого узла, найденного по селектору
func requiredAttr(sel *goquery.Selection, selector string, name string) (string, error) {
	value, exists := sel.Find(selector).First().Attr(name)
	if !exists || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("%w: missing %s attribute in %q", provider.ErrBrokenOffer, name, selector)
	}

	return strings.TrimSpace(value), nil
}

// requiredChildAttr - обязательный атрибут прямого потомка: одинаковый itemprop есть и у вложенных областей
func requiredChildAttr(sel *goquery.Selection, selector string, name string) (string, error) {
	value, exists := sel.ChildrenFiltered(selector).First().Attr(name)
	if !exists || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("%w: missing %s attribute in child %q", provider.ErrBrokenOffer, name, selector)
	}

	return strings.TrimSpace(value), nil
}

// parseNumber - номер и тип ТС из заголовка вида "А123АА 77"
func parseNumber(title string) (string, domain.NumberType, error) {
	number := strings.ToUpper(strings.Join(strings.Fields(title), ""))

	runes := []rune(number)
	if len(runes) < plateLength+2 || len(runes) > plateLength+3 {
		return "", "", fmt.Errorf("%w: unexpected number %q", provider.ErrBrokenOffer, title)
	}

	if !isDigits(runes[plateLength:]) {
		return "", "", fmt.Errorf("%w: unexpected region in number %q", provider.ErrBrokenOffer, title)
	}

	numberType, err := parseNumberType(runes[:plateLength])
	if err != nil {
		return "", "", fmt.Errorf("%w: unknown vehicle type in number %q", provider.ErrBrokenOffer, title)
	}

	return number, numberType, nil
}

// parseNumberType - тип ТС по раскладке букв и цифр в номере
func parseNumberType(plate []rune) (domain.NumberType, error) {
	switch {
	case isLetters(plate[0:1]) && isDigits(plate[1:4]) && isLetters(plate[4:6]):
		return domain.NumberTypeCar, nil
	case isDigits(plate[0:4]) && isLetters(plate[4:6]):
		return domain.NumberTypeMoto, nil
	case isLetters(plate[0:2]) && isDigits(plate[2:6]):
		return domain.NumberTypeTrailer, nil
	default:
		return "", fmt.Errorf("unknown vehicle type")
	}
}

// isLetters - все символы буквы или маска: провайдер прячет часть символов у ~5% номеров
func isLetters(runes []rune) bool {
	for _, r := range runes {
		if !unicode.IsLetter(r) && r != numberMaskChar {
			return false
		}
	}

	return len(runes) > 0
}

// isDigits - все символы цифры или маска
func isDigits(runes []rune) bool {
	for _, r := range runes {
		if !unicode.IsDigit(r) && r != numberMaskChar {
			return false
		}
	}

	return len(runes) > 0
}

// isNegotiable - предложение без цены
func isNegotiable(sel *goquery.Selection, priceSelector string, priceCellSelector string) (bool, error) {
	if sel.Find(priceSelector).Length() > 0 {
		return false, nil
	}

	price := strings.TrimSpace(sel.Find(priceCellSelector).First().Text())
	if !strings.EqualFold(price, negotiablePrice) {
		return false, fmt.Errorf("%w: no price microdata, got %q instead of %q", provider.ErrBrokenOffer, price, negotiablePrice)
	}

	return true, nil
}

// parseRowNumber - номер и тип ТС строки выдачи
func parseRowNumber(sel *goquery.Selection, ref offerRef, negotiable bool) (string, domain.NumberType, error) {
	title, err := requiredAttr(sel, numberSelector, "content")
	if err != nil {
		if !negotiable {
			return "", "", err
		}

		// Микроразметки нет - остаётся слаг, но маскированные символы провайдер из него выбрасывает
		if strings.ContainsRune(sel.Find(offerLinkSelector).Text(), numberMaskChar) {
			return "", "", fmt.Errorf("%w: masked number is not restorable from %q", provider.ErrRowSkipped, ref.url)
		}

		title = ref.number
	}

	numberStr, numberType, err := parseNumber(title)
	if err != nil {
		return "", "", err
	}

	// Раздел в адресе надёжнее раскладки номера
	if ref.numberType != "" {
		numberType = ref.numberType
	}

	return numberStr, numberType, nil
}

// parsePriceStatus - цена и статус предложения
func parsePriceStatus(sel *goquery.Selection, negotiable bool) (*float64, domain.OfferStatus, error) {
	if negotiable {
		return nil, domain.OfferStatusActive, nil
	}

	price, err := parseListPrice(sel)
	if err != nil {
		return nil, "", err
	}

	status, err := parseStatus(sel)
	if err != nil {
		return nil, "", err
	}

	return price, status, nil
}

// parsePostedAt - дата публикации
func parsePostedAt(sel *goquery.Selection) (time.Time, error) {
	date := strings.TrimSpace(sel.Find(dateSelector).First().Text())
	if date == "" {
		return time.Time{}, fmt.Errorf("%w: empty date cell", provider.ErrBrokenOffer)
	}

	if date == dateInvisible {
		return time.Now(), nil
	}
	postedAt, err := time.Parse(dateLayout, date)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: parse date %q: %w", provider.ErrBrokenOffer, date, err)
	}

	return postedAt, nil
}

// parseListPrice - цена в рублях из меты предложения
func parseListPrice(sel *goquery.Selection) (*float64, error) {
	content, err := requiredAttr(sel, listPriceSelector, "content")
	if err != nil {
		return nil, err
	}

	price, err := strconv.ParseFloat(content, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: price %q is not a number", provider.ErrRowSkipped, content)
	}

	// Домен не принимает неположительную цену, а провайдер её не обещает
	if price <= 0 {
		return nil, nil
	}

	return &price, nil
}

// parseStatus - статус предложения по schema.org-наличию
func parseStatus(sel *goquery.Selection) (domain.OfferStatus, error) {
	availability, err := requiredAttr(sel, availabilitySelector, "href")
	if err != nil {
		return "", err
	}

	switch availability {
	case availabilityInStock:
		return domain.OfferStatusActive, nil
	case availabilitySoldOut:
		return domain.OfferStatusSold, nil
	case availabilityOutOfStock, availabilityDiscontinued:
		return domain.OfferStatusInactive, nil
	default:
		return "", fmt.Errorf("%w: unknown availability %q", provider.ErrBrokenOffer, availability)
	}
}

// MapOfferDetailToDomain - Маппит карточку предложения в домен
func (m *Mapper) MapOfferDetailToDomain(sel *goquery.Selection) (domain.OfferWithNumber, error) {
	var empty domain.OfferWithNumber

	product := sel.Find(productSelector).First()
	if product.Length() == 0 {
		return empty, fmt.Errorf("%w: no product block on detail page", provider.ErrNotFound)
	}

	raw, err := goquery.OuterHtml(product)
	if err != nil {
		return empty, fmt.Errorf("%w: read detail html: %w", provider.ErrBrokenOffer, err)
	}

	href, err := requiredAttr(sel, canonicalURLSelector, "content")
	if err != nil {
		return empty, err
	}

	ref, err := m.parseOfferRef(href)
	if err != nil {
		return empty, err
	}

	negotiable, err := isNegotiable(product, offerScopeSelector, detailPriceSelector)
	if err != nil {
		return empty, err
	}

	numberStr, numberType, err := parseDetailNumber(product, ref, negotiable)
	if err != nil {
		return empty, err
	}

	price, status, err := parsePriceStatus(product.Find(offerScopeSelector).First(), negotiable)
	if err != nil {
		return empty, err
	}

	info := parseInfoRows(product)

	postedAt, err := parseDetailPostedAt(info[infoQuestionPostedAt])
	if err != nil {
		return empty, err
	}

	whereabouts, err := parseWhereabouts(info[infoQuestionWhereabouts])
	if err != nil {
		return empty, err
	}

	reissueIncluded, err := parseReissueIncluded(info[infoQuestionReissue])
	if err != nil {
		return empty, err
	}

	viewCount, err := parseViewCount(product)
	if err != nil {
		return empty, err
	}

	number, err := domain.NewNumber(numberStr, numberType)
	if err != nil {
		return empty, fmt.Errorf("%w: invalid number %q: %w", provider.ErrRowSkipped, numberStr, err)
	}

	offer, err := domain.NewOffer(
		number.Id,
		domain.ProviderAnomera,
		ref.externalID,
		price,
		status,
		whereabouts,
		reissueIncluded,
		viewCount,
		&postedAt,
		&postedAt,
		ref.url,
		raw,
		&raw,
		parseComment(product),
	)
	if err != nil {
		return empty, fmt.Errorf("%w: invalid offer %q: %w", provider.ErrRowSkipped, ref.externalID, err)
	}

	return domain.OfferWithNumber{Number: number, Offer: offer}, nil
}

// parseDetailNumber - номер и тип ТС карточки
func parseDetailNumber(product *goquery.Selection, ref offerRef, negotiable bool) (string, domain.NumberType, error) {
	title, err := requiredChildAttr(product, numberSelector, "content")
	if err != nil {
		if !negotiable {
			return "", "", err
		}

		title, err = parseDetailNumberFallback(product, ref)
		if err != nil {
			return "", "", err
		}
	}

	numberStr, numberType, err := parseNumber(title)
	if err != nil {
		return "", "", err
	}

	// Раздел в адресе надёжнее раскладки номера
	if ref.numberType != "" {
		numberType = ref.numberType
	}

	return numberStr, numberType, nil
}

// parseDetailNumberFallback - получение номера из названия к картинке (фоллбек)
func parseDetailNumberFallback(product *goquery.Selection, ref offerRef) (string, error) {
	image := product.Find(plateImageSelector).First()

	for _, attr := range plateImageNumberAttrs {
		if number := parsePlateCaption(image.AttrOr(attr, "")); number != "" {
			return number, nil
		}
	}

	// Остаётся слаг, но маскированные символы провайдер из него выбрасывает
	if strings.ContainsRune(image.AttrOr("alt", ""), numberMaskChar) {
		return "", fmt.Errorf("%w: masked number is not restorable from %q", provider.ErrRowSkipped, ref.url)
	}

	return ref.number, nil
}

// parsePlateCaption - номер из подписи картинки, пустая строка если номера в ней нет
func parsePlateCaption(caption string) string {
	groups := plateCaptionPattern.FindStringSubmatch(strings.ToUpper(strings.TrimSpace(caption)))
	if groups == nil {
		return ""
	}

	return groups[1] + " " + groups[2]
}

// parseInfoRows - характеристики карточки в виде вопрос-ответ
func parseInfoRows(product *goquery.Selection) map[string]string {
	rows := make(map[string]string)

	product.Find(infoItemSelector).Each(func(_ int, item *goquery.Selection) {
		question := strings.TrimSpace(item.Find(infoQuestionSelector).First().Text())
		if question == "" {
			return
		}

		rows[question] = strings.TrimSpace(item.Find(infoAnswerSelector).First().Text())
	})

	return rows
}

// parseDetailPostedAt - дата размещения из характеристик карточки
func parseDetailPostedAt(date string) (time.Time, error) {
	if date == "" {
		return time.Time{}, fmt.Errorf("%w: missing %q on detail page", provider.ErrBrokenOffer, infoQuestionPostedAt)
	}

	postedAt, err := time.Parse(dateLayout, date)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: parse date %q: %w", provider.ErrBrokenOffer, date, err)
	}

	return postedAt, nil
}

// parseWhereabouts - где находится номер
func parseWhereabouts(value string) (*domain.OfferWhereabouts, error) {
	switch value {
	case "", whereaboutsUnknown:
		return nil, nil
	case whereaboutsOnCar:
		whereabouts := domain.OfferWhereaboutsOnCar
		return &whereabouts, nil
	case whereaboutsOnStorage:
		whereabouts := domain.OfferWhereaboutsOnStorage
		return &whereabouts, nil
	default:
		return nil, fmt.Errorf("%w: unknown whereabouts %q", provider.ErrBrokenOffer, value)
	}
}

// parseReissueIncluded - входит ли переоформление в стоимость
func parseReissueIncluded(value string) (*bool, error) {
	switch value {
	case "":
		return nil, nil
	case reissueIncludedText:
		included := true
		return &included, nil
	case reissueSeparateText:
		included := false
		return &included, nil
	default:
		return nil, fmt.Errorf("%w: unknown reissue %q", provider.ErrBrokenOffer, value)
	}
}

// parseViewCount - счётчик просмотров, у мото и прицепов его может не быть
func parseViewCount(product *goquery.Selection) (*int, error) {
	counter := product.Find(viewCountSelector).First()
	if counter.Length() == 0 {
		return nil, nil
	}

	text := strings.TrimSpace(counter.Text())
	count, err := strconv.Atoi(text)
	if err != nil {
		return nil, fmt.Errorf("%w: view count %q is not a number", provider.ErrBrokenOffer, text)
	}

	return &count, nil
}

// parseComment - описание предложения от продавца
func parseComment(product *goquery.Selection) *string {
	comment := strings.TrimSpace(product.Find(commentSelector).First().Text())
	if comment == "" {
		return nil
	}

	return &comment
}
