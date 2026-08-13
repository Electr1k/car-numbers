package autonomera

import (
	"context"
	"data-service/internal/provider"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return NewClient(server.URL, slog.New(slog.DiscardHandler))
}

func TestStatusCodeMapsToSentinel(t *testing.T) {
	cases := []struct {
		status int
		want   error
	}{
		{http.StatusTooManyRequests, provider.ErrRateLimitExceeded},
		{http.StatusInternalServerError, provider.ErrProviderUnavailable},
		{http.StatusBadGateway, provider.ErrProviderUnavailable},
		{http.StatusNotFound, provider.ErrInvalidResponse},
		{http.StatusForbidden, provider.ErrInvalidResponse},
	}

	for _, c := range cases {
		client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(c.status)
			io.WriteString(w, "<html>  nope\n\n</html>")
		})

		_, err := client.FetchOffersHTML(context.Background(), SectionActive, 0)
		if err == nil {
			t.Fatalf("status %d: expected error, got nil", c.status)
		}
		if !errors.Is(err, c.want) {
			t.Fatalf("status %d: errors.Is failed, got %q", c.status, err)
		}
		if strings.Contains(err.Error(), "%!w") {
			t.Fatalf("status %d: broken wrap verb: %s", c.status, err)
		}
		// превью схлопнуто в одну строку
		if !strings.Contains(err.Error(), "<html> nope </html>") {
			t.Fatalf("status %d: preview not collapsed: %s", c.status, err)
		}
	}
}

func TestEmptyErrorBodyPreview(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	_, err := client.FetchOffersHTML(context.Background(), SectionActive, 0)
	if err == nil || !strings.Contains(err.Error(), "<empty body>") {
		t.Fatalf("expected <empty body> in error, got %v", err)
	}
}

func TestQueryParams(t *testing.T) {
	cases := []struct {
		section  Section
		wantBlog string
		hasBlog  bool
	}{
		{SectionActive, "", false},
		{SectionArchive, "numbersarchive", true},
	}

	for _, c := range cases {
		var got *http.Request
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			got = r
			io.WriteString(w, "<html></html>")
		})

		if _, err := client.FetchOffersHTML(context.Background(), c.section, 40); err != nil {
			t.Fatalf("%s: unexpected error: %v", c.section, err)
		}

		query := got.URL.Query()
		if query.Get("start") != "40" {
			t.Fatalf("%s: start = %q, want 40", c.section, query.Get("start"))
		}
		if query.Get("order") != "a.`created`" || query.Get("dir") != "DESC" {
			t.Fatalf("%s: order/dir = %q/%q", c.section, query.Get("order"), query.Get("dir"))
		}
		if _, present := query["blog"]; present != c.hasBlog {
			t.Fatalf("%s: blog present = %v, want %v", c.section, present, c.hasBlog)
		}
		if query.Get("blog") != c.wantBlog {
			t.Fatalf("%s: blog = %q, want %q", c.section, query.Get("blog"), c.wantBlog)
		}
	}
}

func TestSuccessReturnsBody(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		io.WriteString(w, "<html>ok</html>")
	})

	body, err := client.FetchOffersHTML(context.Background(), SectionActive, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != "<html>ok</html>" {
		t.Fatalf("got %q", body)
	}
}
