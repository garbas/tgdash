package processor

import (
	"time"

	"github.com/rok/tgdash/internal/parser"
	"github.com/rok/tgdash/internal/state"
)

const GlobalUnitPath = "(global)"

type Processor struct {
	state    *state.AppState
	lastUnit string
}

func New(s *state.AppState) *Processor {
	return &Processor{state: s}
}

func (p *Processor) ProcessLine(raw string) {
	unitPath, line := parser.ExtractUnitPrefix(raw)

	if unitPath == "" {
		unitPath = p.lastUnit
		line = raw
	}
	if unitPath == "" {
		unitPath = GlobalUnitPath
	}

	p.lastUnit = unitPath

	u := p.state.GetOrCreateUnit(unitPath)
	u.AppendLine(line)

	// Detect plan summary
	if summary, ok := parser.DetectPlanSummary(line); ok {
		u.PlanSummary = &state.PlanSummary{
			Add:     summary.Add,
			Change:  summary.Change,
			Destroy: summary.Destroy,
		}
	}

	// Detect apply result
	if summary, ok := parser.DetectApplyResult(line); ok {
		u.PlanSummary = &state.PlanSummary{
			Add:     summary.Add,
			Change:  summary.Change,
			Destroy: summary.Destroy,
		}
		u.Status = state.StatusDone
		if !u.StartTime.IsZero() {
			u.Duration = time.Since(u.StartTime)
		}
	}

	// Detect errors
	if parser.DetectError(line) {
		u.Status = state.StatusError
		u.Errors = append(u.Errors, line)
	}

	// If unit was waiting and we got output, it's running
	if u.Status == state.StatusWaiting {
		u.Status = state.StatusRunning
		u.StartTime = time.Now()
	}
}
