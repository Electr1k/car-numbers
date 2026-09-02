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
	// getNumbers - постраничная выдача номеров
	getNumbers = "/api/v1/plate/search"

	// getNumberDetail - Деталка предложения
	getNumberDetail = "/api/v1/plate/"

	// getLatestNumbers - Последние новые предложения
	getLatestNumbers = "/api/v1/plate/latest"

	// Порядок выдачи: свежие сверху
	orderColumn = "-update_time"

	// Количество объектов на странице
	perPage = 50

	rateLimitHeader = "X-Ratelimit-Remaining"

	// TODO: из расчета на 8 воркеров = 8 * 1 * 60 = 480 RpM (лимит 500)
	rateLimitTimeout = time.Second

	// defaultTimeout - таймаут по умолчанию
	defaultTimeout = 60 * time.Second

	maxResponseBytes = 8 << 20

	maxErrorPreviewBytes = 512

	latestNumberLimit = 20
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

func (c *Client) request(ctx context.Context, method string, url string) ([]byte, error) {
	c.logger.Debug("request", "url", url, "method", method)

	// Спим для рейтлимитов
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(rateLimitTimeout):
	}

	request, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	response, err := c.http.Do(request)
	if err != nil {
		return nil, fmt.Errorf("%w: %s %s: %w", provider.ErrProviderUnavailable, method, url, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, processBadStatus(response, url)
	}

	responseByte, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("%w: read body from %s: %w", provider.ErrInvalidResponse, url, err)
	}
	c.logger.Debug("response",
		"url", url,
		"method", method,
		"status", response.StatusCode,
		"rate_limit", response.Header.Get(rateLimitHeader),
		"body", string(responseByte),
	)

	return responseByte, nil
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

	response, err := c.request(ctx, http.MethodGet, requestURL)
	if err != nil {
		return nil, err
	}

	var jsonResponse OffersResponse
	if err := json.Unmarshal(response, &jsonResponse); err != nil {
		return nil, fmt.Errorf("%w: parse json from %s: %w", provider.ErrInvalidResponse, requestURL, err)
	}

	return &jsonResponse, nil
}

type OfferDetail struct {
	ID           string       `json:"id"`
	IsSource     bool         `json:"isSource"`
	Number       OffersNumber `json:"number"`
	City         string       `json:"city"`
	RegionName   string       `json:"regionName"`
	Price        float64      `json:"price"`
	Date         string       `json:"date"`
	Status       string       `json:"status"`
	Availability *string      `json:"availability"`
	Situation    string       `json:"situation"`
	Categories   []int        `json:"categories"`
	Slug         string       `json:"slug"`
	Style        int          `json:"style"`
	Role         string       `json:"role"`
	RegionID     int          `json:"regionId"`
	CityID       int          `json:"cityId"`
	Description  string       `json:"description"`
	SellerName   string       `json:"sellerName"`
	SellerPhone  *string      `json:"sellerPhone"`
	NeedCar      bool         `json:"needCar"`
	DealIncluded bool         `json:"dealIncluded"`
	ViewCount    int          `json:"viewCount"`
	CreatedAt    string       `json:"createdAt"`
}

// FetchOfferDetail забирает деталку предложения
func (c *Client) FetchOfferDetail(ctx context.Context, externalId string) (*OfferDetail, error) {
	requestUrl := c.baseURL + getNumberDetail + externalId

	response, err := c.request(ctx, http.MethodGet, requestUrl)
	if err != nil {
		return nil, err
	}

	var jsonResponse OfferDetail
	if err := json.Unmarshal(response, &jsonResponse); err != nil {
		return nil, fmt.Errorf("%w: parse json from %s: %w", provider.ErrInvalidResponse, requestUrl, err)
	}

	return &jsonResponse, nil
}

// LatestOffersResponse - Ответ эндпоинта FetchLatestOffers
type LatestOffersResponse struct {
	Items []OffersItem `json:"items"`
}

// FetchLatestOffers забирает последние созданные предложения
func (c *Client) FetchLatestOffers(ctx context.Context) (*LatestOffersResponse, error) {
	requestUrl := c.baseURL + getLatestNumbers + "?limit=" + strconv.Itoa(latestNumberLimit)

	response, err := c.request(ctx, http.MethodGet, requestUrl)
	if err != nil {
		return nil, err
	}

	var jsonResponse LatestOffersResponse
	if err := json.Unmarshal(response, &jsonResponse); err != nil {
		return nil, fmt.Errorf("%w: parse json from %s: %w", provider.ErrInvalidResponse, requestUrl, err)
	}

	return &jsonResponse, nil

}

func (c *Client) buildURL(page int) string {
	query := url.Values{}
	query.Set("sort", orderColumn)
	query.Set("per_page", strconv.Itoa(perPage))
	query.Set("page", strconv.Itoa(page))

	return c.baseURL + getNumbers + "?" + query.Encode()
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
