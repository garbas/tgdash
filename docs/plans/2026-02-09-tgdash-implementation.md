# tgdash Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use
> superpowers:executing-plans to implement this plan
> task-by-task.

**Goal:** Build a Go CLI TUI that reads piped Terragrunt
output from stdin and displays it grouped by unit with
status tracking, plan summaries, error aggregation, and
time estimation.

**Architecture:** Piped stdin -> line parser (goroutine)
-> state store -> Bubble Tea TUI. Two switchable views:
dashboard (table + viewport) and collapsible list. Run
history persisted to `~/.tgdash/history.json` for time
estimation.

**Tech Stack:** Go, Bubble Tea, Bubbles, Lip Gloss

---

## Task 1: Go module init + dependencies

**Files:**

- Create: `go.mod`
- Create: `go.sum`
- Create: `main.go` (minimal placeholder)

### 1.1 Initialize Go module

Run:

```bash
cd /path/to/tgdash && go mod init github.com/garbas/tgdash
```

### 1.2 Add dependencies

Run:

```bash
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/bubbles@latest
go get github.com/charmbracelet/lipgloss@latest
```

### 1.3 Create minimal main.go

```go
package main

import "fmt"

func main() {
    fmt.Println("tgdash")
}
```

### 1.4 Verify it builds

Run: `go build -o tgdash .`

Expected: binary `tgdash` created, no errors.

### 1.5 Commit

```bash
git add go.mod go.sum main.go
git commit -m "feat: initialize Go module with dependencies"
```

---

## Task 2: Data types and state store

**Files:**

- Create: `internal/state/state.go`
- Create: `internal/state/state_test.go`

### 2.1 Write tests for state operations

```go
package state

import (
    "testing"
)

func TestNewAppState(t *testing.T) {
    s := NewAppState()
    if s == nil {
        t.Fatal("expected non-nil state")
    }
    if len(s.Units()) != 0 {
        t.Fatalf("expected 0 units, got %d",
            len(s.Units()))
    }
    if s.InputDone {
        t.Fatal("expected InputDone to be false")
    }
}

func TestGetOrCreateUnit(t *testing.T) {
    s := NewAppState()
    u := s.GetOrCreateUnit("live/dev/vpc")
    if u.Path != "live/dev/vpc" {
        t.Fatalf("expected path live/dev/vpc, got %s",
            u.Path)
    }
    if u.Status != StatusWaiting {
        t.Fatalf("expected status waiting, got %s",
            u.Status)
    }

    // Same path returns same unit
    u2 := s.GetOrCreateUnit("live/dev/vpc")
    if u != u2 {
        t.Fatal("expected same unit pointer")
    }

    // Different path creates new unit
    u3 := s.GetOrCreateUnit("live/dev/rds")
    if u3 == u {
        t.Fatal("expected different unit")
    }
}

func TestUnitsOrdering(t *testing.T) {
    s := NewAppState()
    s.GetOrCreateUnit("b")
    s.GetOrCreateUnit("a")
    s.GetOrCreateUnit("c")
    units := s.Units()
    if len(units) != 3 {
        t.Fatalf("expected 3 units, got %d", len(units))
    }
    // Insertion order preserved
    if units[0].Path != "b" {
        t.Fatalf("expected b first, got %s",
            units[0].Path)
    }
    if units[1].Path != "a" {
        t.Fatalf("expected a second, got %s",
            units[1].Path)
    }
    if units[2].Path != "c" {
        t.Fatalf("expected c third, got %s",
            units[2].Path)
    }
}

func TestAppendLine(t *testing.T) {
    s := NewAppState()
    u := s.GetOrCreateUnit("vpc")
    u.AppendLine("line 1")
    u.AppendLine("line 2")
    if len(u.OutputLines) != 2 {
        t.Fatalf("expected 2 lines, got %d",
            len(u.OutputLines))
    }
}

func TestAppendLineMaxBuffer(t *testing.T) {
    s := NewAppState()
    u := s.GetOrCreateUnit("vpc")
    for i := 0; i < MaxOutputLines+100; i++ {
        u.AppendLine("line")
    }
    if len(u.OutputLines) != MaxOutputLines {
        t.Fatalf("expected %d lines, got %d",
            MaxOutputLines, len(u.OutputLines))
    }
}
```

### 2.2 Run tests to verify they fail

Run: `go test ./internal/state/ -v`

Expected: FAIL — package doesn't exist yet.

### 2.3 Implement state types

