package importautonomera

import (
	"context"
	"data-service/config"
	"data-service/internal/domain"
	"data-service/internal/provider"
	"data-service/internal/provider/autonomera"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type offerProvider interface {
	FetchOffers(ctx context.Context, section autonomera.Section, offset int) (provider.FetchResult, error)
}

type offerSaver interface {
	SaveBatch(ctx context.Context, items []domain.OfferWithNumber) ([]domain.OfferWithNumber, error)
}

type detailDispatcher interface {
	DispatchImportOfferDetail(ctx context.Context, offerId uuid.UUID, provider domain.Provider) (bool, error)
}

type feature interface {
	Enabled(ctx context.Context, key domain.FeatureKey) (bool, error)
}

const (
	// maxEmptyPages - сколько подряд пустых страниц считать концом выдачи
	maxEmptyPages = 2

	// maxCountBrokenOffers - максимальное количество битых офферов в батче
	maxCountBrokenOffers = 5
)

var (
	// ErrTooManyBrokenOffers - превышен лимит неразобранных офферов
	ErrTooManyBrokenOffers = errors.New("too many broken offers on page")

	// ErrPageLimitExceeded - выдача не кончилась за отведённое число страниц
	ErrPageLimitExceeded = errors.New("page limit exceeded")
)

// UseCase - постраничный импорт раздела autonomera777
type UseCase struct {
	provider         offerProvider
	detailDispatcher detailDispatcher
	repository       offerSaver
	features         feature
	config           config.AutoNomeraConfig
	logger           *slog.Logger
}

func New(
	provider offerProvider,
	detailDispatcher detailDispatcher,
	repository offerSaver,
	features feature,
	config config.AutoNomeraConfig,
	logger *slog.Logger,
) *UseCase {
	return &UseCase{
		provider:         provider,
		detailDispatcher: detailDispatcher,
		repository:       repository,
		features:         features,
		config:           config,
		logger:           logger,
	}
}

// Handle - собирает предложения поставщика и сохраняет их в базу
func (uc *UseCase) Handle(ctx context.Context, params Params) error {
	if err := params.validate(); err != nil {
		return err
	}

	var (
		offset     = params.StartOffset
		stopDate   = params.stopDate()
		pageLimit  = params.pageLimit()
		emptyPages int
		saved      int
		skipped    int
		dispatched int
		result     = "failed"
	)

	logger := uc.logger.With("section", params.Section)

	enabled, err := uc.features.Enabled(ctx, domain.FeatureKeyImportAutonomeraOffers)
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}

	logger.Info("import started", "start_offset", offset, "stop_date", stopDate, "page_limit", pageLimit)

	// Итоговый лог
	defer func() {
		logger.Info("import finished", "result", result, "offset", offset, "saved", saved, "skipped", skipped, "dispatched", dispatched)
	}()

	for page := 1; page <= pageLimit; page++ {
		if err := ctx.Err(); err != nil {
			result = "cancelled"
			return err
		}

		// Получение оффера
		offers, err := uc.provider.FetchOffers(ctx, params.Section, offset)
		if err != nil {
			return fmt.Errorf("fetch page at offset %d: %w", offset, err)
		}

		// Проверка достижения конца списка
		if offers.RowsFound == 0 {
			emptyPages++
			if emptyPages >= maxEmptyPages {
				result = "feed exhausted"
				return nil
			}

			logger.Warn("empty page", "page", page, "offset", offset)
			offset += uc.config.BatchSize

			continue
		}
		emptyPages = 0

		// Проверка максимального количества кривых офферов
		if err := checkBrokenOffers(offers); err != nil {
			logger.Error("stopping import",
				"page", page,
				"offset", offset,
				"error", err,
				"errors", offers.ErrorMessages())

			return err
		}

		// Сохранение офферов
		savedOnPage, err := uc.repository.SaveBatch(ctx, offers.Offers)
		if err != nil {
			return fmt.Errorf("save batch at offset %d: %w", offset, err)
		}

		dispatchedOnPage := 0
		// Вызов асинхронного импорта деталки для офферов без них
		for _, offer := range savedOnPage {
			if offer.Offer.RawDetailed != nil {
				continue
			}

			success, err := uc.detailDispatcher.DispatchImportOfferDetail(ctx, offer.Offer.Id, offer.Offer.Provider)
			if err != nil {
				return fmt.Errorf("dispatch offer details at offset %d: %w", offset, err)
			}
			if success {
				dispatchedOnPage++
			}
		}

		saved += len(savedOnPage)
		skipped += len(offers.RowErrors)
		dispatched += dispatchedOnPage

		logPage(logger, page, offset, offers, len(savedOnPage), dispatchedOnPage)

		// Проверка нижнего порога по дате
		if checkStopDate(offers, stopDate) {
			result = "stop date reached"
			return nil
		}

		offset += offers.RowsFound
	}

	return fmt.Errorf("%w: %d pages, stopped at offset %d", ErrPageLimitExceeded, pageLimit, offset)
}

// checkStopDate - проверка достижения нижней даты импорта
func checkStopDate(result provider.FetchResult, stopDate time.Time) bool {
	if stopDate.IsZero() {
		return false
	}

	oldest := result.OldestRefreshedAt()

	return !oldest.IsZero() && oldest.Before(stopDate)
}

// checkBrokenOffers - не слишком ли много строк на странице не разобралось
func checkBrokenOffers(result provider.FetchResult) error {
	countBrokenOffers := result.BrokenOfferCount()
	if countBrokenOffers <= maxCountBrokenOffers {
		return nil
	}

	return fmt.Errorf("%w: %d of %d rows (limit %d)",
		ErrTooManyBrokenOffers, countBrokenOffers, result.RowsFound, maxCountBrokenOffers)
}

// logPage - итог по обработанной странице
func logPage(logger *slog.Logger, page, offset int, result provider.FetchResult, saved, dispatched int) {
	attrs := []any{
		"page", page,
		"offset", offset,
		"found", result.RowsFound,
		"saved", saved,
		"dispatched", dispatched,
		"skipped", len(result.RowErrors),
	}

	if len(result.RowErrors) == 0 {
		logger.Info("page processed", attrs...)
		return
	}

	logger.Warn("page processed with skipped rows",
		append(attrs, "errors", result.ErrorMessages())...)
}
