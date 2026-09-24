package provider

import (
	"context"
	"plate-service/internal/domain"
)

// OfferDetailProvider - поставщик, умеющий догрузить деталку оффера
type OfferDetailProvider interface {
	FetchOfferDetail(ctx context.Context, offer domain.OfferWithPlate) (domain.OfferWithPlate, error)
}

// UserProvider - поставщик, умеющий импортировать профиль
type UserProvider interface {
	FetchUser(ctx context.Context, id string) (domain.Profile, error)
}