```go
package state

import "time"

type UnitStatus string

const (
    StatusWaiting UnitStatus = "waiting"
    StatusRunning UnitStatus = "running"
    StatusDone    UnitStatus = "done"
    StatusError   UnitStatus = "error"
)

const MaxOutputLines = 10000

type PlanSummary struct {
    Add     int
    Change  int
    Destroy int
}

type Unit struct {
    Path        string
    Status      UnitStatus
    OutputLines []string
    PlanSummary *PlanSummary
    Errors      []string
    StartTime   time.Time
    Duration    time.Duration
}

func (u *Unit) AppendLine(line string) {
    u.OutputLines = append(u.OutputLines, line)
    if len(u.OutputLines) > MaxOutputLines {
        u.OutputLines = u.OutputLines[1:]
    }
}

type ViewMode int

const (
    ViewDashboard ViewMode = iota
    ViewList
)

type AppState struct {
    units       []*Unit
    unitIndex   map[string]*Unit
    ActiveView  ViewMode
    SelectedIdx int
    Filter      string
    InputDone   bool
}

func NewAppState() *AppState {
    return &AppState{
        unitIndex: make(map[string]*Unit),
    }
}

func (s *AppState) GetOrCreateUnit(path string) *Unit {
    if u, ok := s.unitIndex[path]; ok {
        return u
    }
    u := &Unit{
        Path:   path,
        Status: StatusWaiting,
    }
    s.units = append(s.units, u)
    s.unitIndex[path] = u
    return u
}

func (s *AppState) Units() []*Unit {
    return s.units
}
```

### 2.4 Run tests to verify they pass

Run: `go test ./internal/state/ -v`

Expected: all PASS.

### 2.5 Commit

```bash
git add internal/state/
git commit -m "feat: add state types and unit store"
```

---

## Task 3: Line parser

**Files:**

- Create: `internal/parser/parser.go`
- Create: `internal/parser/patterns.go`
- Create: `internal/parser/parser_test.go`

### 3.1 Write parser tests

```go
package parser

import (
    "testing"
)

func TestParseLineWithPrefix(t *testing.T) {
    tests := []struct {
        input    string
        wantUnit string
        wantLine string
    }{
        {
            input:    "[live/dev/vpc] Initializing...",
            wantUnit: "live/dev/vpc",
            wantLine: "Initializing...",
        },
        {
            input:    "[app/prod/ecs] Plan: 3 to add, 1 to change, 0 to destroy.",
            wantUnit: "app/prod/ecs",
            wantLine: "Plan: 3 to add, 1 to change, 0 to destroy.",
        },
        {
            input:    "No prefix here",
            wantUnit: "",
            wantLine: "No prefix here",
        },
        {
            input:    "[unit-1] ",
            wantUnit: "unit-1",
            wantLine: "",
        },
    }

    for _, tt := range tests {
        unit, line := ExtractUnitPrefix(tt.input)
        if unit != tt.wantUnit {
            t.Errorf("ExtractUnitPrefix(%q) unit = %q, want %q",
                tt.input, unit, tt.wantUnit)
        }
        if line != tt.wantLine {
            t.Errorf("ExtractUnitPrefix(%q) line = %q, want %q",
                tt.input, line, tt.wantLine)
        }
    }
}

func TestDetectPlanSummary(t *testing.T) {
    tests := []struct {
        line    string
        wantOk  bool
        wantAdd int
        wantChg int
        wantDel int
    }{
        {
            line:    "Plan: 3 to add, 1 to change, 0 to destroy.",
            wantOk:  true,
            wantAdd: 3, wantChg: 1, wantDel: 0,
        },
        {
            line:    "Plan: 0 to add, 0 to change, 8 to destroy.",
            wantOk:  true,
            wantAdd: 0, wantChg: 0, wantDel: 8,
        },
        {
            line:   "Some other line",
            wantOk: false,
        },
    }

    for _, tt := range tests {
        summary, ok := DetectPlanSummary(tt.line)
        if ok != tt.wantOk {
            t.Errorf("DetectPlanSummary(%q) ok = %v, want %v",
                tt.line, ok, tt.wantOk)
            continue
        }
        if ok {
            if summary.Add != tt.wantAdd ||
                summary.Change != tt.wantChg ||
                summary.Destroy != tt.wantDel {
                t.Errorf(
                    "DetectPlanSummary(%q) = %+v, want +%d ~%d -%d",
                    tt.line, summary,
                    tt.wantAdd, tt.wantChg, tt.wantDel)
            }
        }
    }
}

func TestDetectApplyResult(t *testing.T) {
    tests := []struct {
        line    string
        wantOk  bool
        wantAdd int
        wantChg int
        wantDel int
    }{
        {
            line:    "Apply complete! Resources: 2 added, 1 changed, 0 destroyed.",
            wantOk:  true,
            wantAdd: 2, wantChg: 1, wantDel: 0,
        },
        {
            line:   "Not an apply result",
            wantOk: false,
        },
    }

    for _, tt := range tests {
        summary, ok := DetectApplyResult(tt.line)
        if ok != tt.wantOk {
            t.Errorf("DetectApplyResult(%q) ok = %v, want %v",
                tt.line, ok, tt.wantOk)
            continue
        }
        if ok && summary.Add != tt.wantAdd {
            t.Errorf("DetectApplyResult(%q) add = %d, want %d",
                tt.line, summary.Add, tt.wantAdd)
        }
    }
}

func TestDetectError(t *testing.T) {
    tests := []struct {
        line string
        want bool
    }{
        {"Error: Invalid resource type", true},
        {"\u2502 Error: Missing argument", true},
        {"│ Error: Something wrong", true},
        {"No error here", false},
        {"error in lowercase", false},
    }

    for _, tt := range tests {
        got := DetectError(tt.line)
        if got != tt.want {
            t.Errorf("DetectError(%q) = %v, want %v",
                tt.line, got, tt.want)
        }
    }
}
```

