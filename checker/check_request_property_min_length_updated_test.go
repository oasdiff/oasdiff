package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: decreasing minLength of request property
func TestRequestPropertyMinLengthDecreased(t *testing.T) {
	s1, err := open("../data/checker/request_property_min_length_decreased_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_min_length_decreased_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["name"].Value.MinLength = uint64(2)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMinLengthUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	requireApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyMinLengthDecreasedId,
		Args:        []any{"name", uint64(3), uint64(2)},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_property_min_length_decreased_base.yaml"),
		OperationId: "addProduct",
	}, errs[0])
}

// CL: increasing minLength of request property
func TestRequestPropertyMinLengthIncreased(t *testing.T) {
	s1, err := open("../data/checker/request_property_min_length_decreased_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_min_length_decreased_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["name"].Value.MinLength = uint64(5)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMinLengthUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	requireApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyMinLengthIncreasedId,
		Args:        []any{"name", uint64(3), uint64(5)},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_property_min_length_decreased_base.yaml"),
		OperationId: "addProduct",
	}, errs[0])
}

// CL: increasing minLength of request body
func TestRequestBodyMinLengthIncreased(t *testing.T) {
	s1, err := open("../data/checker/request_property_min_length_decreased_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_min_length_decreased_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.MinLength = uint64(100)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMinLengthUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	requireApiChange(t, checker.ApiChange{
		Id:          checker.RequestBodyMinLengthIncreasedId,
		Args:        []any{uint64(10), uint64(100)},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_property_min_length_decreased_base.yaml"),
		OperationId: "addProduct",
	}, errs[0])
}

// CL: decreasing minLength of request body
func TestRequestBodyMinLengthDecreased(t *testing.T) {
	s1, err := open("../data/checker/request_property_min_length_decreased_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_min_length_decreased_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.MinLength = uint64(1)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMinLengthUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	requireApiChange(t, checker.ApiChange{
		Id:          checker.RequestBodyMinLengthDecreasedId,
		Args:        []any{uint64(10), uint64(1)},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_property_min_length_decreased_base.yaml"),
		OperationId: "addProduct",
	}, errs[0])
}
