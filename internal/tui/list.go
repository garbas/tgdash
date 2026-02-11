package tui

import (
	"fmt"
	"strings"

	"github.com/garbas/tgdash/internal/state"
)

type ListView struct {
	expanded map[int]bool
	width    int
	height   int
}

func NewListView() ListView {
	return ListView{
		expanded: make(map[int]bool),
	}
}

func (l *ListView) SetSize(w, h int) {
	l.width = w
	l.height = h
}

func (l *ListView) ToggleExpanded(idx int) {
	l.expanded[idx] = !l.expanded[idx]
}

func (l *ListView) IsExpanded(idx int) bool {
	return l.expanded[idx]
}

func (l *ListView) Render(
	appState *state.AppState,
	estimates map[string]string,
) string {
	units := appState.Units()
	var lines []string

	for i, u := range units {
		arrow := "▶"
		if l.expanded[i] {
			arrow = "▼"
		}

		selected := i == appState.SelectedIdx
		var status, summary, timeStr string
		if selected {
			status = formatStatusPlain(u.Status)
			if u.PlanSummary != nil {
				summary = fmt.Sprintf(
					" +%d ~%d -%d",
					u.PlanSummary.Add,
					u.PlanSummary.Change,
					u.PlanSummary.Destroy)
			}
			timeStr = formatTimePlain(u, estimates)
		} else {
			status = formatStatus(u.Status)
			if u.PlanSummary != nil {
				summary = fmt.Sprintf(" %s%s%s",
					AddStyle.Render(fmt.Sprintf(
						"+%d", u.PlanSummary.Add)),
					ChangeStyle.Render(fmt.Sprintf(
						" ~%d", u.PlanSummary.Change)),
					DestroyStyle.Render(fmt.Sprintf(
						" -%d", u.PlanSummary.Destroy)))
			}
			timeStr = formatTime(u, estimates)
		}

		header := fmt.Sprintf("%s %s [%s]%s  %s",
			arrow, u.Path, status, summary, timeStr)

		if selected {
			header = SelectedStyle.
				Width(l.width - 2).Render(header)
		}

		lines = append(lines, header)

		if l.expanded[i] {
			maxLines := 20
			start := 0
			if len(u.OutputLines) > maxLines {
				start = len(u.OutputLines) - maxLines
			}
			for _, ol := range u.OutputLines[start:] {
				lines = append(lines,
					HelpStyle.Render(
						"  │ "+truncate(ol, l.width-6)))
			}
		}
	}

	content := strings.Join(lines, "\n")

	if len(lines) > l.height {
		visible := strings.Join(
			lines[:min(len(lines), l.height)], "\n")
		return visible
	}

	return content
}
