package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// CL: removing unevaluated constraints from response body
func TestResponseBodyUnevaluatedRemoved(t *testing.T) {
	s1, err := open("../data/checker/unevaluated_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/unevaluated_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyUnevaluatedUpdatedCheck), d, osm, checker.INFO)
	require.True(t, containsId(errs, checker.ResponseBodyUnevaluatedItemsRemovedId))
	require.True(t, containsId(errs, checker.ResponseBodyUnevaluatedPropertiesRemovedId))
}

// CL: adding unevaluated constraints to response property
func TestResponsePropertyUnevaluatedAdded(t *testing.T) {
	s1, err := open("../data/checker/unevaluated_property_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/unevaluated_property_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyUnevaluatedUpdatedCheck), d, osm, checker.INFO)
	require.True(t, containsId(errs, checker.ResponsePropertyUnevaluatedItemsAddedId))
	require.True(t, containsId(errs, checker.ResponsePropertyUnevaluatedPropertiesAddedId))
}

// CL: removing unevaluated constraints from response property
func TestResponsePropertyUnevaluatedRemoved(t *testing.T) {
	s1, err := open("../data/checker/unevaluated_property_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/unevaluated_property_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyUnevaluatedUpdatedCheck), d, osm, checker.INFO)
	require.True(t, containsId(errs, checker.ResponsePropertyUnevaluatedItemsRemovedId))
	require.True(t, containsId(errs, checker.ResponsePropertyUnevaluatedPropertiesRemovedId))
}

// CL: adding unevaluated constraints to response body
func TestResponseBodyUnevaluatedAdded(t *testing.T) {
	s1, err := open("../data/checker/unevaluated_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/unevaluated_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyUnevaluatedUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	ids := make(map[string]bool)
	for _, e := range errs {
		ids[e.GetId()] = true
	}
	require.True(t, ids[checker.ResponseBodyUnevaluatedItemsAddedId], "expected response-body-unevaluated-items-added")
	require.True(t, ids[checker.ResponseBodyUnevaluatedPropertiesAddedId], "expected response-body-unevaluated-properties-added")
}
