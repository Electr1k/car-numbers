package response

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

type errorWrapper struct {
	Error errorApi `json:"error"`
}

type errorApi struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestId string `json:"request_id,omitempty"`
}

// WriteJSON - пишет тело ответа в JSON
func WriteJSON(w http.ResponseWriter, status int, v any) error {
	body, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal response: %w", err)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)

	if _, err := w.Write(body); err != nil {
		return fmt.Errorf("write response: %w", err)
	}

	return nil
}

// WriteError - пишет ошибку в едином формате {"error": {...}}
func WriteError(w http.ResponseWriter, r *http.Request, status int, code string, message string) error {
	return WriteJSON(w, status, errorWrapper{Error: errorApi{
		Code:      code,
		Message:   message,
		RequestId: middleware.GetReqID(r.Context()),
	}})
}
