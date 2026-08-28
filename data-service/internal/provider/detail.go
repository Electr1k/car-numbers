package provider

import (
	"context"
	"data-service/internal/domain"
)

// OfferDetailProvider - поставщик, умеющий догрузить деталку оффера
type OfferDetailProvider interface {
	FetchOfferDetail(ctx context.Context, offer domain.OfferWithNumber) (domain.OfferWithNumber, error)
}
