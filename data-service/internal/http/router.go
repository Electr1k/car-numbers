package http

import (
	"data-service/internal/http/handler/region"

	"github.com/go-chi/chi/v5"
)

func NewRouter(regionHandler *region.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/regions", regionHandler.Handle)
	})

	return r
}
