package profile

import (
	"context"
	"fmt"
	"net/http"
	"plate-service/internal/domain"
	"plate-service/internal/domain/data"
	"plate-service/internal/http/request"
	"plate-service/internal/http/response"
	"plate-service/internal/usecase/fetchprofileoffers"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type profileFetcher interface {
	Handle(ctx context.Context, id uuid.UUID) (*data.ProfileDetail, error)
}

type profileOffersFetcher interface {
	Handle(ctx context.Context, params fetchprofileoffers.Params) (fetchprofileoffers.Result, error)
}

type Handler struct {
	fetchProfileUC       profileFetcher
	fetchProfileOffersUC profileOffersFetcher
}

func New(fetchProfileUC profileFetcher, fetchProfileOffersUC profileOffersFetcher) *Handler {
	return &Handler{fetchProfileUC: fetchProfileUC, fetchProfileOffersUC: fetchProfileOffersUC}
}

func (h *Handler) GetProfileByID(w http.ResponseWriter, r *http.Request) error {
	profileID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return fmt.Errorf("%w: id: %v", domain.ErrInvalidArgument, err)
	}

	res, err := h.fetchProfileUC.Handle(r.Context(), profileID)
	if err != nil {
		return fmt.Errorf("fetch profile: %w", err)
	}

	return response.WriteJSON(w, http.StatusOK, mapProfile(*res))
}

func (h *Handler) GetProfileOffers(w http.ResponseWriter, r *http.Request) error {
	profileID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return fmt.Errorf("%w: id: %v", domain.ErrInvalidArgument, err)
	}

	req := newProfileOffersRequest()
	if err := request.DecodeQuery(r, &req); err != nil {
		return fmt.Errorf("%w: query: %v", domain.ErrInvalidArgument, err)
	}

	statuses := make([]domain.OfferStatus, 0, len(req.Statuses))
	for _, status := range req.Statuses {
		statuses = append(statuses, domain.OfferStatus(status))
	}

	params := fetchprofileoffers.Params{
		ProfileID: profileID,
		Statuses:  statuses,
		Limit:     req.Limit,
	}
	if req.Cursor != "" {
		cursor, err := decodeCursor(req.Cursor)
		if err != nil {
			return fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err)
		}
		params.Cursor = &cursor
	}

	res, err := h.fetchProfileOffersUC.Handle(r.Context(), params)
	if err != nil {
		return fmt.Errorf("fetch profile offers: %w", err)
	}

	var nextCursor *string
	if res.NextCursor != nil {
		encoded, err := encodeCursor(*res.NextCursor)
		if err != nil {
			return err
		}
		nextCursor = &encoded
	}

	return response.WriteJSON(w, http.StatusOK, mapProfileOffers(res.Offers, nextCursor))
}
