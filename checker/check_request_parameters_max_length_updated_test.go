package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: increasing maxLength of request parameters
func TestRequestParameterMaxLengthIncreasedCheck(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_max_length_updated_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_max_length_updated_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterMaxLengthUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	requireApiChange(t, checker.ApiChange{
		Id:        checker.RequestParameterMaxLengthIncreasedId,
		Args:      []any{"query", "category", uint64(10), uint64(15)},
		Level:     checker.INFO,
		Operation: "POST",
		Path:      "/test",
		Source:    load.NewSource("../data/checker/request_parameter_max_length_updated_revision.yaml"),
	}, errs[0])
}

// CL: decreasing maxLength of request parameters
func TestRequestParameterMaxLengthDecreasedCheck(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_max_length_updated_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_max_length_updated_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterMaxLengthUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	requireApiChange(t, checker.ApiChange{
		Id:        checker.RequestParameterMaxLengthDecreasedId,
		Args:      []any{"query", "category", uint64(15), uint64(10)},
		Level:     checker.ERR,
		Operation: "POST",
		Path:      "/test",
		Source:    load.NewSource("../data/checker/request_parameter_max_length_updated_base.yaml"),
	}, errs[0])
}
