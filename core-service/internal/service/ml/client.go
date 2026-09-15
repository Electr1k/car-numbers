package ml

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
	"strings"
)

const (
	predict = "/api/v1/predict"

	maxResponseSize     = 8 << 20
	maxIdleConns        = 100
	maxIdleConnsPerHost = 100
)

// Client - HTTP-клиент к ml-service
type Client struct {
	baseURL string
	http    *http.Client
	logger  *slog.Logger
}

func NewClient(cfg config.MLConfig, logger *slog.Logger) *Client {
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

type PredictResponse struct {
	Number     string           `json:"number"`
	Price      PredictPrice     `json:"price"`
	Confidence string           `json:"confidence"`
	Breakdown  PredictBreakdown `json:"breakdown"`
}

type PredictPrice struct {
	P25 int `json:"p25"`
	P50 int `json:"p50"`
	P75 int `json:"p75"`
}

type PredictBreakdown struct {
	Base  int                    `json:"base"`
	Items []PredictBreakdownItem `json:"items"`
}

type PredictBreakdownItem struct {
	Code       string  `json:"code"`
	Value      string  `json:"value"`
	Title      string  `json:"title"`
	Multiplier float64 `json:"multiplier"`
	Exact      bool    `json:"exact"`
}

// Predict предсказывает цену на номер
func (c *Client) Predict(ctx context.Context, number string) (*PredictResponse, error) {
	query := url.Values{}
	query.Set("number", number)
	requestURL := c.baseURL + predict + "?" + query.Encode()

	response, err := c.request(ctx, http.MethodGet, requestURL)
	if err != nil {
		return nil, err
	}

	var jsonResponse PredictResponse
	if err := json.Unmarshal(response, &jsonResponse); err != nil {
		return nil, fmt.Errorf("parse json from %s: %w", requestURL, err)
	}

	return &jsonResponse, nil
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
