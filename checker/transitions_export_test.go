package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

// Every transition publishes a unique name, a description, its claimed
// kinds, and its reporters; the model repository consumes this metadata.
func TestGetTransitions(t *testing.T) {
	names := map[string]bool{}
	for _, tr := range checker.GetTransitions() {
		require.NotEmpty(t, tr.Name)
		require.False(t, names[tr.Name], tr.Name)
		names[tr.Name] = true
		require.NotEmpty(t, tr.Description, tr.Name)
		require.NotEmpty(t, tr.Claims, tr.Name)
		require.NotEmpty(t, tr.ReportedBy, tr.Name)
	}
	require.Len(t, names, 5)
}
