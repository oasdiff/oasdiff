package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func closedObjectChanges(t *testing.T, base, revision string) []any {
	t.Helper()
	s1, err := open("../data/checker/closed_object_" + base + ".yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/closed_object_" + revision + ".yaml")
	require.NoError(t, err)
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)
	result := []any{}
	for _, e := range errs {
		if e.GetId() == checker.ResponsePropertyAddedToClosedObjectId {
			require.Equal(t, checker.ERR, e.GetLevel())
			result = append(result, e.(checker.ApiChange).Args[0])
		}
	}
	return result
}

// required and optional additions to closed objects are reported at any depth; the open object is not
func TestResponsePropertyAddedToClosedObject(t *testing.T) {
	require.ElementsMatch(t, []any{"b", "closed/y"}, closedObjectChanges(t, "base", "revision"))
}

// removing a property from a closed response object is handled by the removal checks
func TestResponsePropertyRemovedFromClosedObject(t *testing.T) {
	require.Empty(t, closedObjectChanges(t, "revision", "base"))
}
