package fetchprofile

import (
	"context"
	"plate-service/internal/domain/data"

	"github.com/google/uuid"
)

type profileStore interface {
	GetProfileDetail(ctx context.Context, id uuid.UUID) (*data.ProfileDetail, error)
	GetRelatedProfiles(ctx context.Context, id uuid.UUID) ([]data.RelatedProfile, error)
}

// UseCase - возвращает профиль продавца
type UseCase struct {
	repository profileStore
}

func New(repository profileStore) *UseCase {
	return &UseCase{repository: repository}
}

// Handle - профиль продавца вместе с аккаунтами, вероятно принадлежащими ему же
func (uc *UseCase) Handle(ctx context.Context, id uuid.UUID) (*data.ProfileDetail, error) {
	profile, err := uc.repository.GetProfileDetail(ctx, id)
	if err != nil {
		return nil, err
	}

	related, err := uc.repository.GetRelatedProfiles(ctx, id)
	if err != nil {
		return nil, err
	}
	profile.Related = related

	return profile, nil
}
