package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// adding contains constraint to request body
func TestRequestBodyContainsAdded(t *testing.T) {
	s1, err := open("../data/checker/contains_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/contains_added_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyContainsUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	found := false
	for _, e := range errs {
		if e.GetId() == checker.RequestBodyContainsAddedId {
			found = true
			break
		}
	}
	require.True(t, found, "expected request-body-contains-added")
}

// removing contains constraint from request body
func TestRequestBodyContainsRemoved(t *testing.T) {
	s1, err := open("../data/checker/contains_added_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/contains_added_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyContainsUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	found := false
	for _, e := range errs {
		if e.GetId() == checker.RequestBodyContainsRemovedId {
			found = true
			break
		}
	}
	require.True(t, found, "expected request-body-contains-removed")
}

// increasing minContains on request body
func TestRequestBodyMinContainsIncreased(t *testing.T) {
	s1, err := open("../data/checker/contains_min_max_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/contains_min_max_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyContainsUpdatedCheck), d, osm, checker.INFO)
	requireChange(t, errs, checker.RequestBodyMinContainsIncreasedId)
}

// decreasing minContains on request body
func TestRequestBodyMinContainsDecreased(t *testing.T) {
	s1, err := open("../data/checker/contains_min_max_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/contains_min_max_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyContainsUpdatedCheck), d, osm, checker.INFO)
	requireChange(t, errs, checker.RequestBodyMinContainsDecreasedId)
}

// increasing maxContains on request body
func TestRequestBodyMaxContainsIncreased(t *testing.T) {
	s1, err := open("../data/checker/contains_min_max_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/contains_min_max_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyContainsUpdatedCheck), d, osm, checker.INFO)
	requireChange(t, errs, checker.RequestBodyMaxContainsIncreasedId)
}

// decreasing maxContains on request body
func TestRequestBodyMaxContainsDecreased(t *testing.T) {
	s1, err := open("../data/checker/contains_min_max_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/contains_min_max_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyContainsUpdatedCheck), d, osm, checker.INFO)
	requireChange(t, errs, checker.RequestBodyMaxContainsDecreasedId)
}

// adding contains constraint to request property
func TestRequestPropertyContainsAdded(t *testing.T) {
	s1, err := open("../data/checker/contains_property_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/contains_property_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyContainsUpdatedCheck), d, osm, checker.INFO)
	requireChange(t, errs, checker.RequestPropertyContainsAddedId)
}

// removing contains constraint from request property
func TestRequestPropertyContainsRemoved(t *testing.T) {
	s1, err := open("../data/checker/contains_property_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/contains_property_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyContainsUpdatedCheck), d, osm, checker.INFO)
	requireChange(t, errs, checker.RequestPropertyContainsRemovedId)
}

// increasing minContains on request property
func TestRequestPropertyMinContainsIncreased(t *testing.T) {
	s1, err := open("../data/checker/contains_property_min_max_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/contains_property_min_max_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyContainsUpdatedCheck), d, osm, checker.INFO)
	requireChange(t, errs, checker.RequestPropertyMinContainsIncreasedId)
}

// decreasing minContains on request property
func TestRequestPropertyMinContainsDecreased(t *testing.T) {
	s1, err := open("../data/checker/contains_property_min_max_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/contains_property_min_max_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyContainsUpdatedCheck), d, osm, checker.INFO)
	requireChange(t, errs, checker.RequestPropertyMinContainsDecreasedId)
}

// increasing maxContains on request property
func TestRequestPropertyMaxContainsIncreased(t *testing.T) {
	s1, err := open("../data/checker/contains_property_min_max_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/contains_property_min_max_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyContainsUpdatedCheck), d, osm, checker.INFO)
	requireChange(t, errs, checker.RequestPropertyMaxContainsIncreasedId)
}

// decreasing maxContains on request property
func TestRequestPropertyMaxContainsDecreased(t *testing.T) {
	s1, err := open("../data/checker/contains_property_min_max_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/contains_property_min_max_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyContainsUpdatedCheck), d, osm, checker.INFO)
	requireChange(t, errs, checker.RequestPropertyMaxContainsDecreasedId)
}
