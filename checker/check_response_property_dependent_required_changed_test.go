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
	require.NotEmpty(t, errs)

	found := false
	for _, e := range errs {
		if e.GetId() == checker.ResponseBodyDependentRequiredAddedId {
			found = true
			break
		}
	}
	require.True(t, found, "expected response-body-dependent-required-added")
}
