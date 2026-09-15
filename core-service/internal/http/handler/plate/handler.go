package plate

import (
	"context"
	"core-service/internal/http/request"
	"core-service/internal/http/response"
	"core-service/internal/service"
	"core-service/internal/usecase/fetchplate"
	"core-service/internal/usecase/fetchplates"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type platesFetcher interface {
	Handle(ctx context.Context, params fetchplates.Params) (*fetchplates.Result, error)
}

type plateByIDFetcher interface {
	Handle(ctx context.Context, id uuid.UUID) (*fetchplate.Result, error)
}

type Handler struct {
	fetchPlatesUC    platesFetcher
	fetchPlateByIDUC plateByIDFetcher
}

func New(
	fetchPlatesUC platesFetcher,
	fetchPlateByIDUC plateByIDFetcher,
) *Handler {
	return &Handler{
		fetchPlatesUC:    fetchPlatesUC,
		fetchPlateByIDUC: fetchPlateByIDUC,
	}
}

func (h *Handler) FetchPlates(w http.ResponseWriter, r *http.Request) error {
	req := newFetchPlatesRequest()
	if err := request.DecodeQuery(r, &req); err != nil {
		return fmt.Errorf("%w: query decode: %v", service.ErrBadRequest, err)
	}

	params := fetchplates.Params{
		Query:           req.Query,
		RegionID:        req.RegionID,
		PriceFrom:       req.PriceFrom,
		PriceTo:         req.PriceTo,
		ReissueIncluded: req.ReissueIncluded,
		CategoryIDs:     req.CategoryIDs,
		Limit:           req.Limit,
		Cursor:          req.Cursor,
	}

	res, err := h.fetchPlatesUC.Handle(r.Context(), params)
	if err != nil {
		return fmt.Errorf("fetch plates: %w", err)
	}

	return response.WriteJSON(w, http.StatusOK, mapFetchPlatesResponse(*res))
}

func (h *Handler) FetchPlateByID(w http.ResponseWriter, r *http.Request) error {
	idFromPath := chi.URLParam(r, "id")
	id, err := uuid.Parse(idFromPath)
	if err != nil {
		return fmt.Errorf("%w: parse uuid %s: %v", service.ErrBadRequest, idFromPath, err)
	}

	res, err := h.fetchPlateByIDUC.Handle(r.Context(), id)
	if err != nil {
		return fmt.Errorf("fetch plate by id: %w", err)
	}

	return response.WriteJSON(w, http.StatusOK, mapFetchPlateByIDResponse(*res))
}
