package diff_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func TestConfig_Default(t *testing.T) {
	c := diff.NewConfig()
	require.False(t, c.IsExcludeExamples())
	require.False(t, c.IsExcludeDescription())
	require.False(t, c.IsExcludeEndpoints())
	require.False(t, c.IsExcludeTitle())
	require.False(t, c.IsExcludeSummary())
}

func TestConfig_ExcludeElements(t *testing.T) {
	c := diff.NewConfig(diff.WithExcludeElements(diff.GetExcludeDiffOptions()))
	require.True(t, c.IsExcludeExamples())
	require.True(t, c.IsExcludeDescription())
	require.True(t, c.IsExcludeEndpoints())
	require.True(t, c.IsExcludeTitle())
	require.True(t, c.IsExcludeSummary())
}
