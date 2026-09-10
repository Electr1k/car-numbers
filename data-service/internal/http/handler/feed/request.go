package feed

const (
	defaultLimit = 25
)

type feedRequest struct {
	Limit  int     `form:"limit"`
	Cursor *string `form:"cursor"`
}

func newFeedRequest() feedRequest {
	return feedRequest{Limit: defaultLimit, Cursor: nil}
}
