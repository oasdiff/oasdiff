package checker

import "github.com/oasdiff/oasdiff/diff"

// Extensions returns the OpenAPI extensions the checks read. A diff meant for
// the checks has to compare them (see diff.WithKeptExtensions), or a check
// that reads one reports nothing.
func Extensions() []string {
	return []string{
		diff.SunsetExtension,
		diff.XStabilityLevelExtension,
		diff.XExtensibleEnumExtension,
	}
}
