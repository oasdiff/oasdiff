package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// #916: the property-diff traversal helpers must descend into the `not`
// sub-schema, like they do for if/then/else and the other sub-schema fields.
// Before the fix, a change inside `not` was populated in NotDiff but never
// reached the visitor, so every check built on those helpers silently missed
// it.
//
// Exercised through the public CheckBackwardCompatibility rather than the
// traversal helpers directly, so the test does not depend on those helpers
// staying exported (they are slated to be privatized; see #953). Every property
// in the fixtures lives under `not`, so each of these findings firing proves
// the traversal reached inside the `not` sub-schema (a deleted, an added, and a
// modified property respectively); without the fix none of them fire.
func TestBackwardCompatibility_TraversesNotSubSchema(t *testing.T) {
	s1, err := open("../data/checker/not_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/not_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)

	require.True(t, containsId(errs, checker.RequestPropertyRemovedId), "deleted property under `not` should be detected")
	require.True(t, containsId(errs, checker.NewRequiredRequestPropertyId), "added required property under `not` should be detected")
	require.True(t, containsId(errs, checker.RequestPropertyMaxLengthDecreasedId), "modified property under `not` should be detected")

	// Each finding should carry the `not/` path, confirming it came from inside
	// the `not` sub-schema and not a top-level property.
	l := checker.NewDefaultLocalizer()
	for _, e := range errs {
		require.Contains(t, e.GetUncolorizedText(l), "/not/")
	}
}
