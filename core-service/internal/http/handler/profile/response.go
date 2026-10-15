package profile

import (
	"core-service/internal/usecase/fetchprofile"
	"core-service/internal/usecase/fetchprofileoffers"
	"time"

	"github.com/google/uuid"
)

type fetchProfileByIDResponse struct {
	ID       uuid.UUID             `json:"id"`
	Provider string                `json:"provider"`
	URL      string                `json:"url"`
	IconURL  *string               `json:"icon_url"`
	Login    string                `json:"login"`
	Name     *string               `json:"name"`
	Phone    *string               `json:"phone"`
	Email    *string               `json:"email"`
	Offers   fetchProfileOffers    `json:"offers"`
	Related  []fetchProfileRelated `json:"related"`
}

type fetchProfileOffers struct {
	Active int `json:"active"`
	Total  int `json:"total"`
}

type fetchProfileRelated struct {
	ID       uuid.UUID          `json:"id"`
	Provider string             `json:"provider"`
	URL      string             `json:"url"`
	IconURL  *string            `json:"icon_url"`
	Login    string             `json:"login"`
	Name     *string            `json:"name"`
	Offers   fetchProfileOffers `json:"offers"`
	Matches  []string           `json:"matches"`
}

func mapFetchProfileByIDResponse(result fetchprofile.Result) fetchProfileByIDResponse {
	related := make([]fetchProfileRelated, 0, len(result.Related))
	for _, r := range result.Related {
		matches := r.Matches
		if matches == nil {
			matches = []string{}
		}

		related = append(related, fetchProfileRelated{
			ID:       r.ID,
			Provider: r.Provider,
			URL:      r.URL,
			IconURL:  r.IconURL,
			Login:    r.Login,
			Name:     r.Name,
			Offers:   fetchProfileOffers(r.Offers),
			Matches:  matches,
		})
	}

	return fetchProfileByIDResponse{
		ID:       result.ID,
		Provider: result.Provider,
		URL:      result.URL,
		IconURL:  result.IconURL,
		Login:    result.Login,
		Name:     result.Name,
		Phone:    result.Phone,
		Email:    result.Email,
		Offers:   fetchProfileOffers(result.Offers),
		Related:  related,
	}
}

type fetchProfileOffersResponse struct {
	Items      []fetchProfileOffer `json:"items"`
	NextCursor *string             `json:"next_cursor"`
}

type fetchProfileOffer struct {
	ID          uuid.UUID `json:"id"`
	PlateID     uuid.UUID `json:"plate_id"`
	Number      string    `json:"number"`
	Price       *float64  `json:"price"`
	Status      string    `json:"status"`
	PostedAt    time.Time `json:"posted_at"`
	RefreshedAt time.Time `json:"refreshed_at"`
	URL         string    `json:"url"`
}

func mapFetchProfileOffersResponse(result fetchprofileoffers.Result) fetchProfileOffersResponse {
	items := make([]fetchProfileOffer, 0, len(result.Items))
	for _, o := range result.Items {
		items = append(items, fetchProfileOffer(o))
	}

	return fetchProfileOffersResponse{Items: items, NextCursor: result.NextCursor}
}
