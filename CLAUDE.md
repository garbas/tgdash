# tgdash

Go CLI TUI for streaming Terragrunt output. Parses prefixed
lines from `terragrunt run --all` and renders a Bubble Tea
dashboard with per-unit status, plan summaries, and errors.

## Environment

Go is available only via Nix. All commands must run through:

```bash
nix develop --command bash -c '...'
```

## Commands

- **Test:** `go test ./... -v`
- **Build:** `go build -o tgdash .`
- **Vet:** `go vet ./...`

## Pre-commit Hooks

govet, gofmt, markdownlint, nixfmt, actionlint, yamllint,
shellcheck, convco-check.

Always fix linting problems. Never exclude or bypass hooks.

## Commits

Conventional commits enforced by `convco-check`.

## Architecture

```text
stdin -> reader -> parser -> processor -> state -> Bubble Tea TUI
```

## File Structure

- `main.go` - Entry point, wires reader + TUI
- `internal/reader/` - Reads stdin lines, sends Bubble Tea
  messages
- `internal/parser/` - Regex extraction: unit prefix, plan
  summary, apply result, errors
- `internal/processor/` - Feeds parsed lines into AppState
- `internal/state/` - AppState, Unit model, run history
- `internal/estimator/` - Time estimates from run history
- `internal/tui/` - Bubble Tea model, views, keybindings,
  styles
- `internal/integration/` - Integration tests with fixture
  files

## Integration Test Fixtures

Test fixtures live in `internal/integration/testdata/*.txt`.
Each file contains realistic Terragrunt output.

To add an edge case from a bug report:

1. Save output as `testdata/issue_NNN_description.txt`
2. Add a table entry to `TestIntegrationPipeline`
3. Run tests - failure means you have a reproduction

## Worktrees

Worktree location: `.worktrees/` (gitignored).
