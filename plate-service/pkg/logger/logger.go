package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Level  string // "debug", "info", "warn", "error"
	Format string // "json", "text"
}

func New(cfg Config) *slog.Logger {
	level := parseLevel(cfg.Level)

	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: level,
	}

	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(contextHandler{handler})
}

type attrsKey struct{}

// WithAttrs - кладёт атрибуты в контекст, они попадают во все логи с этим контекстом
func WithAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	stored, _ := ctx.Value(attrsKey{}).([]slog.Attr)

	return context.WithValue(ctx, attrsKey{}, append(stored[:len(stored):len(stored)], attrs...))
}

// contextHandler - дописывает в запись атрибуты из контекста
type contextHandler struct {
	slog.Handler
}

func (h contextHandler) Handle(ctx context.Context, record slog.Record) error {
	if attrs, ok := ctx.Value(attrsKey{}).([]slog.Attr); ok {
		record.AddAttrs(attrs...)
	}

	return h.Handler.Handle(ctx, record)
}

func (h contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return contextHandler{h.Handler.WithAttrs(attrs)}
}

func (h contextHandler) WithGroup(name string) slog.Handler {
	return contextHandler{h.Handler.WithGroup(name)}
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
