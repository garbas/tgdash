package state

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

const MaxHistoryRuns = 100

type UnitRecord struct {
	Path        string        `json:"path"`
	Status      UnitStatus    `json:"status"`
	Duration    time.Duration `json:"duration"`
	PlanSummary *PlanSummary  `json:"plan_summary,omitempty"`
}

type RunRecord struct {
	Command   string       `json:"command"`
	Timestamp time.Time    `json:"timestamp"`
	Units     []UnitRecord `json:"units"`
}

type History struct {
	Runs []RunRecord `json:"runs"`
}

func DefaultHistoryPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".tgdash", "history.json")
}

func LoadHistory(path string) (*History, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &History{}, nil
		}
		return &History{}, nil
	}
	var h History
	if err := json.Unmarshal(data, &h); err != nil {
		return &History{}, nil
	}
	return &h, nil
}

func SaveHistory(path string, h *History) error {
	if len(h.Runs) > MaxHistoryRuns {
		h.Runs = h.Runs[len(h.Runs)-MaxHistoryRuns:]
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
