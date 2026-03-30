package autonomera

import (
	"core/pkg/logger"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://autonomera777.net/ajax/"
	defaultTimeout = 60 * time.Second
)

// Config содержит конфигурацию для HTTP клиента
type Config struct {
	BaseURL string
	Timeout time.Duration
}

// Client - HTTP клиент для работы с API autonomera777.net
type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *slog.Logger
}

// NewClient создает новый HTTP клиент для autonomera777.net
// Если BaseURL или Timeout не указаны, используются дефолтные значения
func NewClient(cfg Config) *Client {
	// Устанавливаем дефолтные значения
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}
	
	// Создаем логгер
	log := logger.New(logger.Config{
		Level:  "info",
		Format: "text",
	})
	
	return &Client{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger: log,
	}
}

func (c *Client) FetchNumbersHTML(start int) ([]byte, error) {
	fullURL := c.buildURL(start)
	
	c.logger.Info("Выполнение HTTP запроса",
		"url", fullURL,
		"start", start)
	
	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		c.logger.Error("Ошибка создания HTTP запроса", "error", err)
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Ошибка выполнения HTTP запроса", "error", err)
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Error("Получен не-2xx статус код",
			"status_code", resp.StatusCode,
			"body", string(body))
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Ошибка чтения тела ответа", "error", err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	c.logger.Info("HTTP запрос успешно выполнен",
		"status_code", resp.StatusCode,
		"body_size", len(body))
	
	return body, nil
}

// buildURL формирует полный URL с query параметрами
func (c *Client) buildURL(start int) string {
	queryParams := map[string]string{
		"order": "a.`created`",
		"dir":   "DESC",
		"start": strconv.Itoa(start),
	}
	
	var queryParts []string
	for key, value := range queryParams {
		queryParts = append(queryParts, fmt.Sprintf("%s=%s", key, value))
	}
	
	return c.baseURL + "get_numbers.php?" + strings.Join(queryParts, "&")
}
