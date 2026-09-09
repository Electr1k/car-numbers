package http

import (
	"data-service/internal/http/handler/feed"
	"data-service/internal/http/handler/plate"
	"data-service/internal/http/handler/region"
	"data-service/internal/http/response"
	"log/slog"

	"github.com/go-chi/chi/v5"
)

func NewRouter(
	logger *slog.Logger,
	regionHandler *region.Handler,
	feedHandler *feed.Handler,
	plateHandler *plate.Handler,
) chi.Router {
	r := chi.NewRouter()

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/regions", response.Wrap(logger, regionHandler.Handle))
		r.Get("/feed", response.Wrap(logger, feedHandler.Handle))
		r.Get("/plate/{id}", response.Wrap(logger, plateHandler.Handle))
	})

	return r
}
