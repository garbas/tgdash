package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/garbas/tgdash/internal/estimator"
	"github.com/garbas/tgdash/internal/processor"
	"github.com/garbas/tgdash/internal/reader"
	"github.com/garbas/tgdash/internal/state"
)

type Model struct {
	state     *state.AppState
	proc      *processor.Processor
	estimator *estimator.Estimator
	dashboard DashboardView
	list      ListView
	keys      KeyMap
	width     int
	height    int
	gPending  bool
	estimates map[string]string
}

func NewModel(
	appState *state.AppState,
	est *estimator.Estimator,
) Model {
	return Model{
		state:     appState,
		proc:      processor.New(appState),
		estimator: est,
		dashboard: NewDashboardView(),
		list:      NewListView(),
		keys:      DefaultKeyMap(),
		estimates: make(map[string]string),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(
	msg tea.Msg,
) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case reader.LineMsg:
		m.proc.ProcessLine(msg.Raw)
		m.updateEstimates()
		m.updateViewport()
		return m, nil

	case reader.InputDoneMsg:
		m.state.InputDone = true
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.dashboard.SetSize(msg.Width, msg.Height)
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	// Pass to viewport if in dashboard view.
	if m.state.ActiveView == state.ViewDashboard {
		var cmd tea.Cmd
		m.dashboard.viewport, cmd =
			m.dashboard.viewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleKey(
	msg tea.KeyMsg,
) (tea.Model, tea.Cmd) {
	units := m.state.Units()

	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Tab):
		if m.state.ActiveView == state.ViewDashboard {
			m.state.ActiveView = state.ViewList
		} else {
			m.state.ActiveView = state.ViewDashboard
		}

	case key.Matches(msg, m.keys.Down):
		if m.state.SelectedIdx < len(units)-1 {
			m.state.SelectedIdx++
			m.updateViewport()
		}

	case key.Matches(msg, m.keys.Up):
		if m.state.SelectedIdx > 0 {
			m.state.SelectedIdx--
			m.updateViewport()
		}

	case key.Matches(msg, m.keys.ScrollDown):
		m.dashboard.viewport.LineDown(1)

	case key.Matches(msg, m.keys.ScrollUp):
		m.dashboard.viewport.LineUp(1)

	case key.Matches(msg, m.keys.Top):
		if m.gPending {
			m.state.SelectedIdx = 0
			m.updateViewport()
			m.gPending = false
		} else {
			m.gPending = true
			return m, nil
		}

	case key.Matches(msg, m.keys.Bottom):
		if len(units) > 0 {
			m.state.SelectedIdx = len(units) - 1
			m.updateViewport()
		}

	case key.Matches(msg, m.keys.Enter):
		if m.state.ActiveView == state.ViewList {
			m.list.ToggleExpanded(m.state.SelectedIdx)
		}

	case key.Matches(msg, m.keys.Errors):
		// Toggle error filter (simplified).

	default:
		m.gPending = false
	}

	// Reset g pending on any non-g key.
	if !key.Matches(msg, m.keys.Top) {
		m.gPending = false
	}

	return m, nil
}

func (m *Model) updateViewport() {
	units := m.state.Units()
	if m.state.SelectedIdx < len(units) {
		m.dashboard.UpdateViewport(
			units[m.state.SelectedIdx])
	}
}

func (m *Model) updateEstimates() {
	if m.estimator == nil {
		return
	}
	for _, u := range m.state.Units() {
		if u.Status == state.StatusRunning {
			if est, ok := m.estimator.Estimate(
				u.Path, "plan"); ok {
				m.estimates[u.Path] =
					"~" + est.Truncate(
						time.Second).String() + " est."
			}
		}
	}
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	switch m.state.ActiveView {
	case state.ViewDashboard:
		return m.dashboard.Render(m.state, m.estimates)
	case state.ViewList:
		return m.list.Render(m.state, m.estimates)
	}
	return ""
}

// Viewport returns the dashboard viewport for
// external update (e.g., scroll via PageUp/PageDown).
func (m *Model) Viewport() *viewport.Model {
	return &m.dashboard.viewport
}
