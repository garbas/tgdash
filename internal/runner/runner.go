package runner

import (
	"io"
	"os"
	"os/exec"
)

// Run starts the given command and merges its stdout and stderr
// into a single reader. It returns the reader, a cleanup function
// that kills the process, and any startup error.
func Run(args []string) (io.Reader, func(), error) {
	args = stripLeadingDash(args)
	if len(args) == 0 {
		return nil, func() {}, io.EOF
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin

	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		pw.Close()
		pr.Close()
		return nil, func() {}, err
	}

	go func() {
		_ = cmd.Wait()
		pw.Close()
	}()

	cleanup := func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}

	return pr, cleanup, nil
}

// DetectCommand inspects the args for "plan" or "apply" and
// returns the matching word; otherwise returns the first arg.
func DetectCommand(args []string) string {
	args = stripLeadingDash(args)
	for _, a := range args {
		if a == "plan" || a == "apply" {
			return a
		}
	}
	if len(args) > 0 {
		return args[0]
	}
	return "unknown"
}

func stripLeadingDash(args []string) []string {
	if len(args) > 0 && args[0] == "--" {
		return args[1:]
	}
	return args
}
