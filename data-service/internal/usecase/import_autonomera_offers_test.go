package usecase

import (
	"context"
	"data-service/config"
	"data-service/internal/domain"
	"data-service/internal/provider"
	"data-service/internal/provider/autonomera"
	"errors"
	"io"
	"log/slog"
	"slices"
	"testing"
	"time"
)

// providerFunc - фейковый провайдер по образцу http.HandlerFunc
type providerFunc func(ctx context.Context, section autonomera.Section, offset int) (provider.FetchResult, error)

func (f providerFunc) FetchOffers(ctx context.Context, section autonomera.Section, offset int) (provider.FetchResult, error) {
	return f(ctx, section, offset)
}

type saverFunc func(ctx context.Context, items []domain.OfferWithNumber) (int, error)

func (f saverFunc) SaveBatch(ctx context.Context, items []domain.OfferWithNumber) (int, error) {
	return f(ctx, items)
}

func saveAll(_ context.Context, items []domain.OfferWithNumber) (int, error) {
	return len(items), nil
}

func testConfig() config.AutoNomeraConfig {
	return config.AutoNomeraConfig{BatchSize: 20}
}

func newUseCase(p OfferProvider, saver OfferSaver) *ImportAutonomeraOffersUseCase {
	return NewImportAutonomeraOffersUseCase(p, saver, testConfig(),
		slog.New(slog.NewTextHandler(io.Discard, nil)))
}

// resultWith - страница из count предложений, опубликованных в postedAt
func resultWith(t *testing.T, count int, postedAt time.Time, rowErrs ...error) provider.FetchResult {
	t.Helper()

	result := provider.FetchResult{RowsFound: count + len(rowErrs)}

	for range count {
		number, err := domain.NewNumber("а123аа77", domain.NumberTypeCar)
		if err != nil {
			t.Fatalf("build number: %v", err)
		}

		price := 1000.0

		offer, err := domain.NewOffer(
			number.Id,
			domain.ProviderAutonomera,
			"42",
			&price,
			domain.OfferStatusActive,
			nil,
			nil,
			nil,
			&postedAt,
			&postedAt,
			"https://example.com/42",
			"<tr></tr>",
			nil,
			nil,
		)
		if err != nil {
			t.Fatalf("build offer: %v", err)
		}

		result.Offers = append(result.Offers, domain.OfferWithNumber{Number: number, Offer: offer})
	}

	for i, rowErr := range rowErrs {
		result.RowErrors = append(result.RowErrors, provider.RowError{Index: i, Err: rowErr})
	}

	return result
}

