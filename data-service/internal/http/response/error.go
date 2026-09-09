package response

import (
	"data-service/internal/domain"
	"errors"
	"log/slog"
	"net/http"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// Wrap - оборачивает хендлер для единного вывода ошибок
func Wrap(logger *slog.Logger, handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			writeAPIError(w, r, logger, err)
		}
	}
}

func writeAPIError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	switch {
	case errors.Is(err, domain.ErrNumberNotFound):
		WriteError(w, http.StatusNotFound, "not_found", "Номер не найден")
	case errors.Is(err, domain.ErrInvalidArgument):
		WriteError(w, http.StatusBadRequest, "validation_error", "Некорректные параметры запроса")
	default:
		logger.ErrorContext(r.Context(), "request failed",
			"error", err,
			"method", r.Method,
			"path", r.URL.Path,
		)
		WriteError(w, http.StatusInternalServerError, "internal_error", "Произошла ошибка")
	}
}
