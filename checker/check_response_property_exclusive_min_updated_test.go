package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: increasing exclusiveMinimum value of response property
func TestResponsePropertyExclusiveMinIncreased(t *testing.T) {
	s1, err := open("../data/checker/response_property_exclusive_min_updated_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_exclusive_min_updated_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyExclusiveMinUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.ResponsePropertyExclusiveMinIncreasedId,
		Args:        []any{"data/name", false, true, "200"},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_property_exclusive_min_updated_revision.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// CL: decreasing exclusiveMinimum value of response property
func TestResponsePropertyExclusiveMinDecreased(t *testing.T) {
	s1, err := open("../data/checker/response_property_exclusive_min_updated_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_exclusive_min_updated_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyExclusiveMinUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.ResponsePropertyExclusiveMinDecreasedId,
		Args:        []any{"data/name", true, false, "200"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_property_exclusive_min_updated_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}
