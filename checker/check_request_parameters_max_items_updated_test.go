package checker_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: increasing maxItems of request parameters
func TestRequestParameterMaxItemsIncreased(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_max_items_updated_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_max_items_updated_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterMaxItemsUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestParameterMaxItemsIncreasedId,
		Args:        []any{"query", "category", uint64(10), uint64(20)},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_parameter_max_items_updated_revision.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// CL: decreasing maxItems of request parameters
func TestRequestParameterMaxItemsDecreased(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_max_items_updated_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_max_items_updated_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterMaxItemsUpdatedCheck), d, osm, checker.ERR)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestParameterMaxItemsDecreasedId,
		Args:        []any{"query", "category", uint64(20), uint64(10)},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_parameter_max_items_updated_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// BC: decreasing maxItems of common request parameters without --flatten-params is not breaking
func TestBreaking_RequestParameterMaxItemsWithoutFlatten(t *testing.T) {

	s1, err := open("../data/checker/common_request_parameter_max_items_updated_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/common_request_parameter_max_items_updated_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterMaxItemsUpdatedCheck), d, osm, checker.ERR)
	require.Empty(t, errs)
}

// BC: decreasing maxItems of common request parameters with --flatten-params is breaking
func TestBreaking_RequestParameterMaxItemsWithFlatten(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := load.NewSpecInfo(loader, load.NewSource("../data/checker/common_request_parameter_max_items_updated_revision.yaml"), load.WithFlattenParams())
	require.NoError(t, err)

	s2, err := load.NewSpecInfo(loader, load.NewSource("../data/checker/common_request_parameter_max_items_updated_base.yaml"), load.WithFlattenParams())
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterMaxItemsUpdatedCheck), d, osm, checker.ERR)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestParameterMaxItemsDecreasedId,
		Args:        []any{"query", "category", uint64(20), uint64(10)},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/common_request_parameter_max_items_updated_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// BC: decreasing maxItems on array parameter schema itself (issue #760)
func TestRequestParameterArrayMaxItemsDecreased(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_array_max_items_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_array_max_items_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterMaxItemsUpdatedCheck), d, osm, checker.ERR)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestParameterMaxItemsDecreasedId,
		Args:        []any{"query", "ids", uint64(50), uint64(10)},
		Level:       checker.ERR,
		Operation:   "GET",
		Path:        "/test",
		Source:      load.NewSource("../data/checker/request_parameter_array_max_items_revision.yaml"),
		OperationId: "testOperation",
	}, errs[0])
}
