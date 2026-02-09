package processor

import (
	"testing"

	"github.com/rok/tgdash/internal/state"
)

func TestProcessLineCreatesUnit(t *testing.T) {
	s := state.NewAppState()
	p := New(s)
	p.ProcessLine("[live/dev/vpc] Initializing...")

	units := s.Units()
	if len(units) != 1 {
		t.Fatalf("expected 1 unit, got %d", len(units))
	}
	if units[0].Path != "live/dev/vpc" {
		t.Fatalf("expected path live/dev/vpc, got %s",
			units[0].Path)
	}
	if len(units[0].OutputLines) != 1 {
		t.Fatalf("expected 1 output line, got %d",
			len(units[0].OutputLines))
	}
}

func TestProcessLinePlanSummary(t *testing.T) {
	s := state.NewAppState()
	p := New(s)
	p.ProcessLine(
		"[vpc] Plan: 3 to add, 1 to change, 0 to destroy.")

	u := s.Units()[0]
	if u.PlanSummary == nil {
		t.Fatal("expected plan summary to be set")
	}
	if u.PlanSummary.Add != 3 {
		t.Fatalf("expected add=3, got %d",
			u.PlanSummary.Add)
	}
	if u.PlanSummary.Change != 1 {
		t.Fatalf("expected change=1, got %d",
			u.PlanSummary.Change)
	}
}

func TestProcessLineApplyResult(t *testing.T) {
	s := state.NewAppState()
	p := New(s)
	p.ProcessLine(
		"[vpc] Apply complete! Resources: 2 added, 1 changed, 0 destroyed.")

	u := s.Units()[0]
	if u.PlanSummary == nil {
		t.Fatal("expected plan summary from apply result")
	}
	if u.PlanSummary.Add != 2 {
		t.Fatalf("expected add=2, got %d",
			u.PlanSummary.Add)
	}
	if u.Status != state.StatusDone {
		t.Fatalf("expected status done, got %s",
			u.Status)
	}
}

func TestProcessLineError(t *testing.T) {
	s := state.NewAppState()
	p := New(s)
	p.ProcessLine("[vpc] Error: Invalid resource type")

	u := s.Units()[0]
	if u.Status != state.StatusError {
		t.Fatalf("expected status error, got %s",
			u.Status)
	}
	if len(u.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d",
			len(u.Errors))
	}
}

func TestProcessLineNoPrefixAttachesToLast(t *testing.T) {
	s := state.NewAppState()
	p := New(s)
	p.ProcessLine("[vpc] First line")
	p.ProcessLine("  continuation line")

	units := s.Units()
	if len(units) != 1 {
		t.Fatalf("expected 1 unit, got %d", len(units))
	}
	if len(units[0].OutputLines) != 2 {
		t.Fatalf("expected 2 lines, got %d",
			len(units[0].OutputLines))
	}
}

func TestProcessLineNoPrefixNoUnit(t *testing.T) {
	s := state.NewAppState()
	p := New(s)
	p.ProcessLine("global message before any unit")

	// Should create a global unit
	units := s.Units()
	if len(units) != 1 {
		t.Fatalf("expected 1 unit, got %d", len(units))
	}
	if units[0].Path != GlobalUnitPath {
		t.Fatalf("expected global unit, got %s",
			units[0].Path)
	}
}
