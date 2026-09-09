package http

import (
	"data-service/config"
	"data-service/internal/http/handler/feed"
	"data-service/internal/http/handler/plate"
	"data-service/internal/http/handler/region"
	"data-service/internal/http/response"
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Handlers - хендлеры, обслуживающие маршруты api
type Handlers struct {
	Region *region.Handler
	Feed   *feed.Handler
	Plate  *plate.Handler
}

func NewRouter(cfg config.HTTPServer, logger *slog.Logger, handlers Handlers) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(accessLog(logger))
	r.Use(recoverer(logger))
	r.Use(timeout(cfg.RequestTimeout))

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/regions", response.Wrap(logger, handlers.Region.Handle))
		r.Get("/feed", response.Wrap(logger, handlers.Feed.Handle))
		r.Get("/plate/{id}", response.Wrap(logger, handlers.Plate.Handle))
	})

	return r
}
