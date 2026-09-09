package plate

import (
	"context"
	"data-service/internal/domain"
	"data-service/internal/domain/data"
	"data-service/internal/http/response"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type numberFetcher interface {
	Handle(ctx context.Context, id uuid.UUID) (*data.Number, error)
}

type Handler struct {
	uc numberFetcher
}

func New(uc numberFetcher) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) error {
	plateId, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return fmt.Errorf("%w: id: %v", domain.ErrInvalidArgument, err)
	}

	res, err := h.uc.Handle(r.Context(), plateId)
	if err != nil {
		return fmt.Errorf("fetch plate: %w", err)
	}

	return response.WriteJSON(w, http.StatusOK, mapNumber(*res))
}
