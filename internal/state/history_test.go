package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndLoadHistory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")

	run := RunRecord{
		Command:   "plan",
		Timestamp: time.Now(),
		Units: []UnitRecord{
			{
				Path:     "live/dev/vpc",
				Status:   StatusDone,
				Duration: 12 * time.Second,
				PlanSummary: &PlanSummary{
					Add: 3, Change: 1, Destroy: 0,
				},
			},
		},
	}

	h := &History{Runs: []RunRecord{run}}
	err := SaveHistory(path, h)
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := LoadHistory(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(loaded.Runs) != 1 {
		t.Fatalf("expected 1 run, got %d",
			len(loaded.Runs))
	}
	if loaded.Runs[0].Units[0].Path != "live/dev/vpc" {
		t.Fatalf("expected vpc, got %s",
			loaded.Runs[0].Units[0].Path)
	}
}

func TestLoadHistoryMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nope.json")
	h, err := LoadHistory(path)
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v",
			err)
	}
	if len(h.Runs) != 0 {
		t.Fatalf("expected 0 runs, got %d",
			len(h.Runs))
	}
}

func TestLoadHistoryCorrupted(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")
	os.WriteFile(path, []byte("not json{{{"), 0644)

	h, err := LoadHistory(path)
	if err != nil {
		t.Fatalf("expected no error for corrupted file, got %v",
			err)
	}
	if len(h.Runs) != 0 {
		t.Fatalf("expected 0 runs for corrupted file, got %d",
			len(h.Runs))
	}
}

func TestHistoryMaxRuns(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")

	h := &History{}
	for i := 0; i < MaxHistoryRuns+50; i++ {
		h.Runs = append(h.Runs, RunRecord{
			Command:   "plan",
			Timestamp: time.Now(),
		})
	}
	err := SaveHistory(path, h)
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := LoadHistory(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(loaded.Runs) != MaxHistoryRuns {
		t.Fatalf("expected %d runs, got %d",
			MaxHistoryRuns, len(loaded.Runs))
	}
}
