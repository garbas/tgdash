package parser

import "strings"

func ExtractUnitPrefix(
	raw string,
) (unit string, line string) {
	m := unitPrefixRe.FindStringSubmatch(raw)
	if m == nil {
		return "", raw
	}
	unit = m[1]
	// Strip .terragrunt-stack/ prefix from unit paths.
	unit = strings.TrimPrefix(unit, ".terragrunt-stack/")
	return unit, strings.TrimRight(m[2], " ")
}