### 3.2 Run tests to verify they fail

Run: `go test ./internal/parser/ -v`

Expected: FAIL.

### 3.3 Implement patterns.go

```go
package parser

import (
    "regexp"
    "strconv"
)

var (
    unitPrefixRe = regexp.MustCompile(
        `^\[([^\]]+)\]\s*(.*)$`)
    planSummaryRe = regexp.MustCompile(
        `Plan:\s+(\d+)\s+to add,\s+(\d+)\s+to change,\s+(\d+)\s+to destroy`)
    applyResultRe = regexp.MustCompile(
        `Apply complete!\s+Resources:\s+(\d+)\s+added,\s+(\d+)\s+changed,\s+(\d+)\s+destroyed`)
    errorRe = regexp.MustCompile(
        `(?:^|\s*[│\|]\s*)Error:\s+`)
)

type Summary struct {
    Add     int
    Change  int
    Destroy int
}

func extractThreeInts(
    m []string, i1, i2, i3 int,
) Summary {
    a, _ := strconv.Atoi(m[i1])
    b, _ := strconv.Atoi(m[i2])
    c, _ := strconv.Atoi(m[i3])
    return Summary{Add: a, Change: b, Destroy: c}
}

func DetectPlanSummary(line string) (Summary, bool) {
    m := planSummaryRe.FindStringSubmatch(line)
    if m == nil {
        return Summary{}, false
    }
    return extractThreeInts(m, 1, 2, 3), true
}

func DetectApplyResult(line string) (Summary, bool) {
    m := applyResultRe.FindStringSubmatch(line)
    if m == nil {
        return Summary{}, false
    }
    return extractThreeInts(m, 1, 2, 3), true
}

func DetectError(line string) bool {
    return errorRe.MatchString(line)
}
```

### 3.4 Implement parser.go

```go
package parser

func ExtractUnitPrefix(
    raw string,
) (unit string, line string) {
    m := unitPrefixRe.FindStringSubmatch(raw)
    if m == nil {
        return "", raw
    }
    return m[1], m[2]
}
```

### 3.5 Run tests to verify they pass

Run: `go test ./internal/parser/ -v`

Expected: all PASS.

### 3.6 Commit

```bash
git add internal/parser/
git commit -m "feat: add line parser with pattern detection"
```

---

## Task 4: Stdin reader goroutine

**Files:**

- Create: `internal/reader/reader.go`
- Create: `internal/reader/reader_test.go`

### 4.1 Write tests for reader

```go
package reader

import (
    "strings"
    "testing"
    "time"

    tea "github.com/charmbracelet/bubbletea"
)

func TestReadLines(t *testing.T) {
    input := "line1\nline2\nline3\n"
    r := strings.NewReader(input)

    msgs := make([]tea.Msg, 0)
    ch := make(chan tea.Msg, 10)

    go func() {
        ReadLines(r, func(msg tea.Msg) {
            ch <- msg
        })
        close(ch)
    }()

    for msg := range ch {
        msgs = append(msgs, msg)
    }

    // 3 LineMsg + 1 InputDoneMsg
    lineCount := 0
    doneCount := 0
    for _, msg := range msgs {
        switch msg.(type) {
        case LineMsg:
            lineCount++
        case InputDoneMsg:
            doneCount++
        }
    }

    if lineCount != 3 {
        t.Fatalf("expected 3 LineMsg, got %d",
            lineCount)
    }
    if doneCount != 1 {
        t.Fatalf("expected 1 InputDoneMsg, got %d",
            doneCount)
    }
}

func TestReadLinesContent(t *testing.T) {
    input := "[vpc] hello\n[rds] world\n"
    r := strings.NewReader(input)

    var lines []string
    ch := make(chan tea.Msg, 10)

    go func() {
        ReadLines(r, func(msg tea.Msg) {
            ch <- msg
        })
        close(ch)
    }()

    timeout := time.After(2 * time.Second)
    for {
        select {
        case msg, ok := <-ch:
            if !ok {
                goto done
            }
            if lm, ok := msg.(LineMsg); ok {
                lines = append(lines, lm.Raw)
            }
        case <-timeout:
            t.Fatal("timeout waiting for messages")
        }
    }
done:
    if len(lines) != 2 {
        t.Fatalf("expected 2 lines, got %d", len(lines))
    }
    if lines[0] != "[vpc] hello" {
        t.Fatalf("expected [vpc] hello, got %s",
            lines[0])
    }
}
```

