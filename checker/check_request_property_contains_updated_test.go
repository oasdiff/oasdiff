package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// CL: adding contains constraint to request body
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

// CL: removing contains constraint from request body
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
