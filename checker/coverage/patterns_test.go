package coverage_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker/coverage"
	"github.com/stretchr/testify/require"
)

// every pattern accounts for at least one edit: an entry that matches
// nothing is stale, and a general pattern listed before a specific one
// starves it
func TestPatterns(t *testing.T) {
	patterns := coverage.Patterns()
	require.NotEmpty(t, patterns)

	for _, pattern := range patterns {
		require.Contains(t, []string{"waiver", "non-contract"}, pattern.Kind)
		require.NotEmpty(t, pattern.Reason, pattern.Pattern)
		require.Positive(t, pattern.Edits, "pattern %q accounts for no edit", pattern.Pattern)
	}
}
