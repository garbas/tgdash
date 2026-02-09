package estimator

import (
	"sort"
	"time"

	"github.com/garbas/tgdash/internal/state"
)

type Estimator struct {
	history *state.History
}

func New(h *state.History) *Estimator {
	return &Estimator{history: h}
}

func (e *Estimator) Estimate(
	unitPath string, command string,
) (time.Duration, bool) {
	var durations []time.Duration

	for _, run := range e.history.Runs {
		if run.Command != command {
			continue
		}
		for _, u := range run.Units {
			if u.Path == unitPath &&
				u.Status == state.StatusDone {
				durations = append(durations, u.Duration)
			}
		}
	}

	if len(durations) == 0 {
		return 0, false
	}

	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})

	median := durations[len(durations)/2]
	return median, true
}
