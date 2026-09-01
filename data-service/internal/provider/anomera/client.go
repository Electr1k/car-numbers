package anomera

import (
	"context"
	"data-service/internal/provider"
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
	// getAutoNumbers - постраничная выдача автономеров
	getAutoNumbers = "/gos-nomera-v-baze"

	// getMotoNumbers - постраничная выдача мотономеров
	getMotoNumbers = getAutoNumbers + "/nomera-na-moto"

	// getTrailerNumbers - постраничная выдача номеров на прицепы
	getTrailerNumbers = getAutoNumbers + "/nomera-na-pricep"

	orderColumn = "date"

	// defaultTimeout - таймаут по умолчанию
	defaultTimeout = 60 * time.Second

	maxResponseBytes = 8 << 20

	maxErrorPreviewBytes = 512
)

// Client - HTTP-клиент к anomera
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

// FetchOffers забирает одну страницу предложений и отдаёт сырой объект json
func (c *Client) FetchOffers(ctx context.Context, page int) ([]byte, error) {
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

	c.logger.Debug("page fetched",
		"page", page,
		"status_code", response.StatusCode,
		"bytes", len(body))

	return body, nil
}

func (c *Client) buildURL(page int) string {
	query := url.Values{}
	query.Set("sort", orderColumn)
	query.Set("page", strconv.Itoa(page))

	return c.baseURL + getAutoNumbers + "?" + query.Encode()
}

// FetchOfferDetail забирает деталку предложения
func (c *Client) FetchOfferDetail(ctx context.Context, url string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	c.logger.Debug("fetching offer detail page", "url", url)

	response, err := c.http.Do(request)
	if err != nil {
		return nil, fmt.Errorf("%w: get %s: %w", provider.ErrProviderUnavailable, url, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, processBadStatus(response, url)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("%w: read body from %s: %w", provider.ErrInvalidResponse, url, err)
	}

	c.logger.Debug("offer detail fetched",
		"url", url,
		"status_code", response.StatusCode,
		"bytes", len(body))

	return body, nil
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
		responsePart = "<empty body>"
	} else {
		responsePart = strings.Join(strings.Fields(string(preview)), " ")
	}

	return fmt.Errorf("%w: got %d from %s: %s", statesError, response.StatusCode, requestURL, responsePart)
}
