package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// CL: adding pattern property to response body
func TestResponseBodyPatternPropertyAdded(t *testing.T) {
	s1, err := open("../data/checker/pattern_properties_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/pattern_properties_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyPatternPropertiesUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	found := false
	for _, e := range errs {
		if e.GetId() == checker.ResponseBodyPatternPropertyAddedId {
			found = true
			break
		}
	}
	require.True(t, found, "expected response-body-pattern-property-added")
}

// CL: removing pattern property from response body
func TestResponseBodyPatternPropertyRemoved(t *testing.T) {
	s1, err := open("../data/checker/pattern_properties_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/pattern_properties_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyPatternPropertiesUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.ResponseBodyPatternPropertyRemovedId), "expected response-body-pattern-property-removed")
}

// CL: adding pattern property to response property
func TestResponsePropertyPatternPropertyAdded(t *testing.T) {
	s1, err := open("../data/checker/pattern_properties_prop_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/pattern_properties_prop_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyPatternPropertiesUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.ResponsePropertyPatternPropertyAddedId), "expected response-property-pattern-property-added")
}

// CL: removing pattern property from response property
func TestResponsePropertyPatternPropertyRemoved(t *testing.T) {
	s1, err := open("../data/checker/pattern_properties_prop_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/pattern_properties_prop_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyPatternPropertiesUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.ResponsePropertyPatternPropertyRemovedId), "expected response-property-pattern-property-removed")
}
