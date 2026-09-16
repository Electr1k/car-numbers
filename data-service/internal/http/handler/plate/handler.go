package plate

import (
	"context"
	"data-service/internal/domain"
	"data-service/internal/domain/data"
	"data-service/internal/http/request"
	"data-service/internal/http/response"
	"data-service/internal/usecase/fetchplates"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type platesFetcher interface {
	Handle(ctx context.Context, params fetchplates.Params) (fetchplates.Result, error)
}

type plateByIDFetcher interface {
	Handle(ctx context.Context, id uuid.UUID) (*data.Plate, error)
}

type Handler struct {
	fetchPlateByIDUC plateByIDFetcher
	fetchPlatesUC    platesFetcher
}

func New(
	fetchPlateByIDUC plateByIDFetcher,
	fetchPlatesUC platesFetcher,
) *Handler {
	return &Handler{fetchPlateByIDUC: fetchPlateByIDUC, fetchPlatesUC: fetchPlatesUC}
}

func (h *Handler) GetPlates(w http.ResponseWriter, r *http.Request) error {
	req := newPlatesRequest()
	if err := request.DecodeQuery(r, &req); err != nil {
		return fmt.Errorf("%w: query: %v", domain.ErrInvalidArgument, err)
	}

	params := fetchplates.Params{
		Query:           req.Query,
		RegionId:        req.RegionId,
		PriceFrom:       req.PriceFrom,
		PriceTo:         req.PriceTo,
		ReissueIncluded: req.ReissueIncluded,
		CategoryIds:     req.CategoryIds,
		Limit:           req.Limit,
	}
	if req.Cursor != "" {
		cursor, err := decodeCursor(req.Cursor)
		if err != nil {
			return fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err)
		}
		params.Cursor = &cursor
	}

	res, err := h.fetchPlatesUC.Handle(r.Context(), params)
	if err != nil {
		return fmt.Errorf("fetch plates: %w", err)
	}

	var nextCursor *string
	if res.NextCursor != nil {
		encoded, err := encodeCursor(*res.NextCursor)
		if err != nil {
			return err
		}
		nextCursor = &encoded
	}

	return response.WriteJSON(w, http.StatusOK, mapPlates(res.Plates, nextCursor))
}

func (h *Handler) GetPlateByID(w http.ResponseWriter, r *http.Request) error {
	plateID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return fmt.Errorf("%w: id: %v", domain.ErrInvalidArgument, err)
	}

	res, err := h.fetchPlateByIDUC.Handle(r.Context(), plateID)
	if err != nil {
		return fmt.Errorf("fetch plate: %w", err)
	}

	return response.WriteJSON(w, http.StatusOK, mapPlate(*res))
}
