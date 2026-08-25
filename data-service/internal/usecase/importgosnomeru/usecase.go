package importgosnomeru

import (
	"context"
	"data-service/internal/domain"
	"data-service/internal/provider"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type offerProvider interface {
	FetchOffers(ctx context.Context, page int) (provider.FetchResult, error)
}

type offerSaver interface {
	SaveBatch(ctx context.Context, items []domain.OfferWithNumber) ([]domain.OfferWithNumber, error)
}

type detailDispatcher interface {
	DispatchImportOfferDetail(ctx context.Context, offerId uuid.UUID) (bool, error)
}

type feature interface {
	Enabled(ctx context.Context, key domain.FeatureKey) (bool, error)
}

const (
	// maxCountBrokenOffers - максимальное количество битых офферов в батче
	maxCountBrokenOffers = 0
)

var (
	// ErrTooManyBrokenOffers - превышен лимит неразобранных офферов
	ErrTooManyBrokenOffers = errors.New("too many broken offers on page")

	// ErrPageLimitExceeded - выдача не кончилась за отведённое число страниц
	ErrPageLimitExceeded = errors.New("page limit exceeded")
)

// UseCase - постраничный импорт выдачи gosnomeru
type UseCase struct {
	provider         offerProvider
	detailDispatcher detailDispatcher
	repository       offerSaver
	features         feature
	logger           *slog.Logger
}

func New(
	provider offerProvider,
	detailDispatcher detailDispatcher,
	repository offerSaver,
	features feature,
	logger *slog.Logger,
) *UseCase {
	return &UseCase{
		provider:         provider,
		detailDispatcher: detailDispatcher,
		repository:       repository,
		features:         features,
		logger:           logger,
	}
}

// Handle - собирает предложения поставщика и сохраняет их в базу
func (uc *UseCase) Handle(ctx context.Context, params Params) error {
	if err := params.validate(); err != nil {
		return err
	}

	var (
		startPage  = params.StartPage
		stopDate   = params.stopDate()
		pageLimit  = params.pageLimit()
		page       = startPage
		saved      int
		skipped    int
		dispatched int
		result     = "failed"
	)

	logger := uc.logger

	enabled, err := uc.features.Enabled(ctx, domain.FeatureKeyImportGosnomeruOffers)
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}

	logger.Info("import started", "start_page", startPage, "stop_date", stopDate, "page_limit", pageLimit)

	// Итоговый лог
	defer func() {
		logger.Info("import finished", "result", result, "page", page, "saved", saved, "skipped", skipped, "dispatched", dispatched)
	}()

	for taken := range pageLimit {
		page = startPage + taken

		if err := ctx.Err(); err != nil {
			result = "cancelled"
			return err
		}

		// Получение страницы
		offers, err := uc.provider.FetchOffers(ctx, page)
		if err != nil {
			return fmt.Errorf("fetch page %d: %w", page, err)
		}

		// Проверка максимального количества кривых офферов
		if err := checkBrokenOffers(offers); err != nil {
			logger.Error("stopping import",
				"page", page,
				"error", err,
				"errors", offers.ErrorMessages())

			return err
		}

		// Сохранение офферов
		savedOnPage, err := uc.repository.SaveBatch(ctx, offers.Offers)
		if err != nil {
			return fmt.Errorf("save batch at page %d: %w", page, err)
		}

		dispatchedOnPage := 0
		// Вызов асинхронного импорта деталки для офферов без них
		for _, offer := range savedOnPage {
			if offer.Offer.RawDetailed != nil {
				continue
			}

			success, err := uc.detailDispatcher.DispatchImportOfferDetail(ctx, offer.Offer.Id)
			if err != nil {
				return fmt.Errorf("dispatch offer details at page %d: %w", page, err)
			}
			if success {
				dispatchedOnPage++
			}
		}

		saved += len(savedOnPage)
		skipped += len(offers.RowErrors)
		dispatched += dispatchedOnPage

		logPage(logger, page, offers, len(savedOnPage), dispatchedOnPage)

		// Проверка нижнего порога по дате
		if checkStopDate(offers, stopDate) {
			result = "stop date reached"
			return nil
		}

		// Поставщик сам сообщает, сколько всего страниц
		if offers.TotalPages > 0 && page >= offers.TotalPages {
			result = "feed exhausted"
			return nil
		}
	}

	return fmt.Errorf("%w: %d pages, stopped at page %d", ErrPageLimitExceeded, pageLimit, page)
}

// checkStopDate - проверка достижения нижней даты импорта
func checkStopDate(result provider.FetchResult, stopDate time.Time) bool {
	if stopDate.IsZero() {
		return false
	}

	oldest := result.OldestPostedAt()

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
func logPage(logger *slog.Logger, page int, result provider.FetchResult, saved, dispatched int) {
	attrs := []any{
		"page", page,
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