func TestHandleStopsWhenFeedExhausted(t *testing.T) {
	now := time.Now()
	var offsets []int

	p := providerFunc(func(_ context.Context, _ autonomera.Section, offset int) (provider.FetchResult, error) {
		offsets = append(offsets, offset)
		if offset >= 40 {
			return provider.FetchResult{}, nil
		}

		return resultWith(t, 20, now), nil
	})

	var saved int
	saver := saverFunc(func(_ context.Context, items []domain.OfferWithNumber) (int, error) {
		saved += len(items)
		return len(items), nil
	})

	if err := newUseCase(p, saver).Handle(context.Background(), ImportParams{Section: autonomera.SectionActive}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if saved != 40 {
		t.Errorf("saved = %d, want 40", saved)
	}

	// Две полные страницы по 20, затем две пустые подряд - обе по BatchSize
	want := []int{0, 20, 40, 60}
	if len(offsets) != len(want) {
		t.Fatalf("offsets = %v, want %v", offsets, want)
	}
	for i, offset := range want {
		if offsets[i] != offset {
			t.Fatalf("offsets = %v, want %v", offsets, want)
		}
	}
}

func TestHandleStopsAtStopDate(t *testing.T) {
	old := time.Now().Add(-90 * 24 * time.Hour)

	var pages int
	p := providerFunc(func(_ context.Context, _ autonomera.Section, _ int) (provider.FetchResult, error) {
		pages++
		if pages == 1 {
			return resultWith(t, 20, time.Now()), nil
		}

		return resultWith(t, 20, old), nil
	})

	if err := newUseCase(p, saverFunc(saveAll)).
		Handle(context.Background(), ImportParams{Section: autonomera.SectionActive, StopAfter: 30 * 24 * time.Hour}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pages != 2 {
		t.Errorf("fetched %d pages, want 2", pages)
	}
}

func TestHandleIgnoresStopDateForArchive(t *testing.T) {
	old := time.Now().Add(-10 * 365 * 24 * time.Hour)

	var pages int
	p := providerFunc(func(_ context.Context, _ autonomera.Section, _ int) (provider.FetchResult, error) {
		pages++
		if pages > 3 {
			return provider.FetchResult{}, nil
		}

		return resultWith(t, 20, old), nil
	})

	// stopAfter = 0 - архив забирается целиком, старые даты не останавливают
	if err := newUseCase(p, saverFunc(saveAll)).
		Handle(context.Background(), ImportParams{Section: autonomera.SectionArchive}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pages != 5 {
		t.Errorf("fetched %d pages, want 5 (3 с данными + 2 пустые)", pages)
	}
}

func TestHandleStopsOnBrokenOffers(t *testing.T) {
	now := time.Now()
	drift := slices.Repeat([]error{provider.ErrBrokenOffer}, maxCountBrokenOffers+1)

	p := providerFunc(func(_ context.Context, _ autonomera.Section, _ int) (provider.FetchResult, error) {
		return resultWith(t, 14, now, drift...), nil
	})

	err := newUseCase(p, saverFunc(saveAll)).Handle(context.Background(), ImportParams{Section: autonomera.SectionActive})
	if !errors.Is(err, ErrTooManyBrokenOffers) {
		t.Fatalf("want ErrTooManyBrokenOffers, got %v", err)
	}
}

func TestHandleToleratesDriftOnTinyPage(t *testing.T) {
	now := time.Now()

	var pages int
	p := providerFunc(func(_ context.Context, _ autonomera.Section, _ int) (provider.FetchResult, error) {
		pages++
		if pages == 1 {
			// Одна строка, и та не разобралась: 100%, но выборка слишком мала
			return resultWith(t, 0, now, provider.ErrBrokenOffer), nil
		}

		return provider.FetchResult{}, nil
	})

	if err := newUseCase(p, saverFunc(saveAll)).
		Handle(context.Background(), ImportParams{Section: autonomera.SectionActive}); err != nil {
		t.Fatalf("tiny page must not abort the import, got %v", err)
	}
}

func TestHandleStopsWhenProviderClampsOffset(t *testing.T) {
	now := time.Now()

	// Сайт за пределами выдачи отдаёт последнюю страницу снова: ни пустых
	// страниц, ни ошибок - без потолка цикл был бы бесконечным
	p := providerFunc(func(_ context.Context, _ autonomera.Section, _ int) (provider.FetchResult, error) {
		return resultWith(t, 1, now), nil
	})

	err := newUseCase(p, saverFunc(saveAll)).Handle(context.Background(), ImportParams{Section: autonomera.SectionActive})
	if !errors.Is(err, ErrPageLimitExceeded) {
		t.Fatalf("want ErrPageLimitExceeded, got %v", err)
	}
}

func TestHandlePassesSectionToProvider(t *testing.T) {
	var got autonomera.Section

	p := providerFunc(func(_ context.Context, section autonomera.Section, _ int) (provider.FetchResult, error) {
		got = section
		return provider.FetchResult{}, nil
	})

	if err := newUseCase(p, saverFunc(saveAll)).
		Handle(context.Background(), ImportParams{Section: autonomera.SectionArchive}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != autonomera.SectionArchive {
		t.Errorf("provider got section %q, want %q", got, autonomera.SectionArchive)
	}
}

func TestHandleStopsOnCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	p := providerFunc(func(_ context.Context, _ autonomera.Section, _ int) (provider.FetchResult, error) {
		cancel()
		return resultWith(t, 20, time.Now()), nil
	})

	err := newUseCase(p, saverFunc(saveAll)).Handle(ctx, ImportParams{Section: autonomera.SectionActive})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("want context.Canceled, got %v", err)
	}
}

func TestHandlePropagatesProviderError(t *testing.T) {
	p := providerFunc(func(_ context.Context, _ autonomera.Section, _ int) (provider.FetchResult, error) {
		return provider.FetchResult{}, provider.ErrRateLimitExceeded
	})

	err := newUseCase(p, saverFunc(saveAll)).Handle(context.Background(), ImportParams{Section: autonomera.SectionActive})
	if !errors.Is(err, provider.ErrRateLimitExceeded) {
		t.Fatalf("want ErrRateLimitExceeded, got %v", err)
	}
}

func TestHandlePropagatesSaverError(t *testing.T) {
	saveFailed := errors.New("disk on fire")

	p := providerFunc(func(_ context.Context, _ autonomera.Section, _ int) (provider.FetchResult, error) {
		return resultWith(t, 20, time.Now()), nil
	})
	saver := saverFunc(func(context.Context, []domain.OfferWithNumber) (int, error) {
		return 0, saveFailed
	})

	err := newUseCase(p, saver).Handle(context.Background(), ImportParams{Section: autonomera.SectionActive})
	if !errors.Is(err, saveFailed) {
		t.Fatalf("want saveFailed, got %v", err)
	}
}

func TestHandleToleratesDriftUpToLimit(t *testing.T) {
	now := time.Now()
	drift := slices.Repeat([]error{provider.ErrBrokenOffer}, maxCountBrokenOffers)

	var pages int
	p := providerFunc(func(_ context.Context, _ autonomera.Section, _ int) (provider.FetchResult, error) {
		pages++
		if pages == 1 {
			return resultWith(t, 15, now, drift...), nil
		}

		return provider.FetchResult{}, nil
	})

	// Ровно на пороге импорт продолжается: авария начинается строго выше
	if err := newUseCase(p, saverFunc(saveAll)).
		Handle(context.Background(), ImportParams{Section: autonomera.SectionActive}); err != nil {
		t.Fatalf("drift at the limit must not abort the import, got %v", err)
	}
}

func TestHandleStartsFromGivenOffset(t *testing.T) {
	var first = -1

	p := providerFunc(func(_ context.Context, _ autonomera.Section, offset int) (provider.FetchResult, error) {
		if first < 0 {
			first = offset
		}

		return provider.FetchResult{}, nil
	})

	if err := newUseCase(p, saverFunc(saveAll)).
		Handle(context.Background(), ImportParams{Section: autonomera.SectionActive, StartOffset: 4000}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if first != 4000 {
		t.Errorf("first offset = %d, want 4000", first)
	}
}

func TestHandleRespectsMaxPages(t *testing.T) {
	now := time.Now()

	var pages int
	p := providerFunc(func(_ context.Context, _ autonomera.Section, _ int) (provider.FetchResult, error) {
		pages++
		return resultWith(t, 20, now), nil
	})

	err := newUseCase(p, saverFunc(saveAll)).
		Handle(context.Background(), ImportParams{Section: autonomera.SectionActive, MaxPages: 3})

	if !errors.Is(err, ErrPageLimitExceeded) {
		t.Fatalf("want ErrPageLimitExceeded, got %v", err)
	}
	if pages != 3 {
		t.Errorf("fetched %d pages, want 3", pages)
	}
}

func TestHandleRejectsNegativeParams(t *testing.T) {
	cases := map[string]ImportParams{
		"отрицательный offset":     {Section: autonomera.SectionActive, StartOffset: -1},
		"отрицательный max pages":  {Section: autonomera.SectionActive, MaxPages: -1},
		"отрицательный stop after": {Section: autonomera.SectionActive, StopAfter: -time.Hour},
	}

	var called bool
	p := providerFunc(func(_ context.Context, _ autonomera.Section, _ int) (provider.FetchResult, error) {
		called = true
		return provider.FetchResult{}, nil
	})

	for name, params := range cases {
		t.Run(name, func(t *testing.T) {
			if err := newUseCase(p, saverFunc(saveAll)).
				Handle(context.Background(), params); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}

	if called {
		t.Error("provider was called despite invalid params")
	}
}
