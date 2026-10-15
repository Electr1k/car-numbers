package fetchprofileoffers

import (
	"context"
	"plate-service/internal/domain/data"
	"plate-service/internal/repository"
)

type offerStore interface {
	GetProfileOffers(ctx context.Context, params repository.GetProfileOffersParams) ([]data.ProfileOffer, bool, error)
}

// UseCase - возвращает предложения продавца
type UseCase struct {
	repository offerStore
}

func New(
	repository offerStore,
) *UseCase {
	return &UseCase{
		repository: repository,
	}
}

// Handle - возвращает страницу предложений продавца
func (uc *UseCase) Handle(ctx context.Context, params Params) (Result, error) {
	err := params.validate()
	if err != nil {
		return Result{}, err
	}

	offers, hasNext, err := uc.repository.GetProfileOffers(ctx, repository.GetProfileOffersParams{
		ProfileID: params.ProfileID,
		Statuses:  params.Statuses,
		Limit:     params.Limit,
		Cursor:    params.Cursor,
	})
	if err != nil {
		return Result{}, err
	}

	if !hasNext {
		return Result{Offers: offers}, nil
	}

	return Result{
		Offers:     offers,
		NextCursor: nextCursor(offers[len(offers)-1]),
	}, nil
}

func nextCursor(last data.ProfileOffer) *data.ProfileOffersCursor {
	return &data.ProfileOffersCursor{ID: last.ID, RefreshedAt: last.RefreshedAt}
}
