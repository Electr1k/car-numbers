package fetchprofile

import (
	"context"
	"core-service/internal/service/plate"

	"github.com/google/uuid"
)

type profileProvider interface {
	FetchProfileByID(ctx context.Context, id uuid.UUID) (*plate.FetchProfileByIDResponse, error)
}

// UseCase - возвращает профиль продавца
type UseCase struct {
	plateClient profileProvider
}

func New(plateClient profileProvider) *UseCase {
	return &UseCase{plateClient: plateClient}
}

// Handle - профиль продавца вместе со связанными аккаунтами
func (uc *UseCase) Handle(ctx context.Context, id uuid.UUID) (*Result, error) {
	profile, err := uc.plateClient.FetchProfileByID(ctx, id)
	if err != nil {
		return nil, err
	}

	related := make([]Related, 0, len(profile.Related))
	for _, r := range profile.Related {
		related = append(related, Related{
			ID:       r.ID,
			Provider: r.Provider,
			URL:      r.URL,
			IconURL:  r.IconURL,
			Login:    r.Login,
			Name:     r.Name,
			Offers:   Offers(r.Offers),
			Matches:  r.Matches,
		})
	}

	return &Result{
		ID:       profile.ID,
		Provider: profile.Provider,
		URL:      profile.URL,
		IconURL:  profile.IconURL,
		Login:    profile.Login,
		Name:     profile.Name,
		Phone:    profile.Phone,
		Email:    profile.Email,
		Offers:   Offers(profile.Offers),
		Related:  related,
	}, nil
}
