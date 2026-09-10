package fetchfeednumbers

import (
	"data-service/internal/domain/data"
)

// Result - результат выборки свежих номеров
type Result struct {
	Numbers []data.FeedNumber
	Cursor  *string
}
