# tgdash Design

A Go CLI tool that reads piped Terragrunt output from stdin,
parses interleaved unit output in real-time, and renders a
structured TUI using Charmbracelet libraries.

## Problem

When running `terragrunt run --all plan` or `apply`, multiple
units execute in parallel. Terragrunt prefixes each output line
with `[path/to/unit]`, but the interleaved output makes it
nearly impossible to follow what's happening in any single unit.

## Usage

```bash
terragrunt run --all plan 2>&1 | tgdash
terragrunt run --all apply 2>&1 | tgdash
```

## Architecture

### Data Flow

```text
stdin (piped from terragrunt)
  -> Stdin Reader (goroutine, reads lines)
  -> Line Parser (extracts unit prefix, detects patterns)
  -> State Store (per-unit state: output, status, summaries)
  -> TUI (Bubble Tea renders current state)
```

### Line Parser

Each line from stdin is processed in two steps:

1. **Unit identification** -- extract the `[path/to/unit]` prefix.
   Lines without a prefix attach to the last seen unit. Lines
   before any unit is seen go to a "global" bucket.

2. **Pattern matching** -- scan against compiled regexes:

   - Status: `Executing terraform plan/apply` -> running
   - Plan summary: `Plan: N to add, N to change, N to destroy`
   - Apply result: `Apply complete! Resources: ...`
   - Errors: `Error:` or `| Error:` lines
   - Timing: Terragrunt execution time per unit
   - Completion: unit finished (success or failure)

Unknown line formats are stored as raw output, never dropped.
The parser degrades gracefully if Terragrunt changes its format.

### Data Model

```text
Unit:
  Path         string
  Status       enum(waiting, running, done, error)
  OutputLines  []string
  PlanSummary  {Add, Change, Destroy int} (nil until parsed)
  Errors       []string
  StartTime    time.Time
  Duration     time.Duration

AppState:
  Units        ordered map[string]*Unit
  ActiveView   enum(dashboard, list)
  SelectedIdx  int
  ScrollPos    int
  Filter       string
  InputDone    bool
```

Units are created lazily on first occurrence. Ordering is
by insertion order (first seen in output).

When stdin closes, `InputDone` is set to true and the TUI
stays open for browsing. User quits with `q` or `Ctrl+C`.

## TUI Views

Two views, switchable with `Tab`.

### View A: Dashboard (summary table + detail pane)

```text
+- Units -----------------------------------------+
| STATUS | UNIT            | +  ~  - | TIME        |
| done   | live/dev/vpc    | 3  1  0 | 12s         |
| run    | live/dev/rds    | -  -  - | 8s / ~15s   |
| err    | live/dev/ecs    | 2  0  1 | 5s          |
| wait   | live/dev/dns    | -  -  - | -           |
+- Output: live/dev/rds ---------------------------+
| Terraform will perform the following actions:    |
|                                                  |
|   # aws_db_instance.main will be created         |
|   + resource "aws_db_instance" "main" {          |
|       ...                                        |
+-------------------------------------------------+
```

Navigate units with arrow keys or `j`/`k`. Selected unit's
output streams in the bottom pane. Bottom pane auto-scrolls
but supports manual scroll with `j`/`k`, `PgUp`/`PgDn`,
`G`/`gg`.

### View B: Collapsible list

```text
v live/dev/vpc [done] +3 ~1 -0  12s
  | Terraform will perform the following...
  | Plan: 3 to add, 1 to change, 0 to destroy
> live/dev/rds [running] 8s / ~15s est.
v live/dev/ecs [error] +2 ~0 -1  5s
  | Error: Invalid resource type
  |   on main.tf line 12...
> live/dev/dns [waiting]
```

Expand/collapse with `Enter`. Multiple units can be open.

### Keybindings

| Key | Action |
| --- | ------ |
| `j` / `Down` | Next unit |
| `k` / `Up` | Previous unit |
| `G` | Jump to last unit |
| `gg` | Jump to first unit |
| `Enter` | Expand/collapse (list view) |
| `Tab` | Switch view |
| `e` | Filter to errored units only |
| `/` | Search/filter units by name |
| `q` / `Ctrl+C` | Quit |
| `PgUp` / `PgDn` | Scroll output pane |

## Run Persistence & Time Estimation

### Storage

Completed runs are persisted to `~/.tgdash/history.json`.

```text
Run:
  ID         TBD
  Command    string (plan/apply)
  Timestamp  time.Time
  Units[]:
    Path          string
    Status        enum(done, error)
    Duration      time.Duration
    PlanSummary   {Add, Change, Destroy int}
```

History is capped at the last 100 runs.

### Time Estimation

When a unit starts executing, we look up its path in previous
runs and compute the median duration of the last N runs for
the same unit+command combination.

Displayed as:

```text
live/dev/rds  [running 8s / ~15s est.]
```

The `~` and `est.` make it clear this is an estimate. If no
history exists for a unit, only elapsed time is shown.

### Run ID

The identifier for correlating runs across invocations is TBD.
Will be determined during implementation based on what makes
the most practical sense (stack path, git branch, hash of
unit set, etc.).

## Edge Cases

- **No prefix on a line** -- attach to last seen unit, or
  global bucket if no unit seen yet
- **Not a pipe** -- detect terminal on stdin, print usage
  and exit
- **Huge output** -- cap per-unit buffer at 10,000 lines,
  drop oldest
- **Broken pipe** -- handle SIGPIPE gracefully on early quit
- **Corrupted history** -- log warning, start fresh
- **Small terminal** -- show "resize terminal" message
- **No color support** -- Lip Gloss auto-degrades

## Project Structure

```text
tgdash/
  main.go
  internal/
    parser/
      parser.go
      patterns.go
    state/
      state.go
      history.go
    estimator/
      estimator.go
    tui/
      app.go
      dashboard.go
      list.go
      keys.go
      styles.go
  go.mod
  go.sum
```

## Dependencies

- `github.com/charmbracelet/bubbletea`
- `github.com/charmbracelet/bubbles`
- `github.com/charmbracelet/lipgloss`
- Go standard library for everything else

## Scope

### Phase 1 (now)

- `plan` and `apply` command support
- Both TUI views
- Run persistence and time estimation

### Phase 2 (later)

- `destroy` and other commands
- Run comparison view
- Configurable run ID strategy
