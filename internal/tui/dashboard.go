package tui

import (
	"fmt"
	"regexp"
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

	// Column widths
	const (
		colStatus = 8
		colUnit   = 30
		colAdd    = 4
		colChg    = 4
		colDel    = 4
		colTime   = 10
	)
	sep := " │ "

	// Render table
	var tableRows []string
	header := padRight("STATUS", colStatus) + sep +
		padRight("UNIT", colUnit) + sep +
		padRight("+", colAdd) + sep +
		padRight("~", colChg) + sep +
		padRight("-", colDel) + sep +
		padRight("TIME", colTime)
	tableRows = append(tableRows,
		TitleStyle.Width(d.width-2).Render(header))

	for i, u := range units {
		selected := i == appState.SelectedIdx
		var status string
		add, chg, del := "/", "/", "/"
		var timeStr string
		if selected {
			status = formatStatusPlain(u.Status)
			if u.PlanSummary != nil {
				add = fmt.Sprintf(
					"%d", u.PlanSummary.Add)
				chg = fmt.Sprintf(
					"%d", u.PlanSummary.Change)
				del = fmt.Sprintf(
					"%d", u.PlanSummary.Destroy)
			}
			timeStr = formatTimePlain(u, estimates)
		} else {
			status = formatStatus(u.Status)
			if u.PlanSummary != nil {
				add = AddStyle.Render(fmt.Sprintf(
					"%d", u.PlanSummary.Add))
				chg = ChangeStyle.Render(fmt.Sprintf(
					"%d", u.PlanSummary.Change))
				del = DestroyStyle.Render(fmt.Sprintf(
					"%d", u.PlanSummary.Destroy))
			}
			timeStr = formatTime(u, estimates)
		}

		line := " " +
			padRight(status, colStatus) + sep +
			padRight(truncate(u.Path, colUnit),
				colUnit) + sep +
			padRight(add, colAdd) + sep +
			padRight(chg, colChg) + sep +
			padRight(del, colDel) + sep +
			padRight(timeStr, colTime)

		if selected {
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
	divider := HelpStyle.Render(
		fmt.Sprintf("─── %s ", selectedName) +
			strings.Repeat("─",
				max(0, d.width-len(selectedName)-6)))

	parts := []string{table, divider, d.viewport.View()}
	if appState.InputDone {
		parts = append(parts, HelpStyle.Render(
			"Run complete. ↑/↓ units, j/k scroll, q quit."))
	}
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
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

func formatStatusPlain(s state.UnitStatus) string {
	switch s {
	case state.StatusWaiting:
		return "○ wait"
	case state.StatusRunning:
		return "● run"
	case state.StatusDone:
		return "✓ done"
	case state.StatusError:
		return "✗ err"
	case state.StatusSkipped:
		return "⊘ skip"
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
	return "/"
}

func formatTimePlain(
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
			return elapsed + " / " + est
		}
		return elapsed
	}
	return "/"
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// padRight pads s to width based on visible length
// (ignoring ANSI escape sequences).
func padRight(s string, width int) string {
	visible := visibleLen(s)
	if visible >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visible)
}

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func visibleLen(s string) int {
	return len(ansiRe.ReplaceAllString(s, ""))
}
