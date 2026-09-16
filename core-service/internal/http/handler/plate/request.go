package plate

const defaultLimit = 25

type fetchPlatesRequest struct {
	Query           *string  `form:"query"`
	RegionID        *int     `form:"region_id"`
	PriceFrom       *float64 `form:"price_from"`
	PriceTo         *float64 `form:"price_to"`
	ReissueIncluded *bool    `form:"reissue_included"`
	CategoryIDs     []int    `form:"category_ids"`
	Sort            string   `form:"sort"`
	Limit           int      `form:"limit"`
	Cursor          string   `form:"cursor"`
}

func newFetchPlatesRequest() fetchPlatesRequest {
	return fetchPlatesRequest{Limit: defaultLimit}
}
