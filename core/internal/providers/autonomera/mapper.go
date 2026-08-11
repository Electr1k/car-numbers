package autonomera

import (
	"core/config"
	"core/internal/domain"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Mapper struct {
	config *config.AutoNomeraConfig
}

func NewMapper(config *config.AutoNomeraConfig) *Mapper {
	return &Mapper{config: config}
}

func (m *Mapper) MapToDomain(string *goquery.Selection) (*domain.Number, *domain.Offer, error) {
	raw, _ := string.Html()

	id, exists := string.Attr("id")
	if !exists || id == "" {
		return nil, nil, fmt.Errorf("missing id attribute")
	}
	id = strings.Replace(id, "item_number_id", "", -1)

	href, exists := string.Attr("href")
	if !exists || href == "" {
		return nil, nil, fmt.Errorf("missing href attribute")
	}
	url := m.config.BaseURL + href

	hrefSplit := strings.Split(href, "/")
	vehicleType := ""
	if len(hrefSplit) >= 2 {
		vehicleType = hrefSplit[1]
	}

	title, exists := string.Attr("title")
	if !exists || title == "" {
		return nil, nil, fmt.Errorf("missing title attribute")
	}

	date := string.Find(".table-date span").Text()
	postedAt, err := time.Parse("02.01.2006", strings.TrimSpace(date))
	if err != nil {
		return nil, nil, fmt.Errorf("parse date '%s': %w", date, err)
	}

	priceText := string.Find(".table-price").Text()
	price, err := m.mapPrice(priceText)
	if err != nil {
		return nil, nil, fmt.Errorf("parse price '%s': %w", priceText, err)
	}

	number, err := domain.NewNumber(
		nil,
		title,
		m.mapType(vehicleType),
		nil,
		nil,
	)

	if err != nil {
		return nil, nil, fmt.Errorf("parse number: %w", err)
	}

	offer, err := domain.NewOffer(
		nil,
		nil,
		domain.AutonomeraProvider,
		id,
		price,
		domain.ACTIVE,
		&postedAt,
		url,
		raw,
		nil,
		nil,
	)

	if err != nil {
		return nil, nil, fmt.Errorf("parse offer: %w", err)
	}

	return number, offer, nil
}

func (m *Mapper) mapPrice(priceText string) (float32, error) {
	priceText = strings.TrimSpace(priceText)
	priceText = strings.ReplaceAll(priceText, "₽", "")
	priceText = strings.ReplaceAll(priceText, "\u00a0", "")
	priceText = strings.ReplaceAll(priceText, " ", "")

	if priceText == "" {
		return 0, fmt.Errorf("empty price")
	}

	price, err := strconv.ParseFloat(priceText, 32)
	if err != nil {
		return 0, err
	}

	return float32(price), nil
}

func (m *Mapper) mapType(vehicleType string) string {
	switch vehicleType {
	case "standart":
		return domain.CarType
	case "moto":
		return domain.MotoType
	case "trailer":
		return domain.TrailerType
	default:
		return domain.CarType
	}
}
