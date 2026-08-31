package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// minItems is a value keyword whose zero means no constraint, so the diff
// reports its absence as 0. Setting it where it was absent fires the set
// checks, not the increased ones
func TestRequestMinItemsSet(t *testing.T) {
	s1, err := open("../data/checker/min_items_set_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/min_items_set_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestPropertyMinItemsSetCheck), d, osm)

	require.Len(t, errs, 2)
	body := requireChange(t, errs, checker.RequestBodyMinItemsSetId)
	require.Equal(t, checker.ERR, body.GetLevel())
	require.NotEmpty(t, body.GetComment(checker.NewDefaultLocalizer()))
	prop := requireChange(t, errs, checker.RequestPropertyMinItemsSetId)
	require.Equal(t, checker.ERR, prop.GetLevel())
}

// tightening an existing bound stays with the increased checks: the set checks
// only cover the transition from no constraint
func TestRequestMinItemsIncreasedNotSet(t *testing.T) {
	s1, err := open("../data/checker/min_items_set_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/min_items_set_revision.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/batch").Post.RequestBody.Value.Content["application/json"].Schema.Value.MinItems = 5

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	require.Empty(t, checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestPropertyMinItemsSetCheck), d, osm))
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestPropertyMinItemsIncreasedCheck), d, osm)
	requireSingleChange(t, errs, checker.RequestBodyMinItemsIncreasedId)
}
