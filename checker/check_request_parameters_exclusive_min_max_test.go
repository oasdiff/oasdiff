package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// CL: request parameter exclusiveMinimum increased
func TestRequestParameterExclusiveMinIncreased(t *testing.T) {
	s1, err := open("../data/checker/exclusive_min_max_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/exclusive_min_max_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterMinUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	ids := make(map[string]bool)
	for _, e := range errs {
		ids[e.GetId()] = true
	}
	require.True(t, ids[checker.RequestParameterExclusiveMinIncreasedId], "expected request-parameter-exclusive-min-increased")
}

// CL: request parameter exclusiveMaximum decreased
func TestRequestParameterExclusiveMaxDecreased(t *testing.T) {
	s1, err := open("../data/checker/exclusive_min_max_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/exclusive_min_max_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterMaxUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	ids := make(map[string]bool)
	for _, e := range errs {
		ids[e.GetId()] = true
	}
	require.True(t, ids[checker.RequestParameterExclusiveMaxDecreasedId], "expected request-parameter-exclusive-max-decreased")
}

// CL: request parameter exclusiveMinimum set
func TestRequestParameterExclusiveMinSet(t *testing.T) {
	s1, err := open("../data/checker/exclusive_min_max_set_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/exclusive_min_max_set_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterMinSetCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	ids := make(map[string]bool)
	for _, e := range errs {
		ids[e.GetId()] = true
	}
	require.True(t, ids[checker.RequestParameterExclusiveMinSetId], "expected request-parameter-exclusive-min-set")
}

// CL: request parameter exclusiveMaximum set
func TestRequestParameterExclusiveMaxSet(t *testing.T) {
	s1, err := open("../data/checker/exclusive_min_max_set_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/exclusive_min_max_set_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterMaxSetCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	ids := make(map[string]bool)
	for _, e := range errs {
		ids[e.GetId()] = true
	}
	require.True(t, ids[checker.RequestParameterExclusiveMaxSetId], "expected request-parameter-exclusive-max-set")
}
