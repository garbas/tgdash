package main

import (
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/garbas/tgdash/internal/estimator"
	"github.com/garbas/tgdash/internal/reader"
	"github.com/garbas/tgdash/internal/runner"
	"github.com/garbas/tgdash/internal/state"
	"github.com/garbas/tgdash/internal/tui"
	"golang.org/x/term"
)

func main() {
	var (
		input   io.Reader
		cleanup func()
		command string
	)

	args := os.Args[1:]

	switch {
	case len(args) > 0:
		// Exec mode: tgdash -- terragrunt run --all plan
		var err error
		input, cleanup, err = runner.Run(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer cleanup()
		command = runner.DetectCommand(args)

	case !term.IsTerminal(int(os.Stdin.Fd())):
		// Pipe mode: terragrunt run --all 2>&1 | tgdash
		input = os.Stdin
		command = "plan"

	default:
		fmt.Fprintln(os.Stderr,
			"Usage:\n"+
				"  tgdash -- <command>\n"+
				"  <command> 2>&1 | tgdash\n"+
				"\n"+
				"Examples:\n"+
				"  tgdash -- terragrunt run --all plan\n"+
				"  terragrunt run --all plan 2>&1 | tgdash")
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

	go reader.ReadLines(input,
		func(msg tea.Msg) { p.Send(msg) })

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Save history after TUI exits
	run := state.RunRecord{
		Command: command,
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
