package feed

const (
	defaultLimit  = 25
	defaultOffset = 0
)

type feedRequest struct {
	Limit  int `form:"limit" validate:"min=0,max=25"`
	Offset int `form:"offset" validate:"min=0"`
}

func newFeedRequest() feedRequest {
	return feedRequest{Limit: defaultLimit, Offset: defaultOffset}
}
