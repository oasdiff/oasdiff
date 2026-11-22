package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// adding not to response body
func TestResponseBodyNotAdded(t *testing.T) {
	s1, err := open("../data/checker/not_subschema_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/not_subschema_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyNotUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	found := false
	for _, e := range errs {
		if e.GetId() == checker.ResponseBodyNotAddedId {
			found = true
			break
		}
	}
	require.True(t, found, "expected response-body-not-added")
}

// removing not from response body
func TestResponseBodyNotRemoved(t *testing.T) {
	s1, err := open("../data/checker/not_subschema_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/not_subschema_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyNotUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.ResponseBodyNotRemovedId), "expected response-body-not-removed")
}

// adding not to response property
func TestResponsePropertyNotAdded(t *testing.T) {
	s1, err := open("../data/checker/not_subschema_prop_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/not_subschema_prop_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyNotUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.ResponsePropertyNotAddedId), "expected response-property-not-added")
}

// removing not from response property
func TestResponsePropertyNotRemoved(t *testing.T) {
	s1, err := open("../data/checker/not_subschema_prop_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/not_subschema_prop_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyNotUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.ResponsePropertyNotRemovedId), "expected response-property-not-removed")
}
