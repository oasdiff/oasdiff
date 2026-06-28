package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: increasing minimum value of request parameter
func TestRequestParameterMinIncreased(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_min_increased_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_min_increased_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterMinUpdatedCheck), d, osm, checker.ERR)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestParameterMinIncreasedId,
		Args:        []any{"path", "groupId", 1.0, 10.0},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_parameter_min_increased_revision.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// CL: decreasing minimum value of request parameter
func TestRequestParameterMinDecreased(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_min_increased_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_min_increased_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterMinUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestParameterMinDecreasedId,
		Args:        []any{"path", "groupId", 10.0, 1.0},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_parameter_min_increased_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}
