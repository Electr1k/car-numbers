package response

import (
	"context"
	"data-service/internal/domain"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

const statusClientClosedRequest = 499

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// Wrap - оборачивает хендлер для единного вывода ошибок
func Wrap(logger *slog.Logger, handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		responseWriter := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		err := handler(responseWriter, r)
		if err == nil {
			return
		}

		if responseWriter.Status() != 0 {
			logger.ErrorContext(r.Context(), "request failed after response was written", logAttrs(r, err)...)
			return
		}

		writeAPIError(responseWriter, r, logger, err)
	}
}

func writeAPIError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	status, code, message := http.StatusInternalServerError, "internal_error", "Произошла ошибка"

	switch {
	case errors.Is(err, domain.ErrNumberNotFound):
		status, code, message = http.StatusNotFound, "not_found", "Номер не найден"
	case errors.Is(err, domain.ErrInvalidArgument):
		status, code, message = http.StatusBadRequest, "validation_error", "Некорректные параметры запроса"
	case errors.Is(err, context.DeadlineExceeded):
		status, code, message = http.StatusGatewayTimeout, "timeout", "Превышено время обработки запроса"
		logger.WarnContext(r.Context(), "request timed out", logAttrs(r, err)...)
	case errors.Is(err, context.Canceled):
		status, code, message = statusClientClosedRequest, "client_closed_request", "Запрос отменён клиентом"
		logger.WarnContext(r.Context(), "request canceled by client", logAttrs(r, err)...)
	default:
		logger.ErrorContext(r.Context(), "request failed", logAttrs(r, err)...)
	}

	if err := WriteError(w, r, status, code, message); err != nil {
		logger.ErrorContext(r.Context(), "write error response failed", logAttrs(r, err)...)
	}
}

func logAttrs(r *http.Request, err error) []any {
	return []any{
		"error", err,
		"method", r.Method,
		"path", r.URL.Path,
		"request_id", middleware.GetReqID(r.Context()),
	}
}
