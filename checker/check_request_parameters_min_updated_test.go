package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// increasing minimum value of request parameter
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
		OperationId: "createOneGroup",
	}, errs)
}

// decreasing minimum value of request parameter
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
		OperationId: "createOneGroup",
	}, errs)
}
