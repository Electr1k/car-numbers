package car_number

import (
	"bytes"
	"core/internal/domain"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type AutoNomeraHttpClient struct {
	baseURL string
	client  *http.Client
}

func AutoNomeraNewClient() *AutoNomeraHttpClient {
	return &AutoNomeraHttpClient{
		baseURL: "https://autonomera777.net/ajax/",
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *AutoNomeraHttpClient) makeRequest(uri string, method string) ([]byte, error) {
	req, err := http.NewRequest(method, c.baseURL+uri, bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	response, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *AutoNomeraHttpClient) FetchNumbers() ([]domain.CarNumber, error) {
	response, err := c.makeRequest("get_numbers.php?order=a.`created`&dir=ASC", "GET")

	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(response)))
	if err != nil {
		return nil, err
	}

	var numbers []domain.CarNumber
	doc.Find("a.table__tr--td").Each(func(i int, s *goquery.Selection) {
		number := domain.CarNumber{}
		if title, exists := s.Attr("title"); exists {
			number.Number = title
		}

		dateText := s.Find(".table-date span").Text()
		postedAt, err := time.Parse("02.01.2006", strings.TrimSpace(dateText))
		if err != nil {
			panic(err)
		}
		number.PostedAt = postedAt

		priceText := s.Find(".table-price").Text()
		priceText = strings.TrimSpace(priceText)
		priceText = strings.ReplaceAll(priceText, "₽", "")
		priceText = strings.ReplaceAll(priceText, "\u00a0", "")
		priceText = strings.ReplaceAll(priceText, " ", "")
		price, err := strconv.ParseFloat(priceText, 32)
		if err != nil {
			fmt.Errorf(err.Error())
			return
		}
		number.Price = float32(price)

		numbers = append(numbers, number)
	})

	return numbers, nil
}
