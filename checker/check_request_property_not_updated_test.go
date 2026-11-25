package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// adding not to request body
func TestRequestBodyNotAdded(t *testing.T) {
	s1, err := open("../data/checker/not_subschema_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/not_subschema_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyNotUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	found := false
	for _, e := range errs {
		if e.GetId() == checker.RequestBodyNotAddedId {
			found = true
			break
		}
	}
	require.True(t, found, "expected request-body-not-added")
}

// removing not from request body
func TestRequestBodyNotRemoved(t *testing.T) {
	s1, err := open("../data/checker/not_subschema_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/not_subschema_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyNotUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	found := false
	for _, e := range errs {
		if e.GetId() == checker.RequestBodyNotRemovedId {
			found = true
			break
		}
	}
	require.True(t, found, "expected request-body-not-removed")
}

// adding not to request property
func TestRequestPropertyNotAdded(t *testing.T) {
	s1, err := open("../data/checker/not_subschema_prop_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/not_subschema_prop_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyNotUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.RequestPropertyNotAddedId), "expected request-property-not-added")
}

// removing not from request property
func TestRequestPropertyNotRemoved(t *testing.T) {
	s1, err := open("../data/checker/not_subschema_prop_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/not_subschema_prop_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyNotUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.RequestPropertyNotRemovedId), "expected request-property-not-removed")
}
