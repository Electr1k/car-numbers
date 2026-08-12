package usecase

import (
	"context"
	"core/config"
	"core/internal/domain"
	"core/internal/providers"
	"core/internal/providers/autonomera"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

// OfferProvider - источник предложений
type OfferProvider interface {
	FetchOffers(ctx context.Context, section autonomera.Section, offset int) (providers.FetchResult, error)
}

// OfferSaver сохраняет батч атомарно
type OfferSaver interface {
	SaveBatch(ctx context.Context, items []domain.OfferWithNumber) (int, error)
}

const (
	// maxEmptyPages - сколько подряд пустых страниц считать концом выдачи
	maxEmptyPages = 2

	// defaultMaxPages - потолок страниц, когда он не задан явно
	//
	// Совсем без потолка импорт оставлять нельзя: сайт за пределами выдачи
	// может отдавать последнюю страницу снова и снова, и тогда ни пустых
	// страниц, ни ошибок не будет - цикл станет бесконечным
	defaultMaxPages = 10000

	// maxCountBrokenOffers - максимальное количество битых офферов в батче
	maxCountBrokenOffers = 5
)

// ImportParams - границы одного прогона импорта
//
// Нулевое поле означает «не ограничивать»: так можно догрузить недостающий
// участок выдачи или выкачать раздел целиком
type ImportParams struct {
	// StartOffset - с какой позиции выдачи начинать
	StartOffset int

	// MaxPages - потолок страниц за прогон
	// Ноль подставляет defaultMaxPages, а не снимает потолок совсем
	MaxPages int

	// StopAfter - не забирать публикации старше этого возраста
	// Ноль - без ограничения по дате, так забирают архив
	StopAfter time.Duration
}

func (p ImportParams) validate() error {
	switch {
	case p.StartOffset < 0:
		return fmt.Errorf("start offset must not be negative, got %d", p.StartOffset)
	case p.MaxPages < 0:
		return fmt.Errorf("max pages must not be negative, got %d", p.MaxPages)
	case p.StopAfter < 0:
		return fmt.Errorf("stop after must not be negative, got %s", p.StopAfter)
	}

	return nil
}

// pageLimit - потолок страниц с подставленным умолчанием
func (p ImportParams) pageLimit() int {
	if p.MaxPages == 0 {
		return defaultMaxPages
	}

	return p.MaxPages
}

// stopDate - нижняя граница импорта по дате публикации
// Нулевое время означает, что проверка не выполняется
func (p ImportParams) stopDate() time.Time {
	if p.StopAfter == 0 {
		return time.Time{}
	}

	return time.Now().Add(-p.StopAfter)
}

var (
	// ErrTooManySchemaErrors - Превышен лимит неразобранных офферов
	ErrTooManySchemaErrors = errors.New("too many schema errors on page")

	// ErrPageLimitExceeded - выдача не кончилась за отведённое число страниц
	ErrPageLimitExceeded = errors.New("page limit exceeded")
)

// ImportAutonomeraOffersUseCase - постраничный импорт раздела autonomera777
type ImportAutonomeraOffersUseCase struct {
	provider   OfferProvider
	repository OfferSaver
	config     config.AutoNomeraConfig
	logger     *slog.Logger
}

func NewImportAutonomeraOffersUseCase(
	provider OfferProvider,
	repository OfferSaver,
	config config.AutoNomeraConfig,
	logger *slog.Logger,
) *ImportAutonomeraOffersUseCase {
	return &ImportAutonomeraOffersUseCase{
		provider:   provider,
		repository: repository,
		config:     config,
		logger:     logger,
	}
}

// Handle постранично забирает предложения раздела и складывает их в базу
// Границы прогона задаёт params, см. ImportParams
func (uc *ImportAutonomeraOffersUseCase) Handle(ctx context.Context, section autonomera.Section, params ImportParams) error {
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
		result     = "failed"
	)

	logger := uc.logger.With("section", section)

	logger.Info("import started",
		"start_offset", offset,
		"stop_date", stopDate,
		"page_limit", pageLimit)

	// Итоговый лог
	defer func() {
		logger.Info("import finished",
			"outcome", result,
			"offset", offset,
			"saved", saved,
			"skipped", skipped)
	}()

	for page := 1; page <= pageLimit; page++ {
		if err := ctx.Err(); err != nil {
			result = "cancelled"
			return err
		}

		offers, err := uc.provider.FetchOffers(ctx, section, offset)
		if err != nil {
			return fmt.Errorf("fetch page at offset %d: %w", offset, err)
		}

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

		if err := checkSchemaDrift(offers); err != nil {
			logger.Error("stopping import",
				"page", page,
				"offset", offset,
				"error", err,
				"reasons", offers.Reasons())

			return err
		}

		savedOnPage, err := uc.repository.SaveBatch(ctx, offers.Offers)
		if err != nil {
			return fmt.Errorf("save batch at offset %d: %w", offset, err)
		}
		saved += savedOnPage
		skipped += len(offers.RowErrors)

		logPage(logger, page, offset, offers, savedOnPage)

		if checkStopDate(offers, stopDate) {
			result = "stop date reached"
			return nil
		}

		offset += offers.RowsFound
	}

	return fmt.Errorf("%w: %d pages, stopped at offset %d", ErrPageLimitExceeded, pageLimit, offset)
}

// checkStopDate - проверка достижения нижней даты импорта
func checkStopDate(result providers.FetchResult, stopDate time.Time) bool {
	if stopDate.IsZero() {
		return false
	}

	oldest := oldestPostedAt(result.Offers)

	return !oldest.IsZero() && oldest.Before(stopDate)
}

// checkSchemaDrift - не слишком ли много строк на странице не разобралось
//
// Порог абсолютный, а не в долях: на хвосте выдачи страница бывает из одной
// строки, и процент там ничего не значит
func checkSchemaDrift(result providers.FetchResult) error {
	countBrokenOffers := result.SchemaDriftCount()
	if countBrokenOffers <= maxCountBrokenOffers {
		return nil
	}

	return fmt.Errorf("%w: %d of %d rows (limit %d)",
		ErrTooManySchemaErrors, countBrokenOffers, result.RowsFound, maxCountBrokenOffers)
}

func logPage(logger *slog.Logger, page, offset int, result providers.FetchResult, saved int) {
	attrs := []any{
		"page", page,
		"offset", offset,
		"found", result.RowsFound,
		"saved", saved,
		"skipped", len(result.RowErrors),
	}

	if len(result.RowErrors) == 0 {
		logger.Info("page processed", attrs...)
		return
	}

	logger.Warn("page processed with skipped rows",
		append(attrs, "reasons", result.Reasons())...)
}

// oldestPostedAt возвращает самую раннюю дату публикации в батче
func oldestPostedAt(offers []domain.OfferWithNumber) time.Time {
	var oldest time.Time

	for _, item := range offers {
		if item.Offer.PostedAt == nil {
			continue
		}

		if oldest.IsZero() || item.Offer.PostedAt.Before(oldest) {
			oldest = *item.Offer.PostedAt
		}
	}

	return oldest
}
