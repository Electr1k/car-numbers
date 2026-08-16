package usecase

import (
	"context"
	"data-service/internal/domain"
	"errors"
	"fmt"
	"log/slog"
)

type FeatureStorage interface {
	GetFeatureByKey(ctx context.Context, key domain.FeatureKey) (*domain.Feature, error)
}

func featureEnabled(
	ctx context.Context,
	storage FeatureStorage,
	key domain.FeatureKey,
	logger *slog.Logger,
) (bool, error) {
	feature, err := storage.GetFeatureByKey(ctx, key)
	if errors.Is(err, domain.ErrFeatureNotFound) {
		logger.Warn("feature not found, treating as disabled", "feature", key)
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("get feature %s: %w", key, err)
	}

	if !feature.Active {
		logger.Info("feature disabled, skipping", "feature", key)
	}

	return feature.Active, nil
}
