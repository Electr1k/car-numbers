package provider

import (
	"context"
	"plate-service/internal/domain"
)

// OfferDetailProvider - поставщик, умеющий догрузить деталку оффера
type OfferDetailProvider interface {
	FetchOfferDetail(ctx context.Context, offer domain.OfferWithPlate) (domain.OfferWithPlate, error)
}
