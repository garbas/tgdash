package parser

import (
	"regexp"
	"strconv"
)

var (
	// Matches both old format: [unit/path] content
	// and new Terragrunt format:
	// 23:47:15.520 STDOUT [.terragrunt-stack/unit] terraform: content
	// Also handles tofu: prefix for OpenTofu users.
	unitPrefixRe = regexp.MustCompile(
		`^(?:\d{2}:\d{2}:\d{2}\.\d{3}\s+\w+\s+)?` +
			`\[([^\]]+)\]\s*(?:(?:terraform|tofu):\s*)?(.*)$`)
	planSummaryRe = regexp.MustCompile(
		`Plan:\s+(\d+)\s+to add,\s+(\d+)\s+to change,\s+(\d+)\s+to destroy`)
	applyResultRe = regexp.MustCompile(
		`Apply complete!\s+Resources:\s+(\d+)\s+added,\s+(\d+)\s+changed,\s+(\d+)\s+destroyed`)
	errorRe = regexp.MustCompile(
		`(?:^|\s*[│\|]\s*)Error:\s+`)
	noChangesRe = regexp.MustCompile(
		`No changes\.\s+Your infrastructure matches`)
	didNotRunRe = regexp.MustCompile(
		`\*\s+unit\s+\./\.terragrunt-stack/(\S+)\s+did not run`)
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

func DetectNoChanges(line string) bool {
	return noChangesRe.MatchString(line)
}

// DetectDidNotRun checks for "did not run due to early exit"
// messages and returns the unit name if found.
func DetectDidNotRun(line string) (string, bool) {
	m := didNotRunRe.FindStringSubmatch(line)
	if m == nil {
		return "", false
	}
	return m[1], true
}
