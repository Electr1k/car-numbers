package autonomera

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"plate-service/internal/provider"
	"strconv"
	"strings"
	"time"
)

const (
	// getNumbersPath - постраничная выдача номеров
	getNumbersPath = "/ajax/get_numbers.php"
	getUserPath    = "/user?user_id="

	// Порядок выдачи: свежие сверху
	orderColumn    = "a.`created`"
	orderDirection = "DESC"

	// defaultTimeout - таймаут по умолчанию
	defaultTimeout = 60 * time.Second

	maxResponseBytes = 8 << 20

	maxErrorPreviewBytes = 512

	userAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36"
)

// Client - HTTP-клиент к autonomera777
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

	request, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	request.Header.Set("User-Agent", userAgent)

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
		"body", string(responseByte),
	)

	return responseByte, nil
}

// FetchOffersHTML забирает одну страницу предложений и отдаёт сырой HTML
func (c *Client) FetchOffersHTML(ctx context.Context, section Section, start int) ([]byte, error) {
	requestURL := c.buildURL(section, start)

	return c.request(ctx, http.MethodGet, requestURL)
}

// FetchOfferDetailHTML забирает одну страницу предложений и отдаёт сырой HTML
func (c *Client) FetchOfferDetailHTML(ctx context.Context, url string) ([]byte, error) {
	return c.request(ctx, http.MethodGet, url)
}

// FetchUserHTML забирает страницу пользователя и отдаёт сырой HTML
func (c *Client) FetchUserHTML(ctx context.Context, id string) ([]byte, error) {
	return c.request(ctx, http.MethodGet, c.baseURL+getUserPath+url.QueryEscape(id))
}

func (c *Client) buildURL(section Section, start int) string {
	query := url.Values{}
	query.Set("order", orderColumn)
	query.Set("dir", orderDirection)
	query.Set("start", strconv.Itoa(start))

	if value, ok := section.queryValue(); ok {
		query.Set("blog", value)
	}

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
		responsePart = "<empty body>"
	} else {
		responsePart = strings.Join(strings.Fields(string(preview)), " ")
	}

	return fmt.Errorf("%w: got %d from %s: %s", statesError, response.StatusCode, requestURL, responsePart)
}
