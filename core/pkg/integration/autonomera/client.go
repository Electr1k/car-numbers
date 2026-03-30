package autonomera

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type Client struct {
	client  *http.Client
	logger  *slog.Logger
	baseURL string
	timeout time.Duration
}

type Config struct {
	BaseURL string
	Timeout time.Duration
}

func NewClient(cfg Config, logger *slog.Logger) *Client {
	return &Client{
		logger:  logger,
		client:  &http.Client{Timeout: cfg.Timeout},
		baseURL: cfg.BaseURL,
		timeout: cfg.Timeout,
	}
}

func (c *Client) FetchNumbersHTML(ctx context.Context, start int) ([]byte, error) {
	url := c.baseURL + "get_numbers.php"
	queryParams := map[string]string{
		"order": "a.`created`",
		"dir":   "DESC",
		"start": strconv.Itoa(start),
	}

	fullURL := url + "?" + c.buildQueryString(queryParams)

	c.logger.Debug("starting HTTP request",
		"method", http.MethodGet,
		"url", url,
		"start", start)

	startTime := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		duration := time.Since(startTime)
		c.logger.Error("http request failed",
			"url", fullURL,
			"error", err,
			"duration_ms", duration.Milliseconds())
		return nil, fmt.Errorf("execute request to %s: %w", fullURL, err)
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Error("unexpected status code",
			"url", fullURL,
			"status_code", resp.StatusCode,
			"duration_ms", duration.Milliseconds(),
			"response_body_preview", string(body[:min(len(body), 200)]))
		return nil, fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, fullURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("failed to read response body",
			"url", fullURL,
			"error", err,
			"duration_ms", duration.Milliseconds())
		return nil, fmt.Errorf("read response body: %w", err)
	}

	c.logger.Debug("http request completed",
		"url", fullURL,
		"status_code", resp.StatusCode,
		"duration_ms", duration.Milliseconds(),
		"response_size_bytes", len(body))

	return body, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (c *Client) buildQueryString(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}

	var queryParts []string
	for key, value := range params {
		queryParts = append(queryParts, fmt.Sprintf("%s=%s", key, value))
	}

	result := ""
	for i, part := range queryParts {
		if i > 0 {
			result += "&"
		}
		result += part
	}
	return result
}
