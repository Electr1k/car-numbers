package profile

const defaultOffersLimit = 30

type fetchProfileOffersRequest struct {
	Statuses []string `form:"statuses"`
	Limit    int      `form:"limit"`
	Cursor   string   `form:"cursor"`
}

func newFetchProfileOffersRequest() fetchProfileOffersRequest {
	return fetchProfileOffersRequest{Limit: defaultOffersLimit}
}
