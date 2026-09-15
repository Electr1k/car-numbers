package valuation

import (
	"context"
	"core-service/internal/service/ml"
	"core-service/internal/service/plate"
	"log/slog"
)

type valuationProvider interface {
	Predict(ctx context.Context, number string) (*ml.PredictResponse, error)
}

type regionsProvider interface {
	FetchRegions(ctx context.Context) (*plate.FetchRegionsResponse, error)
}

// UseCase - оценивает стоимость номера
type UseCase struct {
	valuationProvider valuationProvider
	regionsProvider   regionsProvider
	logger            *slog.Logger
}

func New(
	valuationProvider valuationProvider,
	regionsProvider regionsProvider,
	logger *slog.Logger,
) *UseCase {
	return &UseCase{
		valuationProvider: valuationProvider,
		regionsProvider:   regionsProvider,
		logger:            logger,
	}
}

// Handle - предсказывает цену
func (uc *UseCase) Handle(ctx context.Context, number string) (*Result, error) {
	prediction, err := uc.valuationProvider.Predict(ctx, number)
	if err != nil {
		return nil, err
	}

	breakdownItems := make([]BreakdownItem, 0, len(prediction.Breakdown.Items))
	for _, item := range prediction.Breakdown.Items {
		breakdownItems = append(breakdownItems, BreakdownItem{
			Code:       item.Code,
			Value:      item.Value,
			Title:      item.Title,
			Multiplier: item.Multiplier,
			Exact:      item.Exact,
		})
	}

	return &Result{
		Number: prediction.Number,
		Region: uc.findRegion(ctx, number),
		Price: Price{
			P25: prediction.Price.P25,
			P50: prediction.Price.P50,
			P75: prediction.Price.P75,
		},
		Confidence: prediction.Confidence,
		Breakdown: Breakdown{
			Base:  prediction.Breakdown.Base,
			Items: breakdownItems,
		},
	}, nil
}

func (uc *UseCase) findRegion(ctx context.Context, number string) *Region {
	regionCode := string([]rune(number)[6:])

	regions, err := uc.regionsProvider.FetchRegions(ctx)
	if err != nil {
		uc.logger.WarnContext(ctx, "fetch regions for valuation failed", "error", err, "number", number)
		return nil
	}

	for _, r := range regions.Items {
		for _, code := range r.Codes {
			if code == regionCode {
				return &Region{Code: code, Name: r.Name}
			}
		}
	}

	return nil
}
