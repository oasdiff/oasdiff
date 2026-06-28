package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: setting maxLength of request parameters
func TestRequestParameterMaxLengthSetCheck(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_max_length_set_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_max_length_set_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterMaxLengthSetCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:        checker.RequestParameterMaxLengthSetId,
		Args:      []any{"query", "category", uint64(15)},
		Level:     checker.WARN,
		Comment:   checker.RequestParameterMaxLengthSetId + "-comment",
		Operation: "POST",
		Path:      "/test",
		Source:    load.NewSource("../data/checker/request_parameter_max_length_set_revision.yaml"),
	}, errs)
}
