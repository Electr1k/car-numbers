package autonomera

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestClientManual(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	client := NewClient(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	html, err := client.FetchNumbersHTML(ctx, 0)
	if err != nil {
		t.Fatalf("FetchNumbersHTML failed: %v", err)
	}

	if len(html) == 0 {
		t.Fatal("Expected non-empty HTML response")
	}

	t.Logf("Successfully fetched %d bytes of HTML", len(html))
}
