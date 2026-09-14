package feed

const defaultLimit = 25

type feedRequest struct {
	Query           *string  `form:"query"`
	RegionId        *int     `form:"region_id"`
	PriceFrom       *float64 `form:"price_from"`
	PriceTo         *float64 `form:"price_to"`
	ReissueIncluded *bool    `form:"reissue_included"`
	CategoryIds     []int    `form:"category_ids"`
	Limit           int      `form:"limit"`
	Cursor          string   `form:"cursor"`
}

func newFeedRequest() feedRequest {
	return feedRequest{Limit: defaultLimit}
}
