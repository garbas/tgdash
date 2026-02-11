package parser

import "testing"

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain text unchanged",
			input: "hello world",
			want:  "hello world",
		},
		{
			name:  "simple color code",
			input: "\x1b[31mred\x1b[0m",
			want:  "red",
		},
		{
			name:  "256-color code",
			input: "\x1b[0;38;5;66mtext\x1b[0m",
			want:  "text",
		},
		{
			name:  "bold and color",
			input: "\x1b[1m\x1b[34mbold blue\x1b[0m",
			want:  "bold blue",
		},
		{
			name:  "dim gray timestamps",
			input: "\x1b[0;90m23:47:15.520\x1b[0m STDOUT",
			want:  "23:47:15.520 STDOUT",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name: "real terragrunt line",
			input: "\x1b[0;90m23:47:15.520\x1b[0m " +
				"\x1b[0;36mSTDOUT\x1b[0m " +
				"[.terragrunt-stack/vpc] " +
				"\x1b[0;36mterraform:\x1b[0m " +
				"Plan: 2 to add, 1 to change, " +
				"0 to destroy.",
			want: "23:47:15.520 STDOUT " +
				"[.terragrunt-stack/vpc] " +
				"terraform: " +
				"Plan: 2 to add, 1 to change, " +
				"0 to destroy.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripANSI(tt.input)
			if got != tt.want {
				t.Errorf(
					"StripANSI() = %q, want %q",
					got, tt.want)
			}
		})
	}
}