### 4.2 Run tests to verify they fail

Run: `go test ./internal/reader/ -v`

Expected: FAIL.

### 4.3 Implement reader

```go
package reader

import (
    "bufio"
    "io"

    tea "github.com/charmbracelet/bubbletea"
)

type LineMsg struct {
    Raw string
}

type InputDoneMsg struct{}

func ReadLines(r io.Reader, send func(tea.Msg)) {
    scanner := bufio.NewScanner(r)
    for scanner.Scan() {
        send(LineMsg{Raw: scanner.Text()})
    }
    send(InputDoneMsg{})
}
```

### 4.4 Run tests to verify they pass

Run: `go test ./internal/reader/ -v`

Expected: all PASS.

### 4.5 Commit

```bash
git add internal/reader/
git commit -m "feat: add stdin line reader"
```

---

## Task 5: Line processing (parser + state integration)

**Files:**

- Create: `internal/processor/processor.go`
- Create: `internal/processor/processor_test.go`

### 5.1 Write tests

```go
package processor

import (
    "testing"

    "github.com/garbas/tgdash/internal/state"
)

func TestProcessLineCreatesUnit(t *testing.T) {
    s := state.NewAppState()
    p := New(s)
    p.ProcessLine("[live/dev/vpc] Initializing...")

    units := s.Units()
    if len(units) != 1 {
        t.Fatalf("expected 1 unit, got %d", len(units))
    }
    if units[0].Path != "live/dev/vpc" {
        t.Fatalf("expected path live/dev/vpc, got %s",
            units[0].Path)
    }
    if len(units[0].OutputLines) != 1 {
        t.Fatalf("expected 1 output line, got %d",
            len(units[0].OutputLines))
    }
}

func TestProcessLinePlanSummary(t *testing.T) {
    s := state.NewAppState()
    p := New(s)
    p.ProcessLine(
        "[vpc] Plan: 3 to add, 1 to change, 0 to destroy.")

    u := s.Units()[0]
    if u.PlanSummary == nil {
        t.Fatal("expected plan summary to be set")
    }
    if u.PlanSummary.Add != 3 {
        t.Fatalf("expected add=3, got %d",
            u.PlanSummary.Add)
    }
    if u.PlanSummary.Change != 1 {
        t.Fatalf("expected change=1, got %d",
            u.PlanSummary.Change)
    }
}

func TestProcessLineApplyResult(t *testing.T) {
    s := state.NewAppState()
    p := New(s)
    p.ProcessLine(
        "[vpc] Apply complete! Resources: 2 added, 1 changed, 0 destroyed.")

    u := s.Units()[0]
    if u.PlanSummary == nil {
        t.Fatal("expected plan summary from apply result")
    }
    if u.PlanSummary.Add != 2 {
        t.Fatalf("expected add=2, got %d",
            u.PlanSummary.Add)
    }
    if u.Status != state.StatusDone {
        t.Fatalf("expected status done, got %s",
            u.Status)
    }
}

func TestProcessLineError(t *testing.T) {
    s := state.NewAppState()
    p := New(s)
    p.ProcessLine("[vpc] Error: Invalid resource type")

    u := s.Units()[0]
    if u.Status != state.StatusError {
        t.Fatalf("expected status error, got %s",
            u.Status)
    }
    if len(u.Errors) != 1 {
        t.Fatalf("expected 1 error, got %d",
            len(u.Errors))
    }
}

func TestProcessLineNoPrefixAttachesToLast(t *testing.T) {
    s := state.NewAppState()
    p := New(s)
    p.ProcessLine("[vpc] First line")
    p.ProcessLine("  continuation line")

    units := s.Units()
    if len(units) != 1 {
        t.Fatalf("expected 1 unit, got %d", len(units))
    }
    if len(units[0].OutputLines) != 2 {
        t.Fatalf("expected 2 lines, got %d",
            len(units[0].OutputLines))
    }
}

func TestProcessLineNoPrefixNoUnit(t *testing.T) {
    s := state.NewAppState()
    p := New(s)
    p.ProcessLine("global message before any unit")

    // Should create a global unit
    units := s.Units()
    if len(units) != 1 {
        t.Fatalf("expected 1 unit, got %d", len(units))
    }
    if units[0].Path != GlobalUnitPath {
        t.Fatalf("expected global unit, got %s",
            units[0].Path)
    }
}
```

### 5.2 Run tests to verify they fail

Run: `go test ./internal/processor/ -v`

Expected: FAIL.

### 5.3 Implement processor

