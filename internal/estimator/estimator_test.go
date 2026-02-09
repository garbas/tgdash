package estimator

import (
	"testing"
	"time"

	"github.com/rok/tgdash/internal/state"
)

func TestEstimateNoHistory(t *testing.T) {
	e := New(&state.History{})
	est, ok := e.Estimate("vpc", "plan")
	if ok {
		t.Fatalf("expected no estimate, got %v", est)
	}
}

func TestEstimateSingleRun(t *testing.T) {
	h := &state.History{
		Runs: []state.RunRecord{
			{
				Command: "plan",
				Units: []state.UnitRecord{
					{
						Path:     "vpc",
						Status:   state.StatusDone,
						Duration: 10 * time.Second,
					},
				},
			},
		},
	}

	e := New(h)
	est, ok := e.Estimate("vpc", "plan")
	if !ok {
		t.Fatal("expected estimate")
	}
	if est != 10*time.Second {
		t.Fatalf("expected 10s, got %v", est)
	}
}

func TestEstimateMedian(t *testing.T) {
	h := &state.History{
		Runs: []state.RunRecord{
			{Command: "plan", Units: []state.UnitRecord{
				{Path: "vpc", Status: state.StatusDone,
					Duration: 5 * time.Second},
			}},
			{Command: "plan", Units: []state.UnitRecord{
				{Path: "vpc", Status: state.StatusDone,
					Duration: 15 * time.Second},
			}},
			{Command: "plan", Units: []state.UnitRecord{
				{Path: "vpc", Status: state.StatusDone,
					Duration: 10 * time.Second},
			}},
		},
	}

	e := New(h)
	est, ok := e.Estimate("vpc", "plan")
	if !ok {
		t.Fatal("expected estimate")
	}
	if est != 10*time.Second {
		t.Fatalf("expected 10s (median), got %v", est)
	}
}

func TestEstimateIgnoresDifferentCommand(t *testing.T) {
	h := &state.History{
		Runs: []state.RunRecord{
			{Command: "apply", Units: []state.UnitRecord{
				{Path: "vpc", Status: state.StatusDone,
					Duration: 30 * time.Second},
			}},
		},
	}

	e := New(h)
	_, ok := e.Estimate("vpc", "plan")
	if ok {
		t.Fatal("expected no estimate for different command")
	}
}

func TestEstimateIgnoresErrors(t *testing.T) {
	h := &state.History{
		Runs: []state.RunRecord{
			{Command: "plan", Units: []state.UnitRecord{
				{Path: "vpc", Status: state.StatusError,
					Duration: 2 * time.Second},
			}},
		},
	}

	e := New(h)
	_, ok := e.Estimate("vpc", "plan")
	if ok {
		t.Fatal("expected no estimate for errored runs")
	}
}
