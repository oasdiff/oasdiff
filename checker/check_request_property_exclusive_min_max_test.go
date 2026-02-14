package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// CL: request body/property exclusiveMinimum increased, exclusiveMaximum decreased
func TestRequestPropertyExclusiveMinIncreased(t *testing.T) {
	s1, err := open("../data/checker/exclusive_min_max_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/exclusive_min_max_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMinIncreasedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	ids := make(map[string]bool)
	for _, e := range errs {
		ids[e.GetId()] = true
	}
	require.True(t, ids[checker.RequestBodyExclusiveMinIncreasedId], "expected request-body-exclusive-min-increased")
	require.True(t, ids[checker.RequestPropertyExclusiveMinIncreasedId], "expected request-property-exclusive-min-increased")
}

// CL: request body/property exclusiveMaximum decreased
func TestRequestPropertyExclusiveMaxDecreased(t *testing.T) {
	s1, err := open("../data/checker/exclusive_min_max_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/exclusive_min_max_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMaxDecreasedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	ids := make(map[string]bool)
	for _, e := range errs {
		ids[e.GetId()] = true
	}
	require.True(t, ids[checker.RequestBodyExclusiveMaxDecreasedId], "expected request-body-exclusive-max-decreased")
	require.True(t, ids[checker.RequestPropertyExclusiveMaxDecreasedId], "expected request-property-exclusive-max-decreased")
}

// CL: request body/property exclusiveMinimum set
func TestRequestPropertyExclusiveMinSet(t *testing.T) {
	s1, err := open("../data/checker/exclusive_min_max_set_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/exclusive_min_max_set_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMinSetCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	ids := make(map[string]bool)
	for _, e := range errs {
		ids[e.GetId()] = true
	}
	require.True(t, ids[checker.RequestBodyExclusiveMinSetId], "expected request-body-exclusive-min-set")
	require.True(t, ids[checker.RequestPropertyExclusiveMinSetId], "expected request-property-exclusive-min-set")
}

// CL: request body/property exclusiveMaximum set
func TestRequestPropertyExclusiveMaxSet(t *testing.T) {
	s1, err := open("../data/checker/exclusive_min_max_set_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/exclusive_min_max_set_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMaxSetCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	ids := make(map[string]bool)
	for _, e := range errs {
		ids[e.GetId()] = true
	}
	require.True(t, ids[checker.RequestBodyExclusiveMaxSetId], "expected request-body-exclusive-max-set")
	require.True(t, ids[checker.RequestPropertyExclusiveMaxSetId], "expected request-property-exclusive-max-set")
}
