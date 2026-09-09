package http

import (
	"context"
	"data-service/internal/http/response"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// accessLog - логирует результат запроса
func accessLog(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()

			defer func() {
				logger.InfoContext(r.Context(), "request completed",
					"method", r.Method,
					"path", r.URL.Path,
					"query", r.URL.RawQuery,
					"status", ww.Status(),
					"bytes", ww.BytesWritten(),
					"duration_ms", time.Since(start).Milliseconds(),
					"remote_addr", r.RemoteAddr,
					"request_id", middleware.GetReqID(r.Context()),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

// recoverer - ловит панику и отдает 500
func recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}
				if rec == http.ErrAbortHandler {
					panic(rec)
				}

				logger.ErrorContext(r.Context(), "panic recovered",
					"panic", fmt.Sprint(rec),
					"stack", string(debug.Stack()),
					"method", r.Method,
					"path", r.URL.Path,
					"request_id", middleware.GetReqID(r.Context()),
				)

				if err := response.WriteError(w, r, http.StatusInternalServerError, "internal_error", "Произошла ошибка"); err != nil {
					logger.ErrorContext(r.Context(), "write panic response failed", "error", err)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// timeout - ограничивает время обработки запроса через контекст
func timeout(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if d <= 0 {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
