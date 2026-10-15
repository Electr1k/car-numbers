package fetchprofileoffers

import (
	"context"
	"core-service/internal/service/plate"
)

type offersProvider interface {
	FetchProfileOffers(ctx context.Context, params plate.FetchProfileOffersParams) (*plate.FetchProfileOffersResponse, error)
}

// UseCase - возвращает предложения продавца
type UseCase struct {
	plateClient offersProvider
}

func New(
	plateClient offersProvider,
) *UseCase {
	return &UseCase{
		plateClient: plateClient,
	}
}

// Handle - возвращает страницу предложений продавца
func (uc *UseCase) Handle(ctx context.Context, params Params) (*Result, error) {
	if err := params.validate(); err != nil {
		return nil, err
	}

	offers, err := uc.plateClient.FetchProfileOffers(ctx, params.toClientParams())
	if err != nil {
		return nil, err
	}

	items := make([]Offer, 0, len(offers.Items))
	for _, o := range offers.Items {
		items = append(items, Offer(o))
	}

	return &Result{
		Items:      items,
		NextCursor: offers.NextCursor,
	}, nil
}
