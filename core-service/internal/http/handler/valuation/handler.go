package valuation

import (
	"context"
	"core-service/internal/domain"
	"core-service/internal/http/request"
	"core-service/internal/http/response"
	"core-service/internal/service"
	"core-service/internal/usecase/valuation"
	"fmt"
	"net/http"
)

type valuationFetcher interface {
	Handle(ctx context.Context, number string) (*valuation.Result, error)
}

type Handler struct {
	fetchValuationUC valuationFetcher
}

func New(
	fetchValuationUC valuationFetcher,
) *Handler {
	return &Handler{
		fetchValuationUC: fetchValuationUC,
	}
}

func (h *Handler) FetchValuation(w http.ResponseWriter, r *http.Request) error {
	req := newValuationRequest()
	if err := request.DecodeQuery(r, &req); err != nil {
		return fmt.Errorf("%w: query decode: %v", service.ErrBadRequest, err)
	}

	number := domain.NormalizePlate(req.Number)
	if err := validateNumber(number); err != nil {
		return fmt.Errorf("validate number %q: %w", req.Number, err)
	}

	res, err := h.fetchValuationUC.Handle(r.Context(), number)
	if err != nil {
		return fmt.Errorf("valuation number: %w", err)
	}

	return response.WriteJSON(w, http.StatusOK, mapValuationResponse(*res))
}
