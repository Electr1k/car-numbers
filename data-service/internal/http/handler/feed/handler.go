package feed

import (
	"context"
	"data-service/internal/domain"
	"data-service/internal/http/request"
	"data-service/internal/http/response"
	"data-service/internal/usecase/fetchfeednumbers"
	"fmt"
	"net/http"
)

type feedFetcher interface {
	Handle(ctx context.Context, params fetchfeednumbers.Params) (fetchfeednumbers.Result, error)
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

	params := fetchfeednumbers.Params{Limit: req.Limit}
	if req.Cursor != "" {
		cursor, err := decodeCursor(req.Cursor)
		if err != nil {
			return fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err)
		}
		params.Cursor = &cursor
	}

	res, err := h.uc.Handle(r.Context(), params)
	if err != nil {
		return fmt.Errorf("fetch feed numbers: %w", err)
	}

	var nextCursor *string
	if res.NextCursor != nil {
		encoded, err := encodeCursor(*res.NextCursor)
		if err != nil {
			return err
		}
		nextCursor = &encoded
	}

	return response.WriteJSON(w, http.StatusOK, mapFeedNumbers(res.Numbers, nextCursor))
}
