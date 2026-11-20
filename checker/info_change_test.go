package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

func infoChange(baseSource, revisionSource *checker.Source) checker.InfoChange {
	return checker.InfoChange{
		Id:    "api-version-changed",
		Args:  []any{"1.0.0", "2.0.0"},
		Level: checker.INFO,
	}.WithSources(baseSource, revisionSource)
}

// A change in the info object is about that section as a whole, so it names no
// element within it.
func TestInfoChange_Subject(t *testing.T) {
	section, element := infoChange(nil, nil).Subject()
	require.Equal(t, "info", section)
	require.Empty(t, element)
	require.Equal(t, "info", infoChange(nil, nil).GetSection())
}

// WithSources keeps each side's location, and either side may be absent.
func TestInfoChange_WithSources(t *testing.T) {
	base := &checker.Source{File: "base.yaml", Line: 4, Column: 3}
	revision := &checker.Source{File: "revision.yaml", Line: 9, Column: 5, EndLine: 9, EndColumn: 20}

	both := infoChange(base, revision)
	require.Equal(t, base, both.GetBaseSource())
	require.Equal(t, revision, both.GetRevisionSource())

	baseOnly := infoChange(base, nil)
	require.Equal(t, base, baseOnly.GetBaseSource())
	require.Nil(t, baseOnly.GetRevisionSource())

	none := infoChange(nil, nil)
	require.Nil(t, none.GetBaseSource())
	require.Nil(t, none.GetRevisionSource())
}

// The message reads the same whatever the locations are.
func TestInfoChange_Text(t *testing.T) {
	require.Equal(t,
		"the API version changed from `1.0.0` to `2.0.0`",
		infoChange(nil, nil).GetUncolorizedText(checker.NewDefaultLocalizer()))
}