```go
package processor

import (
    "time"

    "github.com/garbas/tgdash/internal/parser"
    "github.com/garbas/tgdash/internal/state"
)

const GlobalUnitPath = "(global)"

type Processor struct {
    state    *state.AppState
    lastUnit string
}

func New(s *state.AppState) *Processor {
    return &Processor{state: s}
}

func (p *Processor) ProcessLine(raw string) {
    unitPath, line := parser.ExtractUnitPrefix(raw)

    if unitPath == "" {
        unitPath = p.lastUnit
        line = raw
    }
    if unitPath == "" {
        unitPath = GlobalUnitPath
    }

    p.lastUnit = unitPath

    u := p.state.GetOrCreateUnit(unitPath)
    u.AppendLine(line)

    // Detect plan summary
    if summary, ok := parser.DetectPlanSummary(line); ok {
        u.PlanSummary = &state.PlanSummary{
            Add:     summary.Add,
            Change:  summary.Change,
            Destroy: summary.Destroy,
        }
    }

    // Detect apply result
    if summary, ok := parser.DetectApplyResult(line); ok {
        u.PlanSummary = &state.PlanSummary{
            Add:     summary.Add,
            Change:  summary.Change,
            Destroy: summary.Destroy,
        }
        u.Status = state.StatusDone
        if !u.StartTime.IsZero() {
            u.Duration = time.Since(u.StartTime)
        }
    }

    // Detect errors
    if parser.DetectError(line) {
        u.Status = state.StatusError
        u.Errors = append(u.Errors, line)
    }

    // If unit was waiting and we got output, it's running
    if u.Status == state.StatusWaiting {
        u.Status = state.StatusRunning
        u.StartTime = time.Now()
    }
}
```

### 5.4 Run tests to verify they pass

Run: `go test ./internal/processor/ -v`

Expected: all PASS.

### 5.5 Commit

```bash
git add internal/processor/
git commit -m "feat: add line processor integrating parser and state"
```

---

## Task 6: History persistence

**Files:**

- Create: `internal/state/history.go`
- Create: `internal/state/history_test.go`

### 6.1 Write tests

```go
package state

import (
    "os"
    "path/filepath"
    "testing"
    "time"
)

func TestSaveAndLoadHistory(t *testing.T) {
    dir := t.TempDir()
    path := filepath.Join(dir, "history.json")

    run := RunRecord{
        Command:   "plan",
        Timestamp: time.Now(),
        Units: []UnitRecord{
            {
                Path:     "live/dev/vpc",
                Status:   StatusDone,
                Duration: 12 * time.Second,
                PlanSummary: &PlanSummary{
                    Add: 3, Change: 1, Destroy: 0,
                },
            },
        },
    }

    h := &History{Runs: []RunRecord{run}}
    err := SaveHistory(path, h)
    if err != nil {
        t.Fatalf("save failed: %v", err)
    }

    loaded, err := LoadHistory(path)
    if err != nil {
        t.Fatalf("load failed: %v", err)
    }
    if len(loaded.Runs) != 1 {
        t.Fatalf("expected 1 run, got %d",
            len(loaded.Runs))
    }
    if loaded.Runs[0].Units[0].Path != "live/dev/vpc" {
        t.Fatalf("expected vpc, got %s",
            loaded.Runs[0].Units[0].Path)
    }
}

func TestLoadHistoryMissing(t *testing.T) {
    path := filepath.Join(t.TempDir(), "nope.json")
    h, err := LoadHistory(path)
    if err != nil {
        t.Fatalf("expected no error for missing file, got %v",
            err)
    }
    if len(h.Runs) != 0 {
        t.Fatalf("expected 0 runs, got %d",
            len(h.Runs))
    }
}

func TestLoadHistoryCorrupted(t *testing.T) {
    dir := t.TempDir()
    path := filepath.Join(dir, "history.json")
    os.WriteFile(path, []byte("not json{{{"), 0644)

    h, err := LoadHistory(path)
    if err != nil {
        t.Fatalf("expected no error for corrupted file, got %v",
            err)
    }
    if len(h.Runs) != 0 {
        t.Fatalf("expected 0 runs for corrupted file, got %d",
            len(h.Runs))
    }
}

func TestHistoryMaxRuns(t *testing.T) {
    dir := t.TempDir()
    path := filepath.Join(dir, "history.json")

    h := &History{}
    for i := 0; i < MaxHistoryRuns+50; i++ {
        h.Runs = append(h.Runs, RunRecord{
            Command:   "plan",
            Timestamp: time.Now(),
        })
    }
    err := SaveHistory(path, h)
    if err != nil {
        t.Fatalf("save failed: %v", err)
    }

    loaded, err := LoadHistory(path)
    if err != nil {
        t.Fatalf("load failed: %v", err)
    }
    if len(loaded.Runs) != MaxHistoryRuns {
        t.Fatalf("expected %d runs, got %d",
            MaxHistoryRuns, len(loaded.Runs))
    }
}
```

