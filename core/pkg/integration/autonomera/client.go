package autonomera

import (
	"context"
	"core/config"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

const (
	ProviderName = "autonomera777"

	GetNumbers = "/ajax/get_numbers.php"
)

type Client struct {
	client *http.Client
	logger *slog.Logger
	config config.AutoNomeraConfig
}

func NewClient(cfg config.AutoNomeraConfig, logger *slog.Logger) *Client {
	return &Client{
		client: &http.Client{Timeout: cfg.Timeout},
		logger: logger.With("provider", ProviderName),
		config: cfg,
	}
}

func (c *Client) FetchNumbersHTML(ctx context.Context, start int) ([]byte, error) {
	url := c.config.BaseURL + GetNumbers
	queryParams := map[string]string{
		"order": "a.`created`",
		"dir":   "DESC",
		"start": strconv.Itoa(start),
	}

	fullURL := url + "?" + c.buildQueryString(queryParams)

	c.logger.Debug("starting HTTP request", "method", http.MethodGet, "url", fullURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	response, err := c.client.Do(req)
	if err != nil {
		c.logger.Error("http request failed", "url", fullURL, "error", err)
		return nil, fmt.Errorf("execute request to %s: %w", fullURL, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		c.logger.Error("unexpected status code",
			"url", fullURL,
			"status_code", response.StatusCode,
			"response_body_preview", string(body[:min(len(body), 200)]))
		return nil, fmt.Errorf("unexpected status code %d from %s", response.StatusCode, fullURL)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		c.logger.Error("failed to read response body", "url", fullURL, "error", err)
		return nil, fmt.Errorf("read response body: %w", err)
	}

	c.logger.Debug("http request completed", "url", fullURL, "response", strings.NewReader(string(body)), "status_code", response.StatusCode)

	return body, nil
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
