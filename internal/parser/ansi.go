package parser

import "regexp"

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// StripANSI removes all SGR escape sequences from s.
func StripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}
