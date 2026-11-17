package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// nullabilityLocations returns the base and revision line and column of the
// request-property nullability change for the named property.
func nullabilityLocations(t *testing.T, base, revision, property string) ([2]int, [2]int) {
	t.Helper()
	s1, err := open(base)
	require.NoError(t, err)
	s2, err := open(revision)
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	changes := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyBecameNotNullableCheck), d, osm, checker.INFO)
	for _, change := range changes {
		apiChange := change.(checker.ApiChange)
		if apiChange.Args[0] != property {
			continue
		}
		require.NotNil(t, apiChange.BaseSource)
		require.NotNil(t, apiChange.RevisionSource)
		require.Equal(t, base, apiChange.BaseSource.File)
		require.Equal(t, revision, apiChange.RevisionSource.File)
		return [2]int{apiChange.BaseSource.Line, apiChange.BaseSource.Column},
			[2]int{apiChange.RevisionSource.Line, apiChange.RevisionSource.Column}
	}
	require.Failf(t, "no nullability change", "property %q", property)
	return [2]int{}, [2]int{}
}

// The nullable keyword is located where it is written; the side without it is
// located at the property.
func TestNullabilitySources_NullableKeyword(t *testing.T) {
	base, revision := nullabilityLocations(t,
		"../data/checker/request_property_became_nullable_base.yaml",
		"../data/checker/request_property_became_nullable_revision.yaml",
		"name")
	require.Equal(t, [2]int{15, 17}, base)
	require.Equal(t, [2]int{17, 19}, revision)
}

// A "null" entry in the type array is located at the entry; the side without
// one is located at its type.
func TestNullabilitySources_TypeArray(t *testing.T) {
	base, revision := nullabilityLocations(t,
		"../data/checker/request_property_became_nullable_31_base.yaml",
		"../data/checker/request_property_became_nullable_31_revision.yaml",
		"name")
	require.Equal(t, [2]int{17, 19}, base)
	require.Equal(t, [2]int{19, 23}, revision)
}

// The nullable oneOf wrapping is located at the oneOf; the unwrapped side is
// located at the type of the schema that got wrapped.
func TestNullabilitySources_Wrapping(t *testing.T) {
	base, revision := nullabilityLocations(t,
		"../data/checker/nullable_wrap_base.yaml",
		"../data/checker/nullable_wrap_revision.yaml",
		"optionalRef")
	require.Equal(t, [2]int{21, 7}, base)
	require.Equal(t, [2]int{34, 11}, revision)
}

// A oneOf whose branches decide is located at the branch that was edited, not
// at the property holding the oneOf.
func TestNullabilitySources_ComposedBranch(t *testing.T) {
	base, revision := nullabilityLocations(t, oneOfNullBranchBase, oneOfNullBranchRevision, "address")
	require.Equal(t, [2]int{35, 7}, base)
	require.Equal(t, [2]int{35, 22}, revision)
}

// When the composition's nullability moved without any one branch changing
// its own, the oneOf itself is located.
func TestNullabilitySources_ComposedWithoutChangedBranch(t *testing.T) {
	s1, s2 := oneOfNullBranchSpecs(t)
	// Both sides get the nullable branch, and the base's first branch is an
	// enum where the revision's is null: the revision has a second branch
	// accepting null, but no branch that was typed and gained "null".
	for _, schema := range addressSchemas(s1) {
		schema.OneOf[1] = addressSchemas(s2)[0].OneOf[1]
		schema.OneOf[0].Value.Type = nil
		schema.OneOf[0].Value.Enum = []any{"none"}
	}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	changes := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyBecameNotNullableCheck), d, osm, checker.INFO)
	require.Len(t, changes, 1)
	apiChange := changes[0].(checker.ApiChange)
	require.Equal(t, 17, apiChange.BaseSource.Line)
	require.Equal(t, 17, apiChange.RevisionSource.Line)
}
