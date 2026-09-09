package http

import (
	"data-service/internal/http/handler/feed"
	"data-service/internal/http/handler/plate"
	"data-service/internal/http/handler/region"

	"github.com/go-chi/chi/v5"
)

func NewRouter(regionHandler *region.Handler, feedHandler *feed.Handler, plateHandler *plate.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/regions", regionHandler.Handle)
		r.Get("/feed", feedHandler.Handle)
		r.Get("/plate/{id}", plateHandler.Handle)
	})

	return r
}
