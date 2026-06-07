package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// Property changes nested inside a `not` sub-schema are detected by the
// backward-compatibility checks, the same as changes under if/then/else and the
// other sub-schemas. The fixtures place a removed, an added-required, and a
// modified property under `not` and nothing at the top level, so each finding
// firing confirms the check descended into `not`; every finding also carries
// the `not/` path.
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
