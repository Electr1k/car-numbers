package feed

const (
	defaultLimit  = 25
	defaultOffset = 0
)

type feedRequest struct {
	Limit  int `form:"limit"`
	Offset int `form:"offset"`
}

func newFeedRequest() feedRequest {
	return feedRequest{Limit: defaultLimit, Offset: defaultOffset}
}
