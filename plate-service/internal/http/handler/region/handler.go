package region

import (
	"context"
	"fmt"
	"net/http"
	"plate-service/internal/domain"
	"plate-service/internal/http/response"
)

type regionFetcher interface {
	Handle(ctx context.Context) ([]domain.RegionWithCodes, error)
}

type Handler struct {
	uc regionFetcher
}

func New(uc regionFetcher) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) error {
	res, err := h.uc.Handle(r.Context())
	if err != nil {
		return fmt.Errorf("fetch regions: %w", err)
	}

	return response.WriteJSON(w, http.StatusOK, mapRegions(res))
}
