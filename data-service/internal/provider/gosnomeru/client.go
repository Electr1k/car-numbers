package gosnomeru

import (
	"context"
	"data-service/internal/provider"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// getNumbersPath - постраничная выдача номеров
	getNumbersPath = "/api/v1/plate/search"

	// Порядок выдачи: свежие сверху
	orderColumn = "-update_time"

	// Количество объектов на странице
	perPage = 50

	// defaultTimeout - таймаут по умолчанию
	defaultTimeout = 60 * time.Second

	maxResponseBytes = 8 << 20

	maxErrorPreviewBytes = 512
)

// Client - HTTP-клиент к gosnomeru
type Client struct {
	baseURL string
	http    *http.Client
	logger  *slog.Logger
}

func NewClient(baseURL string, logger *slog.Logger) *Client {
	client := &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: defaultTimeout},
		logger:  logger,
	}

	return client
}

// OffersResponse - Ответ эндпоинта FetchOffers
type OffersResponse struct {
	Items      []OffersItem `json:"items"`
	Total      int          `json:"total"`
	Page       int          `json:"page"`
	PerPage    int          `json:"perPage"`
	TotalPages int          `json:"totalPages"`
}

// OffersItem - Элемент items[] из респонса OffersResponse
type OffersItem struct {
	ID         string       `json:"id"`
	IsSource   bool         `json:"isSource"`
	Number     OffersNumber `json:"number"`
	City       string       `json:"city"`
	RegionName string       `json:"regionName"`
	Price      float64      `json:"price"`
	Date       string       `json:"date"`
	Status     string       `json:"status"`
	Situation  string       `json:"situation"`
	Categories []int        `json:"categories"`
	Slug       string       `json:"slug"`
	Style      int          `json:"style"`
	Role       string       `json:"role"`
	RegionID   int          `json:"regionId"`
	CityID     int          `json:"cityId"`
	OldPrice   int          `json:"oldPrice"`
	ChainCount int          `json:"chainCount"`
}

// OffersNumber Номер автомобиля из респонсе OffersResponse
type OffersNumber struct {
	Letters string `json:"letters"`
	Region  string `json:"region"`
}

// FetchOffers забирает одну страницу предложений и отдаёт сырой объект json
func (c *Client) FetchOffers(ctx context.Context, page int) (*OffersResponse, error) {
	requestURL := c.buildURL(page)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	c.logger.Debug("fetching page", "url", requestURL)

	response, err := c.http.Do(request)
	if err != nil {
		return nil, fmt.Errorf("%w: get %s: %w", provider.ErrProviderUnavailable, requestURL, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, processBadStatus(response, requestURL)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("%w: read body from %s: %w", provider.ErrInvalidResponse, requestURL, err)
	}

	var jsonResponse OffersResponse
	if err := json.Unmarshal(body, &jsonResponse); err != nil {
		return nil, fmt.Errorf("%w: parse json from %s: %w", provider.ErrInvalidResponse, requestURL, err)
	}

	c.logger.Debug("page fetched",
		"page", page,
		"per_page", perPage,
		"status_code", response.StatusCode,
		"bytes", len(body))

	return &jsonResponse, nil
}

//// FetchOfferDetailHTML забирает одну страницу предложений и отдаёт сырой HTML
//func (c *Client) FetchOfferDetailHTML(ctx context.Context, url string) ([]byte, error) {
//	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
//	if err != nil {
//		return nil, fmt.Errorf("create request: %w", err)
//	}
//
//	c.logger.Debug("fetching offer detail page", "url", url)
//
//	response, err := c.http.Do(request)
//	if err != nil {
//		return nil, fmt.Errorf("%w: get %s: %w", provider.ErrProviderUnavailable, url, err)
//	}
//	defer response.Body.Close()
//
//	if response.StatusCode != http.StatusOK {
//		return nil, processBadStatus(response, url)
//	}
//
//	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes))
//	if err != nil {
//		return nil, fmt.Errorf("%w: read body from %s: %w", provider.ErrInvalidResponse, url, err)
//	}
//
//	c.logger.Debug("offer detail fetched",
//		"url", url,
//		"status_code", response.StatusCode,
//		"bytes", len(body))
//
//	return body, nil
//}

func (c *Client) buildURL(page int) string {
	query := url.Values{}
	query.Set("sort", orderColumn)
	query.Set("per_page", strconv.Itoa(perPage))
	query.Set("page", strconv.Itoa(page))

	return c.baseURL + getNumbersPath + "?" + query.Encode()
}

// processBadStatus - Процессинг ошибки по статус коду
func processBadStatus(response *http.Response, requestURL string) error {
	statesError := provider.ErrInvalidResponse

	switch {
	case response.StatusCode == http.StatusTooManyRequests:
		statesError = provider.ErrRateLimitExceeded
	case response.StatusCode == http.StatusNotFound:
		statesError = provider.ErrNotFound
	case response.StatusCode >= http.StatusInternalServerError:
		statesError = provider.ErrProviderUnavailable
	}

	var responsePart string
	preview, err := io.ReadAll(io.LimitReader(response.Body, maxErrorPreviewBytes))
	if err != nil || len(preview) == 0 {
		responsePart = "Body is empty"
	} else {
		responsePart = strings.Join(strings.Fields(string(preview)), " ")
	}

	return fmt.Errorf("%w: got %d from %s: %s", statesError, response.StatusCode, requestURL, responsePart)
}
