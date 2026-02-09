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
