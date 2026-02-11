package processor

import (
	"time"

	"github.com/garbas/tgdash/internal/parser"
	"github.com/garbas/tgdash/internal/state"
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
	raw = parser.StripANSI(raw)

	// Detect "did not run" before normal parsing — these
	// are global lines that reference a specific unit.
	if skippedUnit, ok := parser.DetectDidNotRun(raw); ok {
		u := p.state.GetOrCreateUnit(skippedUnit)
		u.Status = state.StatusSkipped
		u.AppendLine(raw)
		return
	}

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

	// Detect "No changes" (before plan summary, as both
	// can appear on separate lines)
	if parser.DetectNoChanges(line) {
		u.PlanSummary = &state.PlanSummary{}
		u.Status = state.StatusDone
		if !u.StartTime.IsZero() {
			u.Duration = time.Since(u.StartTime)
		}
	}

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
