---
title: "Architecture"
description: "How tgdash works internally"
weight: 4
---

## Pipeline

tgdash processes Terragrunt output through a pipeline:

```text
stdin -> reader -> parser -> processor -> state -> TUI
```

## Packages

| Package | Role |
| ------------ | ---------------------------------- |
| `runner` | Spawns commands in exec mode |
| `reader` | Reads stdin, emits Bubble Tea msgs |
| `parser` | Regex: prefix, plan, apply, errors |
| `processor` | Routes parsed lines into AppState |
| `state` | Unit model, run history |
| `estimator` | Time estimates from past runs |
| `tui` | Bubble Tea views and keybindings |
| `integration` | Fixture-based integration tests |

## How it works

### Runner

In exec mode (`tgdash -- command`), the runner spawns
the child process and provides its output as a reader.
It also detects the Terragrunt command (plan, apply,
etc.) from the arguments.

### Reader

The reader goroutine reads lines from the input source
(stdin in pipe mode, or the child process in exec
mode) and sends them as Bubble Tea messages to the
TUI. This allows the TUI to remain responsive while
processing input.

### Parser

Each line is parsed with regex to extract:

- **Unit prefix** - The `[unit/path]` at the start
- **Plan summary** - Lines matching
  `Plan: N to add, N to change, N to destroy`
- **Apply result** - Lines matching
  `Apply complete! Resources: ...`
- **Errors** - Terraform error blocks

### Processor

The processor takes parsed lines and updates the
application state. It tracks which units exist, their
current status, and accumulates output per unit.

### State

AppState holds the full model: list of units, their
statuses, plan summaries, errors, and output lines.
Units progress through states: waiting, running, done,
or error.

### Estimator

The estimator uses historical run data (stored in
`~/.local/share/tgdash/history.json`) to predict how
long each unit will take based on past runs.

### TUI

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea),
the TUI renders the dashboard and list views, handles
keybindings, and updates in real-time as new messages
arrive.