### 6.2 Run tests to verify they fail

Run: `go test ./internal/state/ -v -run History`

Expected: FAIL.

### 6.3 Implement history

```go
package state

import (
    "encoding/json"
    "errors"
    "os"
    "path/filepath"
    "time"
)

const MaxHistoryRuns = 100

type UnitRecord struct {
    Path        string        `json:"path"`
    Status      UnitStatus    `json:"status"`
    Duration    time.Duration `json:"duration"`
    PlanSummary *PlanSummary  `json:"plan_summary,omitempty"`
}

type RunRecord struct {
    Command   string       `json:"command"`
    Timestamp time.Time    `json:"timestamp"`
    Units     []UnitRecord `json:"units"`
}

type History struct {
    Runs []RunRecord `json:"runs"`
}

func DefaultHistoryPath() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".tgdash", "history.json")
}

func LoadHistory(path string) (*History, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        if errors.Is(err, os.ErrNotExist) {
            return &History{}, nil
        }
        return &History{}, nil
    }
    var h History
    if err := json.Unmarshal(data, &h); err != nil {
        return &History{}, nil
    }
    return &h, nil
}

func SaveHistory(path string, h *History) error {
    if len(h.Runs) > MaxHistoryRuns {
        h.Runs = h.Runs[len(h.Runs)-MaxHistoryRuns:]
    }

    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }

    data, err := json.MarshalIndent(h, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(path, data, 0644)
}
```

### 6.4 Run tests to verify they pass

Run: `go test ./internal/state/ -v`

Expected: all PASS.

### 6.5 Commit

```bash
git add internal/state/history.go internal/state/history_test.go
git commit -m "feat: add run history persistence"
```

---

## Task 7: Time estimator

**Files:**

- Create: `internal/estimator/estimator.go`
- Create: `internal/estimator/estimator_test.go`

### 7.1 Write tests

```go
package estimator

import (
    "testing"
    "time"

    "github.com/garbas/tgdash/internal/state"
)

func TestEstimateNoHistory(t *testing.T) {
    e := New(&state.History{})
    est, ok := e.Estimate("vpc", "plan")
    if ok {
        t.Fatalf("expected no estimate, got %v", est)
    }
}

func TestEstimateSingleRun(t *testing.T) {
    h := &state.History{
        Runs: []state.RunRecord{
            {
                Command: "plan",
                Units: []state.UnitRecord{
                    {
                        Path:     "vpc",
                        Status:   state.StatusDone,
                        Duration: 10 * time.Second,
                    },
                },
            },
        },
    }

    e := New(h)
    est, ok := e.Estimate("vpc", "plan")
    if !ok {
        t.Fatal("expected estimate")
    }
    if est != 10*time.Second {
        t.Fatalf("expected 10s, got %v", est)
    }
}

func TestEstimateMedian(t *testing.T) {
    h := &state.History{
        Runs: []state.RunRecord{
            {Command: "plan", Units: []state.UnitRecord{
                {Path: "vpc", Status: state.StatusDone,
                    Duration: 5 * time.Second},
            }},
            {Command: "plan", Units: []state.UnitRecord{
                {Path: "vpc", Status: state.StatusDone,
                    Duration: 15 * time.Second},
            }},
            {Command: "plan", Units: []state.UnitRecord{
                {Path: "vpc", Status: state.StatusDone,
                    Duration: 10 * time.Second},
            }},
        },
    }

    e := New(h)
    est, ok := e.Estimate("vpc", "plan")
    if !ok {
        t.Fatal("expected estimate")
    }
    if est != 10*time.Second {
        t.Fatalf("expected 10s (median), got %v", est)
    }
}

func TestEstimateIgnoresDifferentCommand(t *testing.T) {
    h := &state.History{
        Runs: []state.RunRecord{
            {Command: "apply", Units: []state.UnitRecord{
                {Path: "vpc", Status: state.StatusDone,
                    Duration: 30 * time.Second},
            }},
        },
    }

    e := New(h)
    _, ok := e.Estimate("vpc", "plan")
    if ok {
        t.Fatal("expected no estimate for different command")
    }
}

func TestEstimateIgnoresErrors(t *testing.T) {
    h := &state.History{
        Runs: []state.RunRecord{
            {Command: "plan", Units: []state.UnitRecord{
                {Path: "vpc", Status: state.StatusError,
                    Duration: 2 * time.Second},
            }},
        },
    }

    e := New(h)
    _, ok := e.Estimate("vpc", "plan")
    if ok {
        t.Fatal("expected no estimate for errored runs")
    }
}
```

### 7.2 Run tests to verify they fail

Run: `go test ./internal/estimator/ -v`

Expected: FAIL.

