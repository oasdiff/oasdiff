package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// CL: decreasing min in response body
func TestResponseBodyMinDecreased(t *testing.T) {
	s1, err := open("../data/checker/response_min_decreased_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_min_decreased_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyMinDecreasedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.ResponseBodyMinDecreasedId), "expected response-body-min-decreased")
}

// CL: decreasing min in response property
func TestResponsePropertyMinDecreased(t *testing.T) {
	s1, err := open("../data/checker/response_min_decreased_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_min_decreased_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyMinDecreasedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.ResponsePropertyMinDecreasedId), "expected response-property-min-decreased")
}

// CL: decreasing exclusiveMinimum in response body
func TestResponseBodyExclusiveMinDecreased(t *testing.T) {
	s1, err := open("../data/checker/response_min_decreased_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_min_decreased_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyMinDecreasedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.ResponseBodyExclusiveMinDecreasedId), "expected response-body-exclusive-min-decreased")
}

// CL: decreasing exclusiveMinimum in response property
func TestResponsePropertyExclusiveMinDecreased(t *testing.T) {
	s1, err := open("../data/checker/response_min_decreased_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_min_decreased_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyMinDecreasedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.ResponsePropertyExclusiveMinDecreasedId), "expected response-property-exclusive-min-decreased")
}
