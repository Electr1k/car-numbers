package plate

import (
	"context"
	"core-service/config"
	"core-service/internal/service"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	fetchPlates  = "/api/v1/numbers"
	fetchRegions = "/api/v1/regions"

	maxResponseSize     = 8 << 20
	maxIdleConns        = 100
	maxIdleConnsPerHost = 100
)

// Client - HTTP-клиент к plate-service
type Client struct {
	baseURL string
	http    *http.Client
	logger  *slog.Logger
}

func NewClient(cfg config.PlateConfig, logger *slog.Logger) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = maxIdleConns
	transport.MaxIdleConnsPerHost = maxIdleConnsPerHost

	client := &Client{
		baseURL: strings.TrimRight(cfg.URL, "/"),
		http: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: transport,
		},
		logger: logger,
	}

	return client
}

func (c *Client) request(ctx context.Context, method string, requestURL string) ([]byte, error) {
	c.logger.DebugContext(ctx, "request", "url", requestURL, "method", method)

	request, err := http.NewRequestWithContext(ctx, method, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	response, err := c.http.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request error: %s %s: %w", method, requestURL, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, processBadStatus(response, requestURL)
	}

	responseByte, err := io.ReadAll(io.LimitReader(response.Body, maxResponseSize+1))
	if err != nil {
		return nil, fmt.Errorf("read body from %s: %w", requestURL, err)
	}
	if len(responseByte) > maxResponseSize {
		return nil, fmt.Errorf("%w: response from %s exceeds %d bytes", service.ErrServiceUnavailable, requestURL, maxResponseSize)
	}

	c.logger.DebugContext(ctx, "response",
		"url", requestURL,
		"method", method,
		"status", response.StatusCode,
		"bytes", len(responseByte),
	)

	return responseByte, nil
}

type FetchPlatesResponse struct {
	Items      []FetchPlatesItem `json:"items"`
	NextCursor *string           `json:"next_cursor"`
}

type FetchPlatesItem struct {
	ID              uuid.UUID          `json:"id"`
	Number          string             `json:"number"`
	Region          *FetchPlatesRegion `json:"region"`
	Price           *float64           `json:"price"`
	Type            string             `json:"type"`
	Count           int                `json:"count"`
	RefreshedAt     time.Time          `json:"refreshed_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	ReissueIncluded *bool              `json:"reissue_included"`
}

type FetchPlatesRegion struct {
	ID   int    `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type FetchPlatesParams struct {
	Query           *string
	RegionID        *int
	PriceFrom       *float64
	PriceTo         *float64
	ReissueIncluded *bool
	CategoryIDs     []int
	Limit           int
	Cursor          string
}

// FetchPlates возвращает список номеров
func (c *Client) FetchPlates(ctx context.Context, params FetchPlatesParams) (*FetchPlatesResponse, error) {
	requestURL := c.buildFetchPlatesURL(params)

	response, err := c.request(ctx, http.MethodGet, requestURL)
	if err != nil {
		return nil, err
	}

	var jsonResponse FetchPlatesResponse
	if err := json.Unmarshal(response, &jsonResponse); err != nil {
		return nil, fmt.Errorf("parse json from %s: %w", requestURL, err)
	}

	return &jsonResponse, nil
}

type FetchRegionsResponse struct {
	Items []FetchRegionsItem `json:"items"`
}

type FetchRegionsItem struct {
	ID    int      `json:"id"`
	Name  string   `json:"name"`
	Codes []string `json:"codes"`
}

// FetchRegions возвращает список регионов с кодами
func (c *Client) FetchRegions(ctx context.Context) (*FetchRegionsResponse, error) {
	requestURL := c.baseURL + fetchRegions

	response, err := c.request(ctx, http.MethodGet, requestURL)
	if err != nil {
		return nil, err
	}

	var jsonResponse FetchRegionsResponse
	if err := json.Unmarshal(response, &jsonResponse); err != nil {
		return nil, fmt.Errorf("parse json from %s: %w", requestURL, err)
	}

	return &jsonResponse, nil
}

type FetchPlateByIDResponse struct {
	ID     uuid.UUID             `json:"id"`
	Number string                `json:"number"`
	Region *FetchPlatesRegion    `json:"region"`
	Offers []FetchPlateByIDOffer `json:"offers"`
}

type FetchPlateByIDOffer struct {
	ID              uuid.UUID `json:"id"`
	Provider        string    `json:"provider"`
	Price           *float64  `json:"price"`
	Status          string    `json:"status"`
	ReissueIncluded *bool     `json:"reissue_included"`
	Whereabouts     *string   `json:"whereabouts"`
	ViewCount       *int      `json:"view_count"`
	Comment         *string   `json:"comment"`
	RefreshedAt     time.Time `json:"refreshed_at"`
	PostedAt        time.Time `json:"posted_at"`
	URL             string    `json:"url"`
}

// FetchPlateByID возвращает деталку номера
func (c *Client) FetchPlateByID(ctx context.Context, id uuid.UUID) (*FetchPlateByIDResponse, error) {
	requestURL := c.baseURL + fetchPlates + "/" + id.String()

	response, err := c.request(ctx, http.MethodGet, requestURL)
	if err != nil {
		return nil, err
	}

	var jsonResponse FetchPlateByIDResponse
	if err := json.Unmarshal(response, &jsonResponse); err != nil {
		return nil, fmt.Errorf("parse json from %s: %w", requestURL, err)
	}

	return &jsonResponse, nil
}

func (c *Client) buildFetchPlatesURL(params FetchPlatesParams) string {
	query := url.Values{}

	if params.Query != nil {
		query.Set("query", *params.Query)
	}
	if params.RegionID != nil {
		query.Set("region_id", strconv.Itoa(*params.RegionID))
	}
	if params.PriceFrom != nil {
		query.Set("price_from", strconv.FormatFloat(*params.PriceFrom, 'f', -1, 64))
	}
	if params.PriceTo != nil {
		query.Set("price_to", strconv.FormatFloat(*params.PriceTo, 'f', -1, 64))
	}
	if params.Limit > 0 {
		query.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Cursor != "" {
		query.Set("cursor", params.Cursor)
	}
	if params.ReissueIncluded != nil {
		query.Set("reissue_included", strconv.FormatBool(*params.ReissueIncluded))
	}
	for _, id := range params.CategoryIDs {
		query.Add("category_ids", strconv.Itoa(id))
	}

	return c.baseURL + fetchPlates + "?" + query.Encode()
}

// processBadStatus - Процессинг ошибки по статус коду
func processBadStatus(response *http.Response, requestURL string) error {
	var sentinel error

	switch {
	case response.StatusCode == http.StatusNotFound:
		sentinel = service.ErrNotFound
	case response.StatusCode >= http.StatusBadRequest && response.StatusCode < http.StatusInternalServerError:
		sentinel = service.ErrBadRequest
	case response.StatusCode >= http.StatusInternalServerError:
		sentinel = service.ErrInternalServiceError
	default:
		sentinel = service.ErrServiceUnavailable
	}

	var responsePart string
	preview, err := io.ReadAll(io.LimitReader(response.Body, 512))
	if err != nil || len(preview) == 0 {
		responsePart = "Body is empty"
	} else {
		responsePart = strings.Join(strings.Fields(string(preview)), " ")
	}

	return fmt.Errorf("%w: got %d from %s: %s", sentinel, response.StatusCode, requestURL, responsePart)
}
