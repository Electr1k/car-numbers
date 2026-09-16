package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewPlateAssignsID(t *testing.T) {
	first, err := NewPlate("а123аа77", PlateTypeCar)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if first.ID == uuid.Nil {
		t.Fatal("expected generated id, got uuid.Nil")
	}

	if version := first.ID.Version(); version != 7 {
		t.Fatalf("expected UUID version 7, got %d", version)
	}

	second, err := NewPlate("а123аа77", PlateTypeCar)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if first.ID == second.ID {
		t.Fatal("expected distinct ids for distinct plates")
	}
}

func TestNewPlateRejectsUnknownType(t *testing.T) {
	if _, err := NewPlate("а123аа77", "bicycle"); err == nil {
		t.Fatal("expected error for unknown plate type, got nil")
	}

	if _, err := NewPlate("а123аа77", ""); err == nil {
		t.Fatal("expected error for empty plate type, got nil")
	}
}

func TestNewPlateAcceptsKnownTypes(t *testing.T) {
	for _, plateType := range []PlateType{PlateTypeCar, PlateTypeMoto, PlateTypeTrailer} {
		if _, err := NewPlate("а123аа77", plateType); err != nil {
			t.Fatalf("%s: unexpected error: %v", plateType, err)
		}
	}
}

func TestRestorePlateRequiresID(t *testing.T) {
	if _, err := RestorePlate(uuid.Nil, "а123аа77", PlateTypeCar, nil, nil); err == nil {
		t.Fatal("expected error for zero id, got nil")
	}

	id := uuid.New()

	restored, err := RestorePlate(id, "а123аа77", PlateTypeCar, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if restored.ID != id {
		t.Fatalf("got id %s, want %s", restored.ID, id)
	}
}
