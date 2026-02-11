package runner

import (
	"io"
	"strings"
	"testing"
	"time"
)

func TestRunCapturesBothStreams(t *testing.T) {
	// "echo" writes to stdout, ">&2 echo" writes to stderr
	r, cleanup, err := Run([]string{
		"bash", "-c",
		"echo stdout-line; echo stderr-line >&2",
	})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	defer cleanup()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	got := string(out)
	if !strings.Contains(got, "stdout-line") {
		t.Errorf("missing stdout-line in output: %q", got)
	}
	if !strings.Contains(got, "stderr-line") {
		t.Errorf("missing stderr-line in output: %q", got)
	}
}

func TestRunCleanupKillsProcess(t *testing.T) {
	r, cleanup, err := Run([]string{"sleep", "60"})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	cleanup()

	// After kill, reading should return quickly
	done := make(chan struct{})
	go func() {
		_, _ = io.ReadAll(r)
		close(done)
	}()

	select {
	case <-done:
		// good
	case <-time.After(3 * time.Second):
		t.Fatal("reader did not close after cleanup")
	}
}

func TestRunStripsLeadingDash(t *testing.T) {
	r, cleanup, err := Run([]string{"--", "echo", "hello"})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	defer cleanup()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	got := strings.TrimSpace(string(out))
	if got != "hello" {
		t.Errorf("expected %q, got %q", "hello", got)
	}
}

func TestDetectCommand(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"terragrunt", "run", "--all", "plan"}, "plan"},
		{[]string{"terragrunt", "run", "--all", "apply"}, "apply"},
		{[]string{"--", "terragrunt", "run", "--all", "plan"}, "plan"},
		{[]string{"terragrunt", "run", "--all"}, "terragrunt"},
		{[]string{"echo", "hello"}, "echo"},
	}

	for _, tt := range tests {
		got := DetectCommand(tt.args)
		if got != tt.want {
			t.Errorf("DetectCommand(%v) = %q, want %q",
				tt.args, got, tt.want)
		}
	}
}
