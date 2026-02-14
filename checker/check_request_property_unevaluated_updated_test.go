package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// CL: adding unevaluated constraints to request body
func TestRequestBodyUnevaluatedAdded(t *testing.T) {
	s1, err := open("../data/checker/unevaluated_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/unevaluated_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyUnevaluatedUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	ids := make(map[string]bool)
	for _, e := range errs {
		ids[e.GetId()] = true
	}
	require.True(t, ids[checker.RequestBodyUnevaluatedItemsAddedId], "expected request-body-unevaluated-items-added")
	require.True(t, ids[checker.RequestBodyUnevaluatedPropertiesAddedId], "expected request-body-unevaluated-properties-added")
}

// CL: removing unevaluated constraints from request body
func TestRequestBodyUnevaluatedRemoved(t *testing.T) {
	s1, err := open("../data/checker/unevaluated_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/unevaluated_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyUnevaluatedUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	ids := make(map[string]bool)
	for _, e := range errs {
		ids[e.GetId()] = true
	}
	require.True(t, ids[checker.RequestBodyUnevaluatedItemsRemovedId], "expected request-body-unevaluated-items-removed")
	require.True(t, ids[checker.RequestBodyUnevaluatedPropertiesRemovedId], "expected request-body-unevaluated-properties-removed")
}
