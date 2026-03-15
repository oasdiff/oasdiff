package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// CL: response body/property exclusiveMaximum increased
func TestResponsePropertyExclusiveMaxIncreased(t *testing.T) {
	s1, err := open("../data/checker/exclusive_min_max_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/exclusive_min_max_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyMaxIncreasedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	ids := make(map[string]bool)
	for _, e := range errs {
		ids[e.GetId()] = true
	}
	require.True(t, ids[checker.ResponseBodyExclusiveMaxIncreasedId], "expected response-body-exclusive-max-increased")
	require.True(t, ids[checker.ResponsePropertyExclusiveMaxIncreasedId], "expected response-property-exclusive-max-increased")
}
