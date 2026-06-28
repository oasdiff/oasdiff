package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: increasing maximum value of request parameter
func TestRequestParameterMaxIncreased(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_max_updated_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_max_updated_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterMaxUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	requireApiChange(t, checker.ApiChange{
		Id:          checker.RequestParameterMaxIncreasedId,
		Args:        []any{"query", "category", 5.0, 10.0},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_parameter_max_updated_revision.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// CL: decreasing maximum value of request parameter
func TestRequestParameterMaxDecreased(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_max_updated_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_max_updated_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterMaxUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	requireApiChange(t, checker.ApiChange{
		Id:          checker.RequestParameterMaxDecreasedId,
		Args:        []any{"query", "category", 10.0, 5.0},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_parameter_max_updated_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}
