package autonomera

import (
	"core/internal/domain"
	"core/internal/provider"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(html []byte) ([]domain.CarNumber, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(html)))
	if err != nil {
		return nil, fmt.Errorf("parse html document: %w", err)
	}

	var numbers []domain.CarNumber
	var parseErrors []error

	doc.Find("a.table__tr--td").Each(func(i int, s *goquery.Selection) {
		number, err := p.parseCarNumber(s)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Errorf("row %d: %w", i, err))
			return
		}

		if err := number.Validate(); err != nil {
			parseErrors = append(parseErrors, fmt.Errorf("row %d validation: %w", i, err))
			return
		}

		numbers = append(numbers, *number)
	})

	if len(parseErrors) > 0 && len(numbers) == 0 {
		return nil, fmt.Errorf("%w: all rows failed to parse", provider.ErrParsingFailed)
	}

	return numbers, nil
}

func (p *Parser) parseCarNumber(s *goquery.Selection) (*domain.CarNumber, error) {
	number := &domain.CarNumber{}

	title, exists := s.Attr("title")
	if !exists || title == "" {
		return nil, fmt.Errorf("missing title attribute")
	}
	number.Number = title

	dateText := s.Find(".table-date span").Text()
	postedAt, err := time.Parse("02.01.2006", strings.TrimSpace(dateText))
	if err != nil {
		return nil, fmt.Errorf("parse date '%s': %w", dateText, err)
	}
	number.PostedAt = postedAt

	priceText := s.Find(".table-price").Text()
	price, err := p.parsePrice(priceText)
	if err != nil {
		return nil, fmt.Errorf("parse price '%s': %w", priceText, err)
	}
	number.Price = price

	return number, nil
}

func (p *Parser) parsePrice(priceText string) (float32, error) {
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
