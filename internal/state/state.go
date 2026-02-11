package state

import "time"

type UnitStatus string

const (
	StatusWaiting UnitStatus = "waiting"
	StatusRunning UnitStatus = "running"
	StatusDone    UnitStatus = "done"
	StatusError   UnitStatus = "error"
	StatusSkipped UnitStatus = "skipped"
)

const MaxOutputLines = 10000

type PlanSummary struct {
	Add     int
	Change  int
	Destroy int
}

type Unit struct {
	Path        string
	Status      UnitStatus
	OutputLines []string
	PlanSummary *PlanSummary
	Errors      []string
	StartTime   time.Time
	Duration    time.Duration
}

func (u *Unit) AppendLine(line string) {
	u.OutputLines = append(u.OutputLines, line)
	if len(u.OutputLines) > MaxOutputLines {
		u.OutputLines = u.OutputLines[1:]
	}
}

type ViewMode int

const (
	ViewDashboard ViewMode = iota
	ViewList
)

type AppState struct {
	units       []*Unit
	unitIndex   map[string]*Unit
	ActiveView  ViewMode
	SelectedIdx int
	Filter      string
	InputDone   bool
}

func NewAppState() *AppState {
	return &AppState{
		unitIndex: make(map[string]*Unit),
	}
}

func (s *AppState) GetOrCreateUnit(path string) *Unit {
	if u, ok := s.unitIndex[path]; ok {
		return u
	}
	u := &Unit{
		Path:   path,
		Status: StatusWaiting,
	}
	s.units = append(s.units, u)
	s.unitIndex[path] = u
	return u
}

func (s *AppState) Units() []*Unit {
	return s.units
}
