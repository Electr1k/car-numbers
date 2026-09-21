package http

import (
	"context"
	"core-service/internal/http/response"
	"core-service/internal/service"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/time/rate"
)

func requestIDHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(middleware.RequestIDHeader, middleware.GetReqID(r.Context()))
		next.ServeHTTP(w, r)
	})
}

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

// visitorTTL
const visitorTTL = 10 * time.Minute

// visitorsCleanupInterval
const visitorsCleanupInterval = 5 * time.Minute

// realIPHeader - заголовок с IP клиента
const realIPHeader = "X-Real-IP"

// visitor - лимитер клиента и время его последнего запроса
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// ipLimiter - рейт лимит запросов по IP клиента
type ipLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    rate.Limit
	burst    int
	logger   *slog.Logger
}

func newIPLimiter(rpm int, logger *slog.Logger) *ipLimiter {
	return &ipLimiter{
		visitors: make(map[string]*visitor),
		limit:    rate.Every(time.Minute / time.Duration(rpm)),
		burst:    rpm,
		logger:   logger,
	}
}

// middleware - рейт лимит
func (l *ipLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limiter := l.visitor(clientIP(r))
		if !limiter.Allow() {
			response.WriteAPIError(w, r, l.logger, service.ErrTooManyRequests)
			return
		}

		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(int(limiter.Tokens())))

		next.ServeHTTP(w, r)
	})
}

// visitor - возвращает лимитер клиента
func (l *ipLimiter) visitor(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	v, ok := l.visitors[ip]
	if !ok {
		v = &visitor{limiter: rate.NewLimiter(l.limit, l.burst)}
		l.visitors[ip] = v
		l.logger.Info("New user call endpoint", "ip", ip)
	}
	v.lastSeen = time.Now()

	return v.limiter
}

// cleanup - удаляет клиентов без запросов дольше visitorTTL
func (l *ipLimiter) cleanup() {
	for {
		time.Sleep(visitorsCleanupInterval)
		l.mu.Lock()
		for ip, v := range l.visitors {
			if time.Now().Sub(v.lastSeen) > visitorTTL {
				delete(l.visitors, ip)
			}
		}
		l.mu.Unlock()
	}
}

// clientIP - достает IP из запроса
func clientIP(r *http.Request) string {
	if ip := r.Header.Get(realIPHeader); ip != "" {
		return ip
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}
