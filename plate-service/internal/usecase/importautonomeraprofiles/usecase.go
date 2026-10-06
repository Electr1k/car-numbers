package importautonomeraprofiles

import (
	"context"
	"fmt"
	"log/slog"
	"plate-service/config"
	"plate-service/internal/domain"
	"slices"
)

type profileProvider interface {
	FetchProfileExternalIds(ctx context.Context, offset int) ([]string, error)
}

type profileDispatcher interface {
	DispatchImportProfile(ctx context.Context, provider domain.Provider, profileExternalId string) (bool, error)
}

type profileRepository interface {
	GetExistingProfileExternalIDs(ctx context.Context, provider domain.Provider, externalIDs []string) ([]string, error)
}

type feature interface {
	Enabled(ctx context.Context, key domain.FeatureKey) (bool, error)
}

// UseCase - постраничный импорт профилей autonomera777
type UseCase struct {
	provider          profileProvider
	profileDispatcher profileDispatcher
	repository        profileRepository
	features          feature
	config            config.AutoNomeraConfig
	logger            *slog.Logger
}

func New(
	provider profileProvider,
	profileDispatcher profileDispatcher,
	repository profileRepository,
	features feature,
	config config.AutoNomeraConfig,
	logger *slog.Logger,
) *UseCase {
	return &UseCase{
		provider:          provider,
		profileDispatcher: profileDispatcher,
		repository:        repository,
		features:          features,
		config:            config,
		logger:            logger,
	}
}

// Handle - собирает профили поставщика и вызываем их импорт
func (uc *UseCase) Handle(ctx context.Context) error {
	var (
		offset     = 0
		dispatched int
		result     = "failed"
	)

	logger := uc.logger

	enabled, err := uc.features.Enabled(ctx, domain.FeatureKeyImportAutonomeraProfiles)
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}
	logger.Info("import profiles started", "start_offset", offset)

	// Итоговый лог
	defer func() {
		logger.Info("import profiles finished", "result", result, "offset", offset, "dispatched", dispatched)
	}()

	for page := 1; true; page++ {
		if err := ctx.Err(); err != nil {
			result = "cancelled"
			return err
		}

		// Получение оффера
		ids, err := uc.provider.FetchProfileExternalIds(ctx, offset)
		if err != nil {
			return fmt.Errorf("fetch page at offset %d: %w", offset, err)
		}
		if len(ids) == 0 {
			break
		}

		existing, err := uc.repository.GetExistingProfileExternalIDs(ctx, domain.ProviderAutonomera, ids)
		if err != nil {
			return fmt.Errorf("get existing profiles at offset %d: %w", offset, err)
		}

		dispatchedOnPage := 0
		// Вызов асинхронного импорта профиля
		for _, id := range ids {
			if slices.Contains(existing, id) {
				continue
			}

			success, err := uc.profileDispatcher.DispatchImportProfile(ctx, domain.ProviderAutonomera, id)
			if err != nil {
				return fmt.Errorf("dispatch profile at offset %d: %w", offset, err)
			}
			if success {
				dispatchedOnPage++
			}
		}
		dispatched += dispatchedOnPage

		logPage(logger, page, offset, ids, dispatchedOnPage)

		offset += len(ids)
	}

	result = "success"

	return nil
}

// logPage - итог по обработанной странице
func logPage(logger *slog.Logger, page, offset int, ids []string, dispatched int) {
	attrs := []any{
		"page", page,
		"offset", offset,
		"found", len(ids),
		"dispatched", dispatched,
	}

	logger.Info("page processed", attrs...)
}
