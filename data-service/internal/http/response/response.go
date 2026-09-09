package response

import (
	"encoding/json"
	"net/http"
)

type errorWrapper struct {
	Error errorApi `json:"error"`
}

type errorApi struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteJSON - пишет тело ответа в JSON
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// WriteError - пишет ошибку в едином формате {"error": msg}
func WriteError(w http.ResponseWriter, status int, code string, message string) {
	WriteJSON(w, status, errorWrapper{Error: errorApi{Code: code, Message: message}})
}
