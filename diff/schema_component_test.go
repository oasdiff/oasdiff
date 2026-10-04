package diff_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func componentNameDiff(t *testing.T, base, revision string) *diff.SchemaDiff {
	t.Helper()
	loader := openapi3.NewLoader()
	s1, err := loader.LoadFromFile(base)
	require.NoError(t, err)
	s2, err := loader.LoadFromFile(revision)
	require.NoError(t, err)
	d, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	return d.PathsDiff.Modified["/example"].OperationsDiff.Modified["GET"].
		ResponsesDiff.Modified["200"].ContentDiff.MediaTypeModified["application/json"].SchemaDiff
}

// A schema reached through a property is named on each side; an inline schema
// is not.
func TestSchemaDiff_ComponentName(t *testing.T) {
	schemaDiff := componentNameDiff(t, "../data/component-name1.yaml", "../data/component-name2.yaml")

	shared := schemaDiff.PropertiesDiff.Modified["left"]
	require.Equal(t, "Shared", shared.BaseComponent)
	require.Equal(t, "Shared", shared.RevisionComponent)

	inline := schemaDiff.PropertiesDiff.Modified["inline"]
	require.Empty(t, inline.BaseComponent)
	require.Empty(t, inline.RevisionComponent)
}

// A renamed component names itself on each side.
func TestSchemaDiff_ComponentNameOfRenamedComponent(t *testing.T) {
	schemaDiff := componentNameDiff(t, "../data/component-renamed1.yaml", "../data/component-renamed2.yaml")

	require.Equal(t, "Old", schemaDiff.BaseComponent)
	require.Equal(t, "New", schemaDiff.RevisionComponent)
}

// The names are context, not a change.
func TestSchemaDiff_ComponentNameIsNotAChange(t *testing.T) {
	schemaDiff := componentNameDiff(t, "../data/component-name1.yaml", "../data/component-name2.yaml")
	require.NotContains(t, schemaDiff.PropertiesDiff.Modified, "untouched")

	loader := openapi3.NewLoader()
	s1, err := loader.LoadFromFile("../data/component-name1.yaml")
	require.NoError(t, err)
	s2, err := loader.LoadFromFile("../data/component-name1.yaml")
	require.NoError(t, err)
	d, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.True(t, d.Empty())
}

// In OpenAPI 3.1 a description beside a $ref overrides the component's, which
// the parser applies to a copy of the component, so the copy is named from
// the $ref. A $ref to a schema inside a component is not named after it.
func TestSchemaDiff_ComponentNameOfRefWithOverride(t *testing.T) {
	schemaDiff := componentNameDiff(t, "../data/component-ref-override1.yaml", "../data/component-ref-override2.yaml")

	left := schemaDiff.PropertiesDiff.Modified["left"]
	require.Equal(t, "Shared", left.BaseComponent)
	require.Equal(t, "Shared", left.RevisionComponent)

	nested := schemaDiff.PropertiesDiff.Modified["nested"]
	require.Empty(t, nested.BaseComponent)
	require.Empty(t, nested.RevisionComponent)
}
