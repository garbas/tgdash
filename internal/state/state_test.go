package state

import (
	"testing"
)

func TestNewAppState(t *testing.T) {
	s := NewAppState()
	if s == nil {
		t.Fatal("expected non-nil state")
	}
	if len(s.Units()) != 0 {
		t.Fatalf("expected 0 units, got %d",
			len(s.Units()))
	}
	if s.InputDone {
		t.Fatal("expected InputDone to be false")
	}
}

func TestGetOrCreateUnit(t *testing.T) {
	s := NewAppState()
	u := s.GetOrCreateUnit("live/dev/vpc")
	if u.Path != "live/dev/vpc" {
		t.Fatalf("expected path live/dev/vpc, got %s",
			u.Path)
	}
	if u.Status != StatusWaiting {
		t.Fatalf("expected status waiting, got %s",
			u.Status)
	}

	// Same path returns same unit
	u2 := s.GetOrCreateUnit("live/dev/vpc")
	if u != u2 {
		t.Fatal("expected same unit pointer")
	}

	// Different path creates new unit
	u3 := s.GetOrCreateUnit("live/dev/rds")
	if u3 == u {
		t.Fatal("expected different unit")
	}
}

func TestUnitsOrdering(t *testing.T) {
	s := NewAppState()
	s.GetOrCreateUnit("b")
	s.GetOrCreateUnit("a")
	s.GetOrCreateUnit("c")
	units := s.Units()
	if len(units) != 3 {
		t.Fatalf("expected 3 units, got %d", len(units))
	}
	// Insertion order preserved
	if units[0].Path != "b" {
		t.Fatalf("expected b first, got %s",
			units[0].Path)
	}
	if units[1].Path != "a" {
		t.Fatalf("expected a second, got %s",
			units[1].Path)
	}
	if units[2].Path != "c" {
		t.Fatalf("expected c third, got %s",
			units[2].Path)
	}
}

func TestAppendLine(t *testing.T) {
	s := NewAppState()
	u := s.GetOrCreateUnit("vpc")
	u.AppendLine("line 1")
	u.AppendLine("line 2")
	if len(u.OutputLines) != 2 {
		t.Fatalf("expected 2 lines, got %d",
			len(u.OutputLines))
	}
}

func TestAppendLineMaxBuffer(t *testing.T) {
	s := NewAppState()
	u := s.GetOrCreateUnit("vpc")
	for i := 0; i < MaxOutputLines+100; i++ {
		u.AppendLine("line")
	}
	if len(u.OutputLines) != MaxOutputLines {
		t.Fatalf("expected %d lines, got %d",
			MaxOutputLines, len(u.OutputLines))
	}
}
