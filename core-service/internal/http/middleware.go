package http

import (
	"context"
	"core-service/internal/http/response"
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

const (
	visitorTTL              = 10 * time.Minute
	visitorsCleanupInterval = 5 * time.Minute
	realIPHeader            = "X-Real-IP"
)

// visitor - лимитер клиента и время его последнего запроса
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// ipLimiter - рейт лимит запросов по IP клиента
type ipLimiter struct {
	mu          sync.Mutex
	visitors    map[string]*visitor
	lastCleanup time.Time
	limit       rate.Limit
	burst       int
	endpoint    string
	logger      *slog.Logger
}

// rateLimit - лимит по IP для эндпоинта
func rateLimit(endpoint string, rpm int, logger *slog.Logger) func(http.Handler) http.Handler {
	if rpm <= 0 {
		return func(next http.Handler) http.Handler { return next }
	}

	return newIPLimiter(endpoint, rpm, logger).handler
}

// newIPLimiter - лимитер на rpm запросов в минуту с одного IP
func newIPLimiter(endpoint string, rpm int, logger *slog.Logger) *ipLimiter {
	return &ipLimiter{
		visitors:    make(map[string]*visitor),
		lastCleanup: time.Now(),
		limit:       rate.Every(time.Minute / time.Duration(rpm)),
		burst:       rpm,
		endpoint:    endpoint,
		logger:      logger,
	}
}

// handler - отдает 429, если клиент превысил лимит
func (l *ipLimiter) handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		ip := clientIP(r)

		limiter := l.getLimiter(ip, now)
		if !limiter.AllowN(now, 1) {
			l.logger.WarnContext(r.Context(), "too many requests",
				"endpoint", l.endpoint,
				"ip", ip,
				"request_id", middleware.GetReqID(r.Context()),
			)
			if err := response.WriteError(w, r, http.StatusTooManyRequests, "too_many_requests", "Превышен лимит запросов"); err != nil {
				l.logger.ErrorContext(r.Context(), "write rate limit response failed", "error", err)
			}
			return
		}

		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(int(limiter.TokensAt(now))))

		next.ServeHTTP(w, r)
	})
}

// visitor - возвращает лимитер клиента
func (l *ipLimiter) getLimiter(ip string, now time.Time) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	if now.Sub(l.lastCleanup) > visitorsCleanupInterval {
		for k, v := range l.visitors {
			if now.Sub(v.lastSeen) > visitorTTL {
				delete(l.visitors, k)
			}
		}
		l.lastCleanup = now
	}

	v, ok := l.visitors[ip]
	if !ok {
		v = &visitor{limiter: rate.NewLimiter(l.limit, l.burst)}
		l.visitors[ip] = v
		l.logger.Info("New user call endpoint", "endpoint", l.endpoint, "ip", ip)
	}
	v.lastSeen = now

	return v.limiter
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
