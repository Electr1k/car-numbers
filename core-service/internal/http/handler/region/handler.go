package region

import (
	"context"
	"core-service/internal/http/response"
	"core-service/internal/usecase/fetchregions"
	"fmt"
	"net/http"
)

type regionsFetcher interface {
	Handle(ctx context.Context) (*fetchregions.Result, error)
}

type Handler struct {
	fetchRegionsUC regionsFetcher
}

func New(fetchRegionsUC regionsFetcher) *Handler {
	return &Handler{fetchRegionsUC: fetchRegionsUC}
}

func (h *Handler) FetchRegions(w http.ResponseWriter, r *http.Request) error {
	res, err := h.fetchRegionsUC.Handle(r.Context())
	if err != nil {
		return fmt.Errorf("fetch regions: %w", err)
	}

	return response.WriteJSON(w, http.StatusOK, mapFetchRegionsResponse(*res))
}
