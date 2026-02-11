package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

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
