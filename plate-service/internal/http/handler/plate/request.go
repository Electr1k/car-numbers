package plate

const defaultLimit = 25

type platesRequest struct {
	Query           *string  `form:"query"`
	RegionId        *int     `form:"region_id"`
	PriceFrom       *float64 `form:"price_from"`
	PriceTo         *float64 `form:"price_to"`
	ReissueIncluded *bool    `form:"reissue_included"`
	CategoryIds     []string `form:"category_ids"`
	Sort            string   `form:"sort"`
	Limit           int      `form:"limit"`
	Cursor          string   `form:"cursor"`
}

func newPlatesRequest() platesRequest {
	return platesRequest{Limit: defaultLimit}
}