### 7.3 Implement estimator

```go
package estimator

import (
    "sort"
    "time"

    "github.com/garbas/tgdash/internal/state"
)

type Estimator struct {
    history *state.History
}

func New(h *state.History) *Estimator {
    return &Estimator{history: h}
}

func (e *Estimator) Estimate(
    unitPath string, command string,
) (time.Duration, bool) {
    var durations []time.Duration

    for _, run := range e.history.Runs {
        if run.Command != command {
            continue
        }
        for _, u := range run.Units {
            if u.Path == unitPath &&
                u.Status == state.StatusDone {
                durations = append(durations, u.Duration)
            }
        }
    }

    if len(durations) == 0 {
        return 0, false
    }

    sort.Slice(durations, func(i, j int) bool {
        return durations[i] < durations[j]
    })

    median := durations[len(durations)/2]
    return median, true
}
```

### 7.4 Run tests to verify they pass

Run: `go test ./internal/estimator/ -v`

Expected: all PASS.

### 7.5 Commit

```bash
git add internal/estimator/
git commit -m "feat: add time estimator from run history"
```

---

## Task 8: TUI keybindings and styles

**Files:**

- Create: `internal/tui/keys.go`
- Create: `internal/tui/styles.go`

### 8.1 Implement keybindings

```go
package tui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
    Up       key.Binding
    Down     key.Binding
    Top      key.Binding
    Bottom   key.Binding
    Enter    key.Binding
    Tab      key.Binding
    Errors   key.Binding
    Search   key.Binding
    Quit     key.Binding
    PageUp   key.Binding
    PageDown key.Binding
}

func DefaultKeyMap() KeyMap {
    return KeyMap{
        Up: key.NewBinding(
            key.WithKeys("k", "up"),
            key.WithHelp("↑/k", "up"),
        ),
        Down: key.NewBinding(
            key.WithKeys("j", "down"),
            key.WithHelp("↓/j", "down"),
        ),
        Top: key.NewBinding(
            key.WithKeys("g"),
            key.WithHelp("gg", "top"),
        ),
        Bottom: key.NewBinding(
            key.WithKeys("G"),
            key.WithHelp("G", "bottom"),
        ),
        Enter: key.NewBinding(
            key.WithKeys("enter"),
            key.WithHelp("enter", "expand/collapse"),
        ),
        Tab: key.NewBinding(
            key.WithKeys("tab"),
            key.WithHelp("tab", "switch view"),
        ),
        Errors: key.NewBinding(
            key.WithKeys("e"),
            key.WithHelp("e", "errors only"),
        ),
        Search: key.NewBinding(
            key.WithKeys("/"),
            key.WithHelp("/", "search"),
        ),
        Quit: key.NewBinding(
            key.WithKeys("q", "ctrl+c"),
            key.WithHelp("q", "quit"),
        ),
        PageUp: key.NewBinding(
            key.WithKeys("pgup"),
            key.WithHelp("pgup", "page up"),
        ),
        PageDown: key.NewBinding(
            key.WithKeys("pgdown"),
            key.WithHelp("pgdn", "page down"),
        ),
    }
}
```

### 8.2 Implement styles

```go
package tui

import "github.com/charmbracelet/lipgloss"

var (
    StatusWaiting = lipgloss.NewStyle().
            Foreground(lipgloss.Color("240"))
    StatusRunning = lipgloss.NewStyle().
            Foreground(lipgloss.Color("33")).
            Bold(true)
    StatusDone = lipgloss.NewStyle().
            Foreground(lipgloss.Color("34"))
    StatusError = lipgloss.NewStyle().
            Foreground(lipgloss.Color("196")).
            Bold(true)

    TitleStyle = lipgloss.NewStyle().
            Bold(true).
            Foreground(lipgloss.Color("230")).
            Background(lipgloss.Color("63")).
            Padding(0, 1)

    SelectedStyle = lipgloss.NewStyle().
            Foreground(lipgloss.Color("229")).
            Background(lipgloss.Color("57"))

    BorderStyle = lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(lipgloss.Color("63"))

    HelpStyle = lipgloss.NewStyle().
            Foreground(lipgloss.Color("240"))

    AddStyle = lipgloss.NewStyle().
            Foreground(lipgloss.Color("34"))
    ChangeStyle = lipgloss.NewStyle().
            Foreground(lipgloss.Color("214"))
    DestroyStyle = lipgloss.NewStyle().
            Foreground(lipgloss.Color("196"))

    EstimateStyle = lipgloss.NewStyle().
            Foreground(lipgloss.Color("240")).
            Italic(true)
)
```

### 8.3 Verify it compiles

Run: `go build ./internal/tui/`

Expected: no errors.

### 8.4 Commit

```bash
git add internal/tui/keys.go internal/tui/styles.go
git commit -m "feat: add TUI keybindings and styles"
```

---

