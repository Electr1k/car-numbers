package plate

import (
	"context"
	"core-service/internal/http/request"
	"core-service/internal/http/response"
	"core-service/internal/service"
	"core-service/internal/usecase/fetchplates"
	"fmt"
	"net/http"
)

type platesFetcher interface {
	Handle(ctx context.Context, params fetchplates.Params) (*fetchplates.Result, error)
}

type Handler struct {
	fetchPlatesUC platesFetcher
}

func New(fetchPlatesUC platesFetcher) *Handler {
	return &Handler{fetchPlatesUC: fetchPlatesUC}
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
