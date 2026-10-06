package importprofile

import (
	"context"
	"errors"
	"log/slog"
	"plate-service/internal/domain"
	"plate-service/internal/provider"
)

// resolver - выбирает провайдера профилей по имени поставщика
type resolver interface {
	ResolveProfileProvider(name domain.Provider) (provider.ProfileProvider, error)
}

type profileRepository interface {
	GetUserByContacts(ctx context.Context, phone *string, email *string) (*domain.User, error)

	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)

	UpsertProfile(ctx context.Context, profile *domain.Profile) (*domain.Profile, error)
}

type feature interface {
	Enabled(ctx context.Context, key domain.FeatureKey) (bool, error)
}

// UseCase - импорт профиля и привязка к доменному пользователю
type UseCase struct {
	resolver   resolver
	repository profileRepository
	features   feature
	logger     *slog.Logger
}

func New(
	resolver resolver,
	repository profileRepository,
	features feature,
	logger *slog.Logger,
) *UseCase {
	return &UseCase{
		resolver:   resolver,
		repository: repository,
		features:   features,
		logger:     logger,
	}
}

// Handle - импортирует профиль и сохраняет его в базу
func (uc *UseCase) Handle(ctx context.Context, params Params) error {
	logger := uc.logger.With("profile_external_id", params.ProfileExternalID, "provider", params.Provider)

	enabled, err := uc.features.Enabled(ctx, domain.FeatureKeyImportProfile)
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}

	logger.Info("import profile started")

	profileProvider, err := uc.resolver.ResolveProfileProvider(params.Provider)
	if err != nil {
		return err
	}

	profile, err := profileProvider.FetchProfile(ctx, params.ProfileExternalID)
	if err != nil {
		return err
	}

	user, err := uc.repository.GetUserByContacts(ctx, profile.Phone, profile.Email)
	if errors.Is(err, domain.ErrUserNotFound) {
		user, err = domain.NewUser(profile.Name, profile.Phone, profile.Email)
		if err != nil {
			return err
		}

		user, err = uc.repository.CreateUser(ctx, user)
	}
	if err != nil {
		return err
	}

	profile.UserID = &user.ID

	if _, err := uc.repository.UpsertProfile(ctx, &profile); err != nil {
		return err
	}

	logger.Info("import profile finished", "user_id", user.ID)

	return nil
}