## Task 9: TUI dashboard view

**Files:**

- Create: `internal/tui/dashboard.go`

### 9.1 Implement dashboard view

This is the table + viewport split view. The dashboard
renders the units table in the top half and the selected
unit's output in a viewport in the bottom half.

```go
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
```

### 9.2 Verify it compiles

Run: `go build ./internal/tui/`

Expected: no errors.

### 9.3 Commit

```bash
git add internal/tui/dashboard.go
git commit -m "feat: add dashboard view (table + viewport)"
```

---

## Task 10: TUI list view

**Files:**

- Create: `internal/tui/list.go`

### 10.1 Implement collapsible list view

```go
package tui

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/lipgloss"
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

        status := formatStatus(u.Status)
        summary := ""
        if u.PlanSummary != nil {
            summary = fmt.Sprintf(" %s%s%s",
                AddStyle.Render(
                    fmt.Sprintf("+%d", u.PlanSummary.Add)),
                ChangeStyle.Render(
                    fmt.Sprintf(" ~%d", u.PlanSummary.Change)),
                DestroyStyle.Render(
                    fmt.Sprintf(" -%d", u.PlanSummary.Destroy)),
            )
        }

        timeStr := formatTime(u, estimates)

        header := fmt.Sprintf("%s %s [%s]%s  %s",
            arrow, u.Path, status, summary, timeStr)

        if i == appState.SelectedIdx {
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
        // Simple scroll: show around selected index
        visible := strings.Join(
            lines[:min(len(lines), l.height)], "\n")
        return visible
    }

    return content
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

### 10.2 Verify it compiles

Run: `go build ./internal/tui/`

Expected: no errors.

### 10.3 Commit

```bash
git add internal/tui/list.go
git commit -m "feat: add collapsible list view"
```

---

## Task 11: TUI root model (Bubble Tea app)

**Files:**

- Create: `internal/tui/app.go`

### 11.1 Implement root Bubble Tea model

This wires everything together: receives `LineMsg` from
the reader goroutine, processes them, and renders the
active view.

```go
package tui

import (
    "github.com/charmbracelet/bubbles/key"
    "github.com/charmbracelet/bubbles/viewport"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/garbas/tgdash/internal/estimator"
    "github.com/garbas/tgdash/internal/processor"
    "github.com/garbas/tgdash/internal/reader"
    "github.com/garbas/tgdash/internal/state"
)

type Model struct {
    state      *state.AppState
    proc       *processor.Processor
    estimator  *estimator.Estimator
    dashboard  DashboardView
    list       ListView
    keys       KeyMap
    width      int
    height     int
    gPending   bool
    estimates  map[string]string
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

    // Pass to viewport if in dashboard view
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
        // Toggle error filter (simplified)
        // Full implementation would filter the unit list

    default:
        m.gPending = false
    }

    // Reset g pending on any non-g key
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
                        1000000000).String() + " est."
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
```

### 11.2 Verify it compiles

Run: `go build ./internal/tui/`

Expected: no errors.

### 11.3 Commit

```bash
git add internal/tui/app.go
git commit -m "feat: add root Bubble Tea model wiring all components"
```

---

## Task 12: main.go — wire it all together

**Files:**

- Modify: `main.go`

### 12.1 Implement main.go

```go
package main

import (
    "fmt"
    "os"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/garbas/tgdash/internal/estimator"
    "github.com/garbas/tgdash/internal/reader"
    "github.com/garbas/tgdash/internal/state"
    "github.com/garbas/tgdash/internal/tui"
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
```

### 12.2 Add golang.org/x/term dependency

Run: `go get golang.org/x/term@latest`

### 12.3 Verify it builds

Run: `go build -o tgdash .`

Expected: binary `tgdash` created.

### 12.4 Manual smoke test

Run:

```bash
echo -e "[vpc] Initializing...\n\
[rds] Initializing...\n\
[vpc] Plan: 3 to add, 1 to change, 0 to destroy.\n\
[rds] Error: Missing argument\n\
[vpc] Apply complete! Resources: 3 added, 1 changed, 0 destroyed." \
  | ./tgdash
```

Expected: TUI opens showing two units (vpc and rds)
with their statuses. Press `Tab` to switch views.
Press `q` to quit.

### 12.5 Commit

```bash
git add main.go go.mod go.sum
git commit -m "feat: wire up main entry point with TUI"
```

---

## Task 13: Run all tests + final verification

### 13.1 Run full test suite

Run: `go test ./... -v`

Expected: all tests pass.

### 13.2 Run build

Run: `go build -o tgdash .`

Expected: clean build.

### 13.3 Run vet and basic checks

Run: `go vet ./...`

Expected: no issues.

### 13.4 Final commit if any fixes needed

```bash
git add -A
git commit -m "fix: address issues from final verification"
```

Only commit if there were actual fixes needed.
