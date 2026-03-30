//go:build manual
// +build manual

package autonomera

import (
	"context"
	"core/pkg/integration/autonomera"
	"log/slog"
	"os"
	"testing"
	"time"
)

// This is a manual test to verify the Provider Service works independently
// Run with: go test -tags=manual -v ./internal/provider/autonomera/
func TestProviderManual(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create HTTP client
	clientCfg := autonomera.Config{
		BaseURL: "https://autonomera777.net/ajax/",
		Timeout: 30 * time.Second,
	}
	client := autonomera.NewClient(clientCfg, logger)

	// Create provider
	provider := NewProvider(client, nil, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	numbers, err := provider.FetchNumbers(ctx, 0)
	if err != nil {
		t.Fatalf("FetchNumbers failed: %v", err)
	}

	if len(numbers) == 0 {
		t.Fatal("Expected at least one car number")
	}

	t.Logf("Successfully fetched %d car numbers", len(numbers))

	// Verify first number is valid
	if err := numbers[0].Validate(); err != nil {
		t.Fatalf("First car number is invalid: %v", err)
	}

	t.Logf("First car number: %s, Price: %.2f, Posted: %s",
		numbers[0].Number,
		numbers[0].Price,
		numbers[0].PostedAt.Format("02.01.2006"))
}
