package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rok/tgdash/internal/estimator"
	"github.com/rok/tgdash/internal/reader"
	"github.com/rok/tgdash/internal/state"
	"github.com/rok/tgdash/internal/tui"
	"golang.org/x/term"
)

func main() {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Fprintln(os.Stderr,
			"Usage: terragrunt run --all plan 2>&1 | tgdash")
		os.Exit(1)
	}

	historyPath := state.DefaultHistoryPath()
	history, _ := state.LoadHistory(historyPath)
	est := estimator.New(history)

	appState := state.NewAppState()
	model := tui.NewModel(appState, est)

	p := tea.NewProgram(model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	go reader.ReadLines(os.Stdin,
		func(msg tea.Msg) { p.Send(msg) })

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Save history after TUI exits
	run := state.RunRecord{
		Command: "plan",
	}
	for _, u := range appState.Units() {
		run.Units = append(run.Units, state.UnitRecord{
			Path:        u.Path,
			Status:      u.Status,
			Duration:    u.Duration,
			PlanSummary: u.PlanSummary,
		})
	}
	history.Runs = append(history.Runs, run)
	_ = state.SaveHistory(historyPath, history)
}
