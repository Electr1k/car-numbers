package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewNumberAssignsID(t *testing.T) {
	first, err := NewNumber("а123аа77", NumberTypeCar)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if first.Id == uuid.Nil {
		t.Fatal("expected generated id, got uuid.Nil")
	}

	if version := first.Id.Version(); version != 7 {
		t.Fatalf("expected UUID version 7, got %d", version)
	}

	second, err := NewNumber("а123аа77", NumberTypeCar)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if first.Id == second.Id {
		t.Fatal("expected distinct ids for distinct numbers")
	}
}

func TestNewNumberRejectsUnknownType(t *testing.T) {
	if _, err := NewNumber("а123аа77", "bicycle"); err == nil {
		t.Fatal("expected error for unknown number type, got nil")
	}

	if _, err := NewNumber("а123аа77", ""); err == nil {
		t.Fatal("expected error for empty number type, got nil")
	}
}

func TestNewNumberAcceptsKnownTypes(t *testing.T) {
	for _, numberType := range []NumberType{NumberTypeCar, NumberTypeMoto, NumberTypeTrailer} {
		if _, err := NewNumber("а123аа77", numberType); err != nil {
			t.Fatalf("%s: unexpected error: %v", numberType, err)
		}
	}
}

func TestRestoreNumberRequiresID(t *testing.T) {
	if _, err := RestoreNumber(uuid.Nil, "а123аа77", NumberTypeCar, nil, nil); err == nil {
		t.Fatal("expected error for zero id, got nil")
	}

	id := uuid.New()

	restored, err := RestoreNumber(id, "а123аа77", NumberTypeCar, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if restored.Id != id {
		t.Fatalf("got id %s, want %s", restored.Id, id)
	}
}
