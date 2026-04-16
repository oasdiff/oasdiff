package checker

import (
	"strings"

	"github.com/oasdiff/oasdiff/diff"
)

// formatDependentRequiredModification returns a human-readable description
// of the changes to a dependentRequired entry's value list.
func formatDependentRequiredModification(stringsDiff *diff.StringsDiff) string {
	parts := make([]string, 0, 2)
	if len(stringsDiff.Added) > 0 {
		parts = append(parts, strings.Join(stringsDiff.Added, ", ")+" added")
	}
	if len(stringsDiff.Deleted) > 0 {
		parts = append(parts, strings.Join(stringsDiff.Deleted, ", ")+" removed")
	}
	return strings.Join(parts, ", ")
}
