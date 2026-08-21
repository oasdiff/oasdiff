package utils

import "strings"

// ToKebabCase converts a camelCase name to kebab-case: maxItems becomes
// max-items. Characters other than ASCII uppercase pass through unchanged.
func ToKebabCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(r - 'A' + 'a')
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
