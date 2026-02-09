# tgdash

**[tgdash.garbas.si](https://tgdash.garbas.si)**

A terminal dashboard for Terragrunt. Pipe the output of
`terragrunt run --all` into `tgdash` and get a real-time
TUI showing per-unit status, plan summaries, errors, and
time estimates.

## The Problem

Running `terragrunt run --all plan` on a large
infrastructure repo produces a wall of interleaved output
from dozens of units. Finding which units failed, what
changed, and how far along the run is requires scrolling
through hundreds of lines.

## Usage

```bash
terragrunt run --all plan 2>&1 | tgdash
```

```bash
terragrunt run --all apply 2>&1 | tgdash
```

Any command that produces `[unit/path]`-prefixed output
works:

```bash
terragrunt run --all validate 2>&1 | tgdash
```

## Features

- Real-time streaming of Terragrunt output
- Per-unit status tracking (waiting, running, done, error)
- Plan summary extraction (resources to add/change/destroy)
- Apply result detection
- Error detection and filtering
- Time estimates based on run history
- Two views: dashboard (detail) and list (overview)
- Vim-style navigation

## Keybindings

| Key | Action |
| ---------- | ---------------- |
| j / Down | Move down |
| k / Up | Move up |
| gg | Jump to top |
| G | Jump to bottom |
| Enter | Expand/collapse |
| Tab | Switch view |
| e | Toggle errors |
| / | Search |
| PgUp/PgDn | Scroll output |
| q / Ctrl+C | Quit |

## Installation

### From source

```bash
go install github.com/rok/tgdash@latest
```

### With Nix

```bash
nix run github:rok/tgdash
```

Or add to your flake inputs:

```nix
{
  inputs.tgdash.url = "github:rok/tgdash";
}
```

## Architecture

```text
stdin -> reader -> parser -> processor -> state -> TUI
```

| Package | Role |
| ------------ | ---------------------------------- |
| `reader` | Reads stdin, emits Bubble Tea msgs |
| `parser` | Regex: prefix, plan, apply, errors |
| `processor` | Routes parsed lines into AppState |
| `state` | Unit model, run history |
| `estimator` | Time estimates from past runs |
| `tui` | Bubble Tea views and keybindings |
| `integration` | Fixture-based integration tests |

## Development

Go is only available via Nix. All commands run through:

```bash
nix develop --command bash -c '...'
```

### Build

```bash
nix develop --command bash -c 'go build -o tgdash .'
```

### Test

```bash
nix develop --command bash -c 'go test ./... -v'
```

### Lint

Pre-commit hooks run automatically: govet, gofmt,
markdownlint, nixfmt, actionlint, yamllint, shellcheck,
convco-check.

### Commits

Conventional commits enforced by `convco-check`. Examples:

```text
feat: add retry logic for failed units
fix: handle empty plan output
test: add edge case for interleaved errors
docs: update keybindings table
```

### Release

Releases use GoReleaser. To create a release:

```bash
git tag v0.1.0
git push --tags
```

GoReleaser builds cross-platform binaries for Linux and
macOS (amd64 + arm64).

## License

[MIT](LICENSE)
