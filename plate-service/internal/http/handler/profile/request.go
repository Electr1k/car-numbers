package profile

const defaultOffersLimit = 30

type profileOffersRequest struct {
	Statuses []string `form:"statuses"`
	Limit    int      `form:"limit"`
	Cursor   string   `form:"cursor"`
}

func newProfileOffersRequest() profileOffersRequest {
	return profileOffersRequest{Limit: defaultOffersLimit}
}
