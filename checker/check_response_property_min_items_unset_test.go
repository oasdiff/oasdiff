package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// removing minItems from a response drops a guarantee about what the server
// returns: the unset checks fire, and the decreased checks stay silent since
// there is no remaining bound to compare
func TestResponseMinItemsUnset(t *testing.T) {
	s1, err := open("../data/checker/min_items_set_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/min_items_set_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.ResponsePropertyMinItemsUnsetCheck), d, osm)

	require.Len(t, errs, 2)
	require.Equal(t, checker.ERR, requireChange(t, errs, checker.ResponseBodyMinItemsUnsetId).GetLevel())
	require.Equal(t, checker.ERR, requireChange(t, errs, checker.ResponsePropertyMinItemsUnsetId).GetLevel())

	require.Empty(t, checker.CheckBackwardCompatibility(singleCheckConfig(checker.ResponsePropertyMinItemsDecreasedCheck), d, osm))
}
