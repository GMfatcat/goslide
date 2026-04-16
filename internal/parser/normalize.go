package parser

import "strings"

func normalizeEnum(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
