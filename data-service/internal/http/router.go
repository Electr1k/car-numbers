package http

import (
	"data-service/config"
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
		r.Get("/plates", response.Wrap(logger, handlers.Plate.GetPlates))
		r.Get("/plates/{id}", response.Wrap(logger, handlers.Plate.GetPlateByID))
	})

	return r
}
