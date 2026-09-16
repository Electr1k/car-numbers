package fetchplates

import (
	"data-service/internal/domain/data"
)

// Result - результат выборки свежих номеров
type Result struct {
	Plates     []data.FeedPlate
	NextCursor *data.FeedCursor
}
