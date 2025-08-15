package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

func TestApiChange_SourceFile(t *testing.T) {
	apiChangeSourceFile := apiChange
	apiChangeSourceFile.SourceFile = ""
	apiChangeSourceFile.Source = load.NewSource("spec.yaml")

	require.Equal(t, "spec.yaml", apiChangeSourceFile.GetSourceFile())
}

func TestApiChange_SourceUrl(t *testing.T) {
	apiChangeSourceFile := apiChange
	apiChangeSourceFile.SourceFile = ""
	apiChangeSourceFile.Source = load.NewSource("http://google.com/spec.yaml")

	require.Equal(t, "", apiChangeSourceFile.GetSourceFile())
}

func TestApiChangeWithSources_DirectConstruction(t *testing.T) {
	// Test direct construction of ApiChange with BaseSource and RevisionSource
	baseSource := checker.NewSource("base.yaml", 10, 5)
	revisionSource := checker.NewSource("revision.yaml", 12, 7)

	change := checker.ApiChange{
		Id:        "test-id",
		Args:      []any{"arg1"},
		Comment:   "test comment",
		Level:     checker.INFO,
		Operation: "GET",
		Path:      "/test",
		CommonChange: checker.CommonChange{
			BaseSource:     baseSource,
			RevisionSource: revisionSource,
		},
	}

	// Test that the new fields are set correctly
	require.Equal(t, baseSource, change.GetBaseSource())
	require.Equal(t, revisionSource, change.GetRevisionSource())
	require.NotEmpty(t, change.GetBaseSource())
	require.NotEmpty(t, change.GetRevisionSource())
	require.Equal(t, "base.yaml", change.GetBaseSource().File)
	require.Equal(t, 10, change.GetBaseSource().Line)
	require.Equal(t, 5, change.GetBaseSource().Column)
	require.Equal(t, "revision.yaml", change.GetRevisionSource().File)
	require.Equal(t, 12, change.GetRevisionSource().Line)
	require.Equal(t, 7, change.GetRevisionSource().Column)
}
