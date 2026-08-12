package autonomera

import (
	"context"
	"core/internal/domain"
	"core/internal/providers"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStatusForSection(t *testing.T) {
	cases := []struct {
		section Section
		want    domain.OfferStatus
	}{
		{SectionActive, domain.OfferStatusActive},
		{SectionArchive, domain.OfferStatusInactive},
	}

	for _, c := range cases {
		got, err := statusForSection(c.section)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", c.section, err)
		}
		if got != c.want {
			t.Errorf("%s: got %s, want %s", c.section, got, c.want)
		}
	}
}

func TestStatusForSectionRejectsUnknown(t *testing.T) {
	// Нулевое значение Section - как раз то, что ловится строковым типом
	for _, section := range []Section{"", "sold", "Active"} {
		if _, err := statusForSection(section); err == nil {
			t.Errorf("section %q: expected error, got nil", section)
		}
	}
}

// page - страница выдачи из нескольких строк
func page(rows ...row) string {
	var b strings.Builder
	b.WriteString(`<html><body><table>`)
	for _, r := range rows {
		b.WriteString(`<tr>` + r.html() + `</tr>`)
	}
	b.WriteString(`</table></body></html>`)

	return b.String()
}

func newTestService(t *testing.T, handler http.HandlerFunc) *Service {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return NewService(
		NewClient(server.URL, slog.New(slog.DiscardHandler)),
		NewMapper(server.URL),
	)
}

func servePage(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		io.WriteString(w, body)
	}
}

func TestFetchOffersStampsStatusFromSection(t *testing.T) {
	cases := []struct {
		section Section
		want    domain.OfferStatus
	}{
		{SectionActive, domain.OfferStatusActive},
		{SectionArchive, domain.OfferStatusInactive},
	}

	for _, c := range cases {
		service := newTestService(t, servePage(page(defaultRow(), defaultRow())))

		result, err := service.FetchOffers(context.Background(), c.section, 0)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", c.section, err)
		}

		if result.RowsFound != 2 || len(result.Offers) != 2 {
			t.Fatalf("%s: found %d rows, mapped %d offers, want 2/2",
				c.section, result.RowsFound, len(result.Offers))
		}

		for _, offer := range result.Offers {
			if offer.Offer.Status != c.want {
				t.Errorf("%s: status = %s, want %s", c.section, offer.Offer.Status, c.want)
			}
		}
	}
}

func TestFetchOffersRejectsUnknownSectionBeforeRequest(t *testing.T) {
	var called bool
	service := newTestService(t, func(w http.ResponseWriter, _ *http.Request) {
		called = true
		io.WriteString(w, page(defaultRow()))
	})

	if _, err := service.FetchOffers(context.Background(), "", 0); err == nil {
		t.Fatal("expected error for unknown section, got nil")
	}

	if called {
		t.Error("provider was called despite an invalid section")
	}
}

func TestFetchOffersCollectsRowErrors(t *testing.T) {
	broken := defaultRow()
	broken.href = "/quadro/а123аа77"

	skipped := defaultRow()
	skipped.price = "Договорная"

	service := newTestService(t, servePage(page(defaultRow(), broken, skipped)))

	result, err := service.FetchOffers(context.Background(), SectionActive, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.RowsFound != 3 {
		t.Fatalf("RowsFound = %d, want 3", result.RowsFound)
	}
	if len(result.Offers) != 1 {
		t.Fatalf("mapped %d offers, want 1", len(result.Offers))
	}
	if len(result.RowErrors) != 2 {
		t.Fatalf("collected %d row errors, want 2", len(result.RowErrors))
	}
	if broken := result.BrokenOfferCount(); broken != 1 {
		t.Errorf("BrokenOfferCount = %d, want 1", broken)
	}
	if !errors.Is(result.RowErrors[1].Err, providers.ErrRowSkipped) {
		t.Errorf("second row error = %v, want ErrRowSkipped", result.RowErrors[1].Err)
	}
}

func TestFetchOffersPropagatesClientError(t *testing.T) {
	service := newTestService(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})

	_, err := service.FetchOffers(context.Background(), SectionActive, 0)
	if !errors.Is(err, providers.ErrRateLimitExceeded) {
		t.Fatalf("want ErrRateLimitExceeded, got %v", err)
	}
}
