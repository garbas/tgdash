package integration

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/garbas/tgdash/internal/processor"
	"github.com/garbas/tgdash/internal/state"
)

func testdataDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata")
}

func runPipeline(
	t *testing.T, fixture string,
) *state.AppState {
	t.Helper()
	path := filepath.Join(testdataDir(), fixture)
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open fixture %s: %v", fixture, err)
	}
	defer f.Close()

	s := state.NewAppState()
	p := processor.New(s)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		p.ProcessLine(sc.Text())
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("reading fixture %s: %v", fixture, err)
	}
	return s
}

type wantUnit struct {
	path     string
	status   state.UnitStatus
	planAdd  int
	planChg  int
	planDel  int
	hasPlan  bool
	minLines int
	errCount int
}

func TestIntegrationPipeline(t *testing.T) {
	tests := []struct {
		name      string
		fixture   string
		wantCount int
		units     []wantUnit
	}{
		{
			name:      "basic plan with 3 units",
			fixture:   "basic_plan.txt",
			wantCount: 3,
			units: []wantUnit{
				{
					path:     "live/dev/vpc",
					status:   state.StatusRunning,
					hasPlan:  true,
					planAdd:  2,
					planChg:  1,
					planDel:  0,
					minLines: 5,
				},
				{
					path:     "live/dev/ecs",
					status:   state.StatusRunning,
					hasPlan:  true,
					planAdd:  5,
					planChg:  0,
					planDel:  0,
					minLines: 3,
				},
				{
					path:     "live/dev/rds",
					status:   state.StatusDone,
					hasPlan:  true,
					planAdd:  0,
					planChg:  0,
					planDel:  0,
					minLines: 3,
				},
			},
		},
		{
			name:      "plan with errors",
			fixture:   "plan_with_errors.txt",
			wantCount: 2,
			units: []wantUnit{
				{
					path:     "live/prod/api",
					status:   state.StatusRunning,
					hasPlan:  true,
					planAdd:  3,
					planChg:  0,
					planDel:  0,
					minLines: 3,
				},
				{
					path:     "live/prod/db",
					status:   state.StatusError,
					hasPlan:  false,
					errCount: 2,
					minLines: 5,
				},
			},
		},
		{
			name:      "multi unit apply",
			fixture:   "multi_unit_apply.txt",
			wantCount: 4,
			units: []wantUnit{
				{
					path:    "infra/network",
					status:  state.StatusDone,
					hasPlan: true,
					planAdd: 1,
					planChg: 0,
					planDel: 0,
				},
				{
					path:    "infra/dns",
					status:  state.StatusDone,
					hasPlan: true,
					planAdd: 2,
					planChg: 1,
					planDel: 0,
				},
				{
					path:    "infra/certs",
					status:  state.StatusDone,
					hasPlan: true,
					planAdd: 1,
					planChg: 0,
					planDel: 0,
				},
				{
					path:    "infra/lb",
					status:  state.StatusDone,
					hasPlan: true,
					planAdd: 3,
					planChg: 0,
					planDel: 1,
				},
			},
		},
		{
			name:      "new terragrunt format with timestamps",
			fixture:   "terragrunt_new_format.txt",
			wantCount: 5,
			units: []wantUnit{
				{
					path:     processor.GlobalUnitPath,
					status:   state.StatusRunning,
					hasPlan:  false,
					minLines: 1,
				},
				{
					path:     "app-api",
					status:   state.StatusRunning,
					hasPlan:  true,
					planAdd:  3,
					planChg:  1,
					planDel:  0,
					minLines: 5,
				},
				{
					path:     "app-web",
					status:   state.StatusDone,
					hasPlan:  true,
					planAdd:  0,
					planChg:  0,
					planDel:  0,
					minLines: 3,
				},
				{
					path:     "infra",
					status:   state.StatusRunning,
					hasPlan:  true,
					planAdd:  5,
					planChg:  0,
					planDel:  2,
					minLines: 3,
				},
				{
					path:     "database",
					status:   state.StatusError,
					hasPlan:  false,
					errCount: 2,
					minLines: 5,
				},
			},
		},
		{
			name:      "real v0.97.0 output with ANSI codes",
			fixture:   "real_v097_output.txt",
			wantCount: 6,
			units: []wantUnit{
				{
					path:     processor.GlobalUnitPath,
					status:   state.StatusRunning,
					hasPlan:  false,
					minLines: 1,
				},
				{
					path:     "app-api",
					status:   state.StatusRunning,
					hasPlan:  true,
					planAdd:  3,
					planChg:  1,
					planDel:  0,
					minLines: 5,
				},
				{
					path:     "app-web",
					status:   state.StatusDone,
					hasPlan:  true,
					planAdd:  0,
					planChg:  0,
					planDel:  0,
					minLines: 3,
				},
				{
					path:     "infra",
					status:   state.StatusRunning,
					hasPlan:  true,
					planAdd:  5,
					planChg:  0,
					planDel:  2,
					minLines: 3,
				},
				{
					path:     "database",
					status:   state.StatusError,
					hasPlan:  false,
					errCount: 2,
					minLines: 5,
				},
				{
					path:     "web-ui",
					status:   state.StatusSkipped,
					hasPlan:  false,
					minLines: 1,
				},
			},
		},
		{
			name:      "new format with ANSI codes",
			fixture:   "terragrunt_new_format_ansi.txt",
			wantCount: 5,
			units: []wantUnit{
				{
					path:     processor.GlobalUnitPath,
					status:   state.StatusRunning,
					hasPlan:  false,
					minLines: 1,
				},
				{
					path:     "app-api",
					status:   state.StatusRunning,
					hasPlan:  true,
					planAdd:  3,
					planChg:  1,
					planDel:  0,
					minLines: 5,
				},
				{
					path:     "app-web",
					status:   state.StatusDone,
					hasPlan:  true,
					planAdd:  0,
					planChg:  0,
					planDel:  0,
					minLines: 3,
				},
				{
					path:     "infra",
					status:   state.StatusRunning,
					hasPlan:  true,
					planAdd:  5,
					planChg:  0,
					planDel:  2,
					minLines: 3,
				},
				{
					path:     "database",
					status:   state.StatusError,
					hasPlan:  false,
					errCount: 2,
					minLines: 5,
				},
			},
		},
		{
			name:      "interleaved output with global lines",
			fixture:   "interleaved_output.txt",
			wantCount: 4,
			units: []wantUnit{
				{
					path:     processor.GlobalUnitPath,
					status:   state.StatusRunning,
					hasPlan:  false,
					minLines: 2,
				},
				{
					path:     "app/web",
					status:   state.StatusRunning,
					hasPlan:  true,
					planAdd:  4,
					planChg:  2,
					planDel:  1,
					minLines: 5,
				},
				{
					path:     "app/api",
					status:   state.StatusRunning,
					hasPlan:  true,
					planAdd:  2,
					planChg:  0,
					planDel:  0,
					minLines: 3,
				},
				{
					path:     "app/worker",
					status:   state.StatusRunning,
					hasPlan:  true,
					planAdd:  1,
					planChg:  1,
					planDel:  0,
					minLines: 3,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := runPipeline(t, tt.fixture)
			units := s.Units()

			if len(units) != tt.wantCount {
				t.Fatalf(
					"unit count = %d, want %d",
					len(units), tt.wantCount,
				)
			}

			for i, want := range tt.units {
				u := units[i]
				if u.Path != want.path {
					t.Errorf(
						"unit[%d] path = %q, want %q",
						i, u.Path, want.path,
					)
				}
				if u.Status != want.status {
					t.Errorf(
						"unit[%d] %q status = %q, want %q",
						i, u.Path, u.Status, want.status,
					)
				}
				if want.hasPlan {
					if u.PlanSummary == nil {
						t.Errorf(
							"unit[%d] %q: expected plan summary",
							i, u.Path,
						)
					} else {
						if u.PlanSummary.Add != want.planAdd {
							t.Errorf(
								"unit[%d] %q plan add = %d, want %d",
								i, u.Path,
								u.PlanSummary.Add,
								want.planAdd,
							)
						}
						if u.PlanSummary.Change != want.planChg {
							t.Errorf(
								"unit[%d] %q plan change = %d, want %d",
								i, u.Path,
								u.PlanSummary.Change,
								want.planChg,
							)
						}
						if u.PlanSummary.Destroy != want.planDel {
							t.Errorf(
								"unit[%d] %q plan destroy = %d, want %d",
								i, u.Path,
								u.PlanSummary.Destroy,
								want.planDel,
							)
						}
					}
				}
				if want.errCount > 0 {
					if len(u.Errors) != want.errCount {
						t.Errorf(
							"unit[%d] %q error count = %d, want %d",
							i, u.Path,
							len(u.Errors),
							want.errCount,
						)
					}
				}
				if want.minLines > 0 {
					if len(u.OutputLines) < want.minLines {
						t.Errorf(
							"unit[%d] %q lines = %d, want >= %d",
							i, u.Path,
							len(u.OutputLines),
							want.minLines,
						)
					}
				}
			}
		})
	}
}
