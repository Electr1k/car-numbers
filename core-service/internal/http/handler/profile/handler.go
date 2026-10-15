package profile

import (
	"context"
	"core-service/internal/http/request"
	"core-service/internal/http/response"
	"core-service/internal/service"
	"core-service/internal/usecase/fetchprofile"
	"core-service/internal/usecase/fetchprofileoffers"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type profileFetcher interface {
	Handle(ctx context.Context, id uuid.UUID) (*fetchprofile.Result, error)
}

type profileOffersFetcher interface {
	Handle(ctx context.Context, params fetchprofileoffers.Params) (*fetchprofileoffers.Result, error)
}

type Handler struct {
	fetchProfileUC       profileFetcher
	fetchProfileOffersUC profileOffersFetcher
}

func New(fetchProfileUC profileFetcher, fetchProfileOffersUC profileOffersFetcher) *Handler {
	return &Handler{fetchProfileUC: fetchProfileUC, fetchProfileOffersUC: fetchProfileOffersUC}
}

func (h *Handler) FetchProfileByID(w http.ResponseWriter, r *http.Request) error {
	idFromPath := chi.URLParam(r, "id")
	id, err := uuid.Parse(idFromPath)
	if err != nil {
		return fmt.Errorf("%w: parse uuid %s: %v", service.ErrBadRequest, idFromPath, err)
	}

	res, err := h.fetchProfileUC.Handle(r.Context(), id)
	if err != nil {
		return fmt.Errorf("fetch profile by id: %w", err)
	}

	return response.WriteJSON(w, http.StatusOK, mapFetchProfileByIDResponse(*res))
}

func (h *Handler) FetchProfileOffers(w http.ResponseWriter, r *http.Request) error {
	idFromPath := chi.URLParam(r, "id")
	id, err := uuid.Parse(idFromPath)
	if err != nil {
		return fmt.Errorf("%w: parse uuid %s: %v", service.ErrBadRequest, idFromPath, err)
	}

	req := newFetchProfileOffersRequest()
	if err := request.DecodeQuery(r, &req); err != nil {
		return fmt.Errorf("%w: query decode: %v", service.ErrBadRequest, err)
	}

	res, err := h.fetchProfileOffersUC.Handle(r.Context(), fetchprofileoffers.Params{
		ProfileID: id,
		Statuses:  req.Statuses,
		Limit:     req.Limit,
		Cursor:    req.Cursor,
	})
	if err != nil {
		return fmt.Errorf("fetch profile offers: %w", err)
	}

	return response.WriteJSON(w, http.StatusOK, mapFetchProfileOffersResponse(*res))
}
