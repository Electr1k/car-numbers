package importprofile

import (
	"plate-service/internal/domain"
)

type Params struct {
	Provider          domain.Provider
	ProfileExternalID string
}
