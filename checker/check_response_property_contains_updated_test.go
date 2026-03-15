package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// CL: adding contains constraint to response body
func TestResponseBodyContainsAdded(t *testing.T) {
	s1, err := open("../data/checker/contains_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/contains_added_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyContainsUpdatedCheck), d, osm, checker.INFO)
	require.True(t, containsId(errs, checker.ResponseBodyContainsAddedId))
}

// CL: removing contains constraint from response body
func TestResponseBodyContainsRemoved(t *testing.T) {
	s1, err := open("../data/checker/contains_added_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/contains_added_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyContainsUpdatedCheck), d, osm, checker.INFO)
	require.True(t, containsId(errs, checker.ResponseBodyContainsRemovedId))
}

// CL: increasing minContains on response body
func TestResponseBodyMinContainsIncreased(t *testing.T) {
	s1, err := open("../data/checker/contains_min_max_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/contains_min_max_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyContainsUpdatedCheck), d, osm, checker.INFO)
	require.True(t, containsId(errs, checker.ResponseBodyMinContainsIncreasedId))
}

// CL: decreasing minContains on response body
func TestResponseBodyMinContainsDecreased(t *testing.T) {
	s1, err := open("../data/checker/contains_min_max_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/contains_min_max_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyContainsUpdatedCheck), d, osm, checker.INFO)
	require.True(t, containsId(errs, checker.ResponseBodyMinContainsDecreasedId))
}

// CL: increasing maxContains on response body
func TestResponseBodyMaxContainsIncreased(t *testing.T) {
	s1, err := open("../data/checker/contains_min_max_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/contains_min_max_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyContainsUpdatedCheck), d, osm, checker.INFO)
	require.True(t, containsId(errs, checker.ResponseBodyMaxContainsIncreasedId))
}

// CL: decreasing maxContains on response body
func TestResponseBodyMaxContainsDecreased(t *testing.T) {
	s1, err := open("../data/checker/contains_min_max_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/contains_min_max_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyContainsUpdatedCheck), d, osm, checker.INFO)
	require.True(t, containsId(errs, checker.ResponseBodyMaxContainsDecreasedId))
}

// CL: adding contains constraint to response property
func TestResponsePropertyContainsAdded(t *testing.T) {
	s1, err := open("../data/checker/contains_property_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/contains_property_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyContainsUpdatedCheck), d, osm, checker.INFO)
	require.True(t, containsId(errs, checker.ResponsePropertyContainsAddedId))
}

// CL: removing contains constraint from response property
func TestResponsePropertyContainsRemoved(t *testing.T) {
	s1, err := open("../data/checker/contains_property_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/contains_property_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyContainsUpdatedCheck), d, osm, checker.INFO)
	require.True(t, containsId(errs, checker.ResponsePropertyContainsRemovedId))
}

// CL: increasing minContains on response property
func TestResponsePropertyMinContainsIncreased(t *testing.T) {
	s1, err := open("../data/checker/contains_property_min_max_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/contains_property_min_max_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyContainsUpdatedCheck), d, osm, checker.INFO)
	require.True(t, containsId(errs, checker.ResponsePropertyMinContainsIncreasedId))
}

// CL: decreasing minContains on response property
func TestResponsePropertyMinContainsDecreased(t *testing.T) {
	s1, err := open("../data/checker/contains_property_min_max_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/contains_property_min_max_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyContainsUpdatedCheck), d, osm, checker.INFO)
	require.True(t, containsId(errs, checker.ResponsePropertyMinContainsDecreasedId))
}

// CL: increasing maxContains on response property
func TestResponsePropertyMaxContainsIncreased(t *testing.T) {
	s1, err := open("../data/checker/contains_property_min_max_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/contains_property_min_max_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyContainsUpdatedCheck), d, osm, checker.INFO)
	require.True(t, containsId(errs, checker.ResponsePropertyMaxContainsIncreasedId))
}

// CL: decreasing maxContains on response property
func TestResponsePropertyMaxContainsDecreased(t *testing.T) {
	s1, err := open("../data/checker/contains_property_min_max_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/contains_property_min_max_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyContainsUpdatedCheck), d, osm, checker.INFO)
	require.True(t, containsId(errs, checker.ResponsePropertyMaxContainsDecreasedId))
}
