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

// The revision is where the change is reported, so its location is the one the
// change answers with.
func TestInfoChange_SourceIsTheRevision(t *testing.T) {
	change := infoChange(
		&checker.Source{File: "base.yaml", Line: 4, Column: 3},
		&checker.Source{File: "revision.yaml", Line: 9, Column: 5, EndLine: 9, EndColumn: 20},
	)

	require.Equal(t, "info\t[api-version-changed] at revision.yaml\n\tin info\n\t\tthe API version changed from `1.0.0` to `2.0.0`",
		change.MultiLineError(checker.NewDefaultLocalizer(), checker.ColorNever))
}

// A change only the base has a location for is reported against the base.
func TestInfoChange_FallsBackToTheBase(t *testing.T) {
	change := infoChange(&checker.Source{File: "base.yaml", Line: 4, Column: 3}, nil)

	require.Equal(t, "info\t[api-version-changed] at base.yaml\n\tin info\n\t\tthe API version changed from `1.0.0` to `2.0.0`",
		change.MultiLineError(checker.NewDefaultLocalizer(), checker.ColorNever))
}

// Without origin tracking there is no location, and the error lines carry no
// "at" segment.
func TestInfoChange_WithoutSources(t *testing.T) {
	change := infoChange(nil, nil)
	l := checker.NewDefaultLocalizer()

	require.Equal(t, "info\t[api-version-changed]\n\tin info\n\t\tthe API version changed from `1.0.0` to `2.0.0`",
		change.MultiLineError(l, checker.ColorNever))
	require.Equal(t, "info, in info the API version changed from `1.0.0` to `2.0.0` [api-version-changed]. ",
		change.SingleLineError(l, checker.ColorNever))
}

func TestInfoChange_ErrorLinesCarryTheSource(t *testing.T) {
	change := infoChange(nil, &checker.Source{File: "revision.yaml", Line: 9, Column: 5})
	l := checker.NewDefaultLocalizer()

	require.Equal(t, "info\t[api-version-changed] at revision.yaml\n\tin info\n\t\tthe API version changed from `1.0.0` to `2.0.0`",
		change.MultiLineError(l, checker.ColorNever))
	require.Equal(t, "info at revision.yaml, in info the API version changed from `1.0.0` to `2.0.0` [api-version-changed]. ",
		change.SingleLineError(l, checker.ColorNever))
}
