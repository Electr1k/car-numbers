package profile

import (
	"plate-service/internal/domain/data"
	"time"

	"github.com/google/uuid"
)

type profileResponse struct {
	ID       uuid.UUID        `json:"id"`
	Provider string           `json:"provider"`
	URL      string           `json:"url"`
	IconURL  *string          `json:"icon_url"`
	Login    string           `json:"login"`
	Name     *string          `json:"name"`
	Phone    *string          `json:"phone"`
	Email    *string          `json:"email"`
	Offers   profileOffers    `json:"offers"`
	Related  []relatedProfile `json:"related"`
}

type profileOffers struct {
	Active int `json:"active"`
	Total  int `json:"total"`
}

type relatedProfile struct {
	ID       uuid.UUID     `json:"id"`
	Provider string        `json:"provider"`
	URL      string        `json:"url"`
	IconURL  *string       `json:"icon_url"`
	Login    string        `json:"login"`
	Name     *string       `json:"name"`
	Offers   profileOffers `json:"offers"`
	Matches  []string      `json:"matches"`
}

// mapProfile - маппинг профиля продавца в ответ API
func mapProfile(profile data.ProfileDetail) profileResponse {
	related := make([]relatedProfile, 0, len(profile.Related))
	for _, r := range profile.Related {
		matches := make([]string, 0, len(r.Matches))
		for _, m := range r.Matches {
			matches = append(matches, string(m))
		}

		related = append(related, relatedProfile{
			ID:       r.ID,
			Provider: string(r.Provider),
			URL:      r.URL,
			IconURL:  r.IconURL,
			Login:    r.Login,
			Name:     r.Name,
			Offers:   profileOffers(r.Offers),
			Matches:  matches,
		})
	}

	return profileResponse{
		ID:       profile.ID,
		Provider: string(profile.Provider),
		URL:      profile.URL,
		IconURL:  profile.IconURL,
		Login:    profile.Login,
		Name:     profile.Name,
		Phone:    profile.Phone,
		Email:    profile.Email,
		Offers:   profileOffers(profile.Offers),
		Related:  related,
	}
}

type profileOffersResponse struct {
	Items      []profileOffer `json:"items"`
	NextCursor *string        `json:"next_cursor"`
}

type profileOffer struct {
	ID          uuid.UUID `json:"id"`
	PlateID     uuid.UUID `json:"plate_id"`
	Number      string    `json:"number"`
	Price       *float64  `json:"price"`
	Status      string    `json:"status"`
	PostedAt    time.Time `json:"posted_at"`
	RefreshedAt time.Time `json:"refreshed_at"`
	URL         string    `json:"url"`
}

// mapProfileOffers - маппинг страницы предложений продавца в ответ API
func mapProfileOffers(offers []data.ProfileOffer, nextCursor *string) profileOffersResponse {
	items := make([]profileOffer, 0, len(offers))
	for _, o := range offers {
		items = append(items, profileOffer{
			ID:          o.ID,
			PlateID:     o.PlateID,
			Number:      o.Number,
			Price:       o.Price,
			Status:      string(o.Status),
			PostedAt:    o.PostedAt,
			RefreshedAt: o.RefreshedAt,
			URL:         o.URL,
		})
	}

	return profileOffersResponse{Items: items, NextCursor: nextCursor}
}
