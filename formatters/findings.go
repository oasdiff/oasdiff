package formatters

import (
	"strings"

	"github.com/oasdiff/oasdiff/checker"
)

// Finding is a single spec-validation finding produced by the `validate`
// command.
//
// Comment, Operation, and Path use omitempty because doc-root findings
// (e.g. info-version-required) have no operation/path scope.
//
// Source.Line and Source.Column are populated when the loader tracks
// origins and the underlying error carries the offending element's
// location; both are 0 for doc-root fields with no per-key origin.
//
// Fingerprint is a stable 12-char identifier (see ComputeFingerprint)
// that lets a downstream tool match the same logical finding across
// base/revision spec versions.
type Finding struct {
	Id          string        `yaml:"id"                    json:"id"`
	Text        string        `yaml:"text"                  json:"text"`
	Comment     string        `yaml:"comment,omitempty"     json:"comment,omitempty"`
	Level       checker.Level `yaml:"level"                 json:"level"`
	Operation   string        `yaml:"operation,omitempty"   json:"operation,omitempty"`
	Path        string        `yaml:"path,omitempty"        json:"path,omitempty"`
	Section     string        `yaml:"section"               json:"section"`
	Source      Source        `yaml:"source"                json:"source"`
	Fingerprint string        `yaml:"fingerprint"           json:"fingerprint"`
}

// Source identifies the spec location of a finding. File is the spec
// path; Line and Column come from origin tracking and are 0 for doc-root
// findings that have no per-key origin.
type Source struct {
	File   string `yaml:"file"             json:"file"`
	Line   int    `yaml:"line,omitempty"   json:"line,omitempty"`
	Column int    `yaml:"column,omitempty" json:"column,omitempty"`
}

type Findings []Finding

// GetLevelCount returns the number of findings at each severity level,
// mirroring checker.Changes.GetLevelCount.
func (findings Findings) GetLevelCount() map[checker.Level]int {
	counts := map[checker.Level]int{}
	for _, finding := range findings {
		counts[finding.Level]++
	}
	return counts
}

// HasLevelOrHigher reports whether any finding is at least as severe as
// level, mirroring checker.Changes.HasLevelOrHigher. Used to decide the
// validate command's exit code against its --fail-on threshold.
func (findings Findings) HasLevelOrHigher(level checker.Level) bool {
	for _, finding := range findings {
		if finding.Level >= level {
			return true
		}
	}
	return false
}

// indentContinuation prefixes every non-empty continuation line of s
// with prefix. The first line is left as-is (the caller's format string
// already supplies its leading indent), and blank lines stay blank
// rather than becoming stray prefix-only lines. Trailing whitespace is
// trimmed.
func indentContinuation(s, prefix string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if i == 0 || line == "" {
			continue
		}
		lines[i] = prefix + line
	}
	return strings.TrimRight(strings.Join(lines, "\n"), " \t\n")
}
