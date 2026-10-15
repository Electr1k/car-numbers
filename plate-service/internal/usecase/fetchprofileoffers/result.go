package fetchprofileoffers

import (
	"plate-service/internal/domain/data"
)

// Result - результат выборки предложений продавца
type Result struct {
	Offers     []data.ProfileOffer
	NextCursor *data.ProfileOffersCursor
}
