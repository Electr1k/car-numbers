package feed

import (
	"context"
	"data-service/internal/domain"
	"data-service/internal/domain/data"
	"data-service/internal/http/request"
	"data-service/internal/http/response"
	"data-service/internal/usecase/fetchfeednumbers"
	"fmt"
	"net/http"
)

type feedFetcher interface {
	Handle(ctx context.Context, params fetchfeednumbers.Params) ([]data.FeedNumber, error)
}

type Handler struct {
	uc feedFetcher
}

func New(uc feedFetcher) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) error {
	req := newFeedRequest()
	if err := request.DecodeQuery(r, &req); err != nil {
		return fmt.Errorf("%w: query: %v", domain.ErrInvalidArgument, err)
	}

	res, err := h.uc.Handle(r.Context(), fetchfeednumbers.Params{Limit: req.Limit, Offset: req.Offset})
	if err != nil {
		return fmt.Errorf("fetch feed numbers: %w", err)
	}

	return response.WriteJSON(w, http.StatusOK, mapFeedNumbers(res))
}
