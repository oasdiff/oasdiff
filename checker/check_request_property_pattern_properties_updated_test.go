package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// CL: adding pattern property to request body
func TestRequestBodyPatternPropertyAdded(t *testing.T) {
	s1, err := open("../data/checker/pattern_properties_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/pattern_properties_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyPatternPropertiesUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	found := false
	for _, e := range errs {
		if e.GetId() == checker.RequestBodyPatternPropertyAddedId {
			found = true
			break
		}
	}
	require.True(t, found, "expected request-body-pattern-property-added")
}

// CL: removing pattern property from request body
func TestRequestBodyPatternPropertyRemoved(t *testing.T) {
	s1, err := open("../data/checker/pattern_properties_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/pattern_properties_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyPatternPropertiesUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	found := false
	for _, e := range errs {
		if e.GetId() == checker.RequestBodyPatternPropertyRemovedId {
			found = true
			break
		}
	}
	require.True(t, found, "expected request-body-pattern-property-removed")
}
