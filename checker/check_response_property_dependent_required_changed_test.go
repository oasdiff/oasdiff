package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// CL: adding dependentRequired to response body
func TestResponseBodyDependentRequiredAdded(t *testing.T) {
	s1, err := open("../data/checker/dependent_required_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/dependent_required_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyDependentRequiredChangedCheck), d, osm, checker.INFO)
	require.True(t, containsId(errs, checker.ResponseBodyDependentRequiredAddedId))
}

// CL: removing dependentRequired from response body
func TestResponseBodyDependentRequiredRemoved(t *testing.T) {
	s1, err := open("../data/checker/dependent_required_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/dependent_required_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyDependentRequiredChangedCheck), d, osm, checker.INFO)
	require.True(t, containsId(errs, checker.ResponseBodyDependentRequiredRemovedId))
}

// CL: changing dependentRequired on response body
func TestResponseBodyDependentRequiredChanged(t *testing.T) {
	s1, err := open("../data/checker/dependent_required_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/dependent_required_changed_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyDependentRequiredChangedCheck), d, osm, checker.INFO)
	require.True(t, containsId(errs, checker.ResponseBodyDependentRequiredChangedId))
}

// CL: changing dependentRequired on response property
func TestResponsePropertyDependentRequiredChanged(t *testing.T) {
	s1, err := open("../data/checker/dependent_required_property_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/dependent_required_property_changed_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyDependentRequiredChangedCheck), d, osm, checker.INFO)
	require.True(t, containsId(errs, checker.ResponsePropertyDependentRequiredChangedId))
}

// CL: adding dependentRequired to response property
func TestResponsePropertyDependentRequiredAdded(t *testing.T) {
	s1, err := open("../data/checker/dependent_required_property_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/dependent_required_property_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyDependentRequiredChangedCheck), d, osm, checker.INFO)
	require.True(t, containsId(errs, checker.ResponsePropertyDependentRequiredAddedId))
}

// CL: removing dependentRequired from response property
func TestResponsePropertyDependentRequiredRemoved(t *testing.T) {
	s1, err := open("../data/checker/dependent_required_property_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/dependent_required_property_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyDependentRequiredChangedCheck), d, osm, checker.INFO)
	require.True(t, containsId(errs, checker.ResponsePropertyDependentRequiredRemovedId))
}
