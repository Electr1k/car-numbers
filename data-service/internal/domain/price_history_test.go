package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewPriceHistoryAssignsID(t *testing.T) {
	price := 1000.0

	history, err := NewPriceHistory(uuid.New(), uuid.New(), &price)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if history.Id == uuid.Nil {
		t.Fatal("expected generated id, got uuid.Nil")
	}

	if version := history.Id.Version(); version != 7 {
		t.Fatalf("expected UUID version 7, got %d", version)
	}
}

// Снятая цена - валидное наблюдение, оно и есть событие "цену убрали"
func TestNewPriceHistoryAcceptsMissingPrice(t *testing.T) {
	if _, err := NewPriceHistory(uuid.New(), uuid.New(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewPriceHistoryRejectsNonPositivePrice(t *testing.T) {
	for _, price := range []float64{0, -1} {
		if _, err := NewPriceHistory(uuid.New(), uuid.New(), &price); err == nil {
			t.Fatalf("%v: expected error for non-positive price, got nil", price)
		}
	}
}

func TestNewPriceHistoryRequiresIDs(t *testing.T) {
	price := 1000.0

	if _, err := NewPriceHistory(uuid.Nil, uuid.New(), &price); err == nil {
		t.Fatal("expected error for zero offer id, got nil")
	}

	if _, err := NewPriceHistory(uuid.New(), uuid.Nil, &price); err == nil {
		t.Fatal("expected error for zero number id, got nil")
	}
}

// Номер наблюдения - снимок номера предложения на этот момент
func TestNewPriceHistoryFromOfferCopiesOfferState(t *testing.T) {
	numberId := uuid.New()

	offer, err := newTestOffer(numberId, ProviderAutonomera, OfferStatusActive)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	history, err := NewPriceHistoryFromOffer(offer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if history.OfferId != offer.Id {
		t.Fatalf("offer id = %s, want %s", history.OfferId, offer.Id)
	}

	if history.NumberId != numberId {
		t.Fatalf("number id = %s, want %s", history.NumberId, numberId)
	}

	if history.Price != offer.Price {
		t.Fatalf("price = %v, want %v", history.Price, offer.Price)
	}
}

func TestRestorePriceHistoryRequiresID(t *testing.T) {
	price := 1000.0

	if _, err := RestorePriceHistory(uuid.Nil, uuid.New(), uuid.New(), &price, nil, nil); err == nil {
		t.Fatal("expected error for zero id, got nil")
	}
}
