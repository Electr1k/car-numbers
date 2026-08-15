package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func newTestOffer(numberId uuid.UUID, provider Provider, status OfferStatus) (*Offer, error) {
	postedAt := time.Now()
	price := 1000.0

	return NewOffer(
		numberId,
		provider,
		"42",
		&price,
		status,
		nil,
		nil,
		nil,
		&postedAt,
		&postedAt,
		"https://example.com/42",
		"<tr></tr>",
		nil,
	)
}

func TestNewOfferAssignsID(t *testing.T) {
	offer, err := newTestOffer(uuid.New(), ProviderAutonomera, OfferStatusActive)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if offer.Id == uuid.Nil {
		t.Fatal("expected generated id, got uuid.Nil")
	}

	if version := offer.Id.Version(); version != 7 {
		t.Fatalf("expected UUID version 7, got %d", version)
	}
}

func TestNewOfferRequiresNumberID(t *testing.T) {
	if _, err := newTestOffer(uuid.Nil, ProviderAutonomera, OfferStatusActive); err == nil {
		t.Fatal("expected error for zero number id, got nil")
	}
}

func TestNewOfferRejectsUnknownEnums(t *testing.T) {
	if _, err := newTestOffer(uuid.New(), ProviderAutonomera, "archived"); err == nil {
		t.Fatal("expected error for unknown offer status, got nil")
	}

	if _, err := newTestOffer(uuid.New(), "avito", OfferStatusActive); err == nil {
		t.Fatal("expected error for unknown provider, got nil")
	}
}

func TestNewOfferAcceptsKnownStatuses(t *testing.T) {
	for _, status := range []OfferStatus{OfferStatusActive, OfferStatusSold, OfferStatusInactive} {
		if _, err := newTestOffer(uuid.New(), ProviderAutonomera, status); err != nil {
			t.Fatalf("%s: unexpected error: %v", status, err)
		}
	}
}

func TestRestoreOfferRequiresID(t *testing.T) {
	postedAt := time.Now()
	price := 1000.0

	_, err := RestoreOffer(
		uuid.Nil,
		uuid.New(),
		ProviderAutonomera,
		"42",
		&price,
		OfferStatusActive,
		nil,
		nil,
		nil,
		&postedAt,
		&postedAt,
		"https://example.com/42",
		"<tr></tr>",
		nil,
		nil,
		nil,
	)
	if err == nil {
		t.Fatal("expected error for zero id, got nil")
	}
}
