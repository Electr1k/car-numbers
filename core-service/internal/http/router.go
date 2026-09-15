package http

import (
	"core-service/config"
	"core-service/internal/http/handler/plate"
	"core-service/internal/http/handler/region"
	"core-service/internal/http/response"
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Handlers - хендлеры, обслуживающие маршруты api
type Handlers struct {
	Plate   *plate.Handler
	Regions *region.Handler
}

func NewRouter(cfg config.HTTPServer, logger *slog.Logger, handlers Handlers) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(requestIDHeader)
	r.Use(accessLog(logger))
	r.Use(recoverer(logger))
	r.Use(timeout(cfg.RequestTimeout))

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/plates", response.Wrap(logger, handlers.Plate.FetchPlates))
		r.Get("/regions", response.Wrap(logger, handlers.Regions.FetchRegions))
	})

	return r
}
