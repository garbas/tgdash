package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/garbas/tgdash/internal/state"
)

type DashboardView struct {
	viewport    viewport.Model
	width       int
	height      int
	tableHeight int
}

func NewDashboardView() DashboardView {
	return DashboardView{
		tableHeight: 8,
	}
}

func (d *DashboardView) SetSize(w, h int) {
	d.width = w
	d.height = h
	vpHeight := h - d.tableHeight - 3
	if vpHeight < 1 {
		vpHeight = 1
	}
	d.viewport = viewport.New(w-4, vpHeight)
}

func (d *DashboardView) UpdateViewport(u *state.Unit) {
	if u == nil {
		d.viewport.SetContent("")
		return
	}
	d.viewport.SetContent(
		strings.Join(u.OutputLines, "\n"))
	d.viewport.GotoBottom()
}

func (d *DashboardView) Render(
	appState *state.AppState,
	estimates map[string]string,
) string {
	units := appState.Units()

	// Render table
	var tableRows []string
	header := fmt.Sprintf(
		"  %-8s %-30s %4s %4s %4s  %s",
		"STATUS", "UNIT", "+", "~", "-", "TIME")
	tableRows = append(tableRows,
		TitleStyle.Width(d.width-2).Render(header))

	for i, u := range units {
		status := formatStatus(u.Status)
		add, chg, del := "-", "-", "-"
		if u.PlanSummary != nil {
			add = AddStyle.Render(
				fmt.Sprintf("%d", u.PlanSummary.Add))
			chg = ChangeStyle.Render(
				fmt.Sprintf("%d", u.PlanSummary.Change))
			del = DestroyStyle.Render(
				fmt.Sprintf("%d", u.PlanSummary.Destroy))
		}

		timeStr := formatTime(u, estimates)

		line := fmt.Sprintf(
			"  %-8s %-30s %4s %4s %4s  %s",
			status, truncate(u.Path, 30),
			add, chg, del, timeStr)

		if i == appState.SelectedIdx {
			line = SelectedStyle.Width(d.width - 2).
				Render(line)
		}

		tableRows = append(tableRows, line)
		if len(tableRows) >= d.tableHeight {
			break
		}
	}

	table := strings.Join(tableRows, "\n")

	// Separator
	selectedName := ""
	if appState.SelectedIdx < len(units) {
		selectedName = units[appState.SelectedIdx].Path
	}
	sep := HelpStyle.Render(
		fmt.Sprintf("─── %s ", selectedName) +
			strings.Repeat("─",
				max(0, d.width-len(selectedName)-6)))

	return lipgloss.JoinVertical(lipgloss.Left,
		table, sep, d.viewport.View())
}

func formatStatus(s state.UnitStatus) string {
	switch s {
	case state.StatusWaiting:
		return StatusWaiting.Render("○ wait")
	case state.StatusRunning:
		return StatusRunning.Render("● run")
	case state.StatusDone:
		return StatusDone.Render("✓ done")
	case state.StatusError:
		return StatusError.Render("✗ err")
	case state.StatusSkipped:
		return StatusSkipped.Render("⊘ skip")
	}
	return string(s)
}

func formatTime(
	u *state.Unit,
	estimates map[string]string,
) string {
	if u.Duration > 0 {
		return fmt.Sprintf("%ds",
			int(u.Duration.Seconds()))
	}
	if u.Status == state.StatusRunning {
		elapsed := fmt.Sprintf("%ds",
			int(u.Duration.Seconds()))
		if est, ok := estimates[u.Path]; ok {
			return elapsed + " / " +
				EstimateStyle.Render(est)
		}
		return elapsed
	}
	return "-"
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
