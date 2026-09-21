package http

import (
	"core-service/config"
	"core-service/internal/http/handler/plate"
	"core-service/internal/http/handler/region"
	"core-service/internal/http/handler/valuation"
	"core-service/internal/http/response"
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Handlers - хендлеры, обслуживающие маршруты api
type Handlers struct {
	Plate     *plate.Handler
	Regions   *region.Handler
	Valuation *valuation.Handler
}

func NewRouter(cfg config.HTTPServer, logger *slog.Logger, handlers Handlers) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(requestIDHeader)
	r.Use(accessLog(logger))
	r.Use(recoverer(logger))
	r.Use(timeout(cfg.RequestTimeout))

	r.Route("/api/v1", func(r chi.Router) {
		r.With(rateLimit("/plates", cfg.RateLimit.Plates, logger)).Get("/plates", response.Wrap(logger, handlers.Plate.FetchPlates))
		r.With(rateLimit("/plates/{id}", cfg.RateLimit.Plate, logger)).Get("/plates/{id}", response.Wrap(logger, handlers.Plate.FetchPlateByID))
		r.Get("/regions", response.Wrap(logger, handlers.Regions.FetchRegions))
		r.With(rateLimit("/valuation", cfg.RateLimit.Valuation, logger)).Get("/valuation", response.Wrap(logger, handlers.Valuation.FetchValuation))
	})

	return r
}
