package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// setting minItems on a parameter that had none fires the set check; the
// updated check reports only changes between two existing bounds
func TestRequestParameterMinItemsSet(t *testing.T) {
	s1, err := open("../data/checker/min_items_set_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/min_items_set_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestParameterMinItemsSetCheck), d, osm)
	change := requireSingleChange(t, errs, checker.RequestParameterMinItemsSetId)
	require.Equal(t, checker.ERR, change.GetLevel())
	require.NotEmpty(t, change.GetComment(checker.NewDefaultLocalizer()))

	require.Empty(t, checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestParameterMinItemsUpdatedCheck), d, osm))
}
