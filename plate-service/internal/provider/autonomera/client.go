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

	// Порядок выдачи: свежие сверху
	orderColumn    = "a.`created`"
	orderDirection = "DESC"

	// detailRateLimitTimeout - пауза перед запросом деталки
	detailRateLimitTimeout = 61 * time.Second

	// defaultTimeout - таймаут по умолчанию
	defaultTimeout = 60 * time.Second

	maxResponseBytes = 8 << 20

	maxErrorPreviewBytes = 512

	acceptLanguage = "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7"

	acceptAjax = "text/plain, */*; q=0.01"

	acceptPage = "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8"

	userAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36"

	// emptyPageBody - тело ответа за концом списка
	emptyPageBody = "no"
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

func (c *Client) request(ctx context.Context, method string, url string, headers http.Header) ([]byte, error) {
	c.logger.DebugContext(ctx, "request", "url", url, "method", method)

	request, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	request.Header = headers

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
	c.logger.DebugContext(ctx, "response",
		"url", url,
		"method", method,
		"status", response.StatusCode,
		"bytes", len(responseByte),
	)

	return responseByte, nil
}

// FetchOffersHTML забирает одну страницу предложений и отдаёт сырой HTML
func (c *Client) FetchOffersHTML(ctx context.Context, section Section, start int) ([]byte, error) {
	body, err := c.request(ctx, http.MethodGet, c.buildURL(section, start), c.ajaxHeaders(section))
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(string(body)) == emptyPageBody {
		return nil, nil
	}

	return body, nil
}

// FetchOfferDetailHTML забирает страницу предложения и отдаёт сырой HTML
func (c *Client) FetchOfferDetailHTML(ctx context.Context, url string) ([]byte, error) {
	// Спим для рейтлимитов
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(detailRateLimitTimeout):
	}

	return c.request(ctx, http.MethodGet, url, c.pageHeaders())
}

// ajaxHeaders - заголовки, с которыми список запрашивает браузер
func (c *Client) ajaxHeaders(section Section) http.Header {
	headers := http.Header{}
	headers.Set("User-Agent", userAgent)
	headers.Set("Accept", acceptAjax)
	headers.Set("Accept-Language", acceptLanguage)
	headers.Set("Referer", c.baseURL+section.pagePath())
	headers.Set("X-Requested-With", "XMLHttpRequest")

	return headers
}

// pageHeaders - заголовки, с которыми браузер открывает страницу
func (c *Client) pageHeaders() http.Header {
	headers := http.Header{}
	headers.Set("User-Agent", userAgent)
	headers.Set("Accept", acceptPage)
	headers.Set("Accept-Language", acceptLanguage)
	headers.Set("Referer", c.baseURL+"/")

	return headers
}

func (c *Client) buildURL(section Section, start int) string {
	query := url.Values{}
	query.Set("order", orderColumn)
	query.Set("dir", orderDirection)
	query.Set("start", strconv.Itoa(start))
	query.Set("blog", section.blog())

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
