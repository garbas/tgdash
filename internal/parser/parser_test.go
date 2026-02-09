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
