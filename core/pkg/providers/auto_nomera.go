package providers

import (
	"bytes"
	"core/internal/domain"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	defaultBaseURL = "https://autonomera777.net/ajax/"
	defaultTimeout = 60 * time.Second
)

type AutoNomeraHttpClient struct {
	baseURL string
	client  *http.Client
}

type RequestParams struct {
	Method  string
	Path    string
	Query   map[string]string
	Body    interface{}
	Headers map[string]string
}

func AutoNomeraNewClient() *AutoNomeraHttpClient {
	return &AutoNomeraHttpClient{
		baseURL: "https://autonomera777.net/ajax/",
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *AutoNomeraHttpClient) makeRequest(params RequestParams) ([]byte, error) {
	fullURL := c.baseURL + params.Path
	if len(params.Query) > 0 {
		fullURL += "?" + c.buildQueryString(params.Query)
	}

	var bodyReader io.Reader
	if params.Body != nil {
		jsonBody, err := json.Marshal(params.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(params.Method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setDefaultHeaders(req)
	c.setCustomHeaders(req, params.Headers)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

func (c *AutoNomeraHttpClient) FetchNumbers(start int) ([]domain.CarNumber, error) {

	params := RequestParams{
		Method: http.MethodGet,
		Path:   "get_numbers.php",
		Query: map[string]string{
			"order": "a.`created`",
			"dir":   "DESC",
			"start": strconv.Itoa(start),
		},
	}

	response, err := c.makeRequest(params)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch numbers with options: %w", err)
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
			return
		}

		number.Price = float32(price)

		if number.Validate() != nil {
			return
		}

		numbers = append(numbers, number)
	})

	return numbers, nil
}

func (c *AutoNomeraHttpClient) buildQueryString(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}

	var queryParts []string
	for key, value := range params {
		queryParts = append(queryParts, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.Join(queryParts, "&")
}

func (c *AutoNomeraHttpClient) setDefaultHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
}

// setCustomHeaders устанавливает пользовательские заголовки
func (c *AutoNomeraHttpClient) setCustomHeaders(req *http.Request, headers map[string]string) {
	for key, value := range headers {
		req.Header.Set(key, value)
	}
}
