package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: decreasing request property maximum value
func TestRequestPropertyMaxDecreasedCheck(t *testing.T) {
	s1, err := open("../data/checker/request_property_max_decreased_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_max_decreased_base.yaml")
	require.NoError(t, err)

	max := float64(10)
	s2.Spec.Paths.Value("/pets").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["name"].Value.Max = &max

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMaxDecreasedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyMaxDecreasedId,
		Args:        []any{"name", 10.0},
		Operation:   "POST",
		Path:        "/pets",
		Source:      load.NewSource("../data/checker/request_property_max_decreased_base.yaml"),
		OperationId: "addPet",
	}, errs)
}

// CL: decreasing request read-only property maximum value
func TestRequestReadOnlyPropertyMaxDecreasedCheck(t *testing.T) {
	s1, err := open("../data/checker/request_property_max_decreased_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_max_decreased_base.yaml")
	require.NoError(t, err)

	max := float64(10)
	s2.Spec.Paths.Value("/pets").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["name"].Value.Max = &max
	s2.Spec.Paths.Value("/pets").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["name"].Value.ReadOnly = true

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMaxDecreasedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestReadOnlyPropertyMaxDecreasedId,
		Args:        []any{"name", 10.0},
		Operation:   "POST",
		Path:        "/pets",
		Source:      load.NewSource("../data/checker/request_property_max_decreased_base.yaml"),
		OperationId: "addPet",
	}, errs)
}

// CL: increasing request property maximum value
func TestRequestPropertyMaxIncreasingCheck(t *testing.T) {
	s1, err := open("../data/checker/request_property_max_decreased_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_max_decreased_base.yaml")
	require.NoError(t, err)

	max := float64(20)
	s2.Spec.Paths.Value("/pets").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["name"].Value.Max = &max

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMaxDecreasedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyMaxIncreasedId,
		Args:        []any{"name", 15.0, 20.0},
		Operation:   "POST",
		Path:        "/pets",
		Source:      load.NewSource("../data/checker/request_property_max_decreased_base.yaml"),
		OperationId: "addPet",
	}, errs)
}

// CL: increasing request body maximum value
func TestRequestBodyMaxIncreasingCheck(t *testing.T) {
	s1, err := open("../data/checker/request_property_max_decreased_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_max_decreased_base.yaml")
	require.NoError(t, err)

	max := float64(20)
	newMax := float64(25)
	s1.Spec.Paths.Value("/pets").Post.RequestBody.Value.Content["application/json"].Schema.Value.Max = &max
	s2.Spec.Paths.Value("/pets").Post.RequestBody.Value.Content["application/json"].Schema.Value.Max = &newMax

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMaxDecreasedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestBodyMaxIncreasedId,
		Args:        []any{20.0, 25.0},
		Operation:   "POST",
		Path:        "/pets",
		Source:      load.NewSource("../data/checker/request_property_max_decreased_base.yaml"),
		OperationId: "addPet",
	}, errs)
}

// CL: decreasing request body maximum value
func TestRequestBodyMaxDecreasedCheck(t *testing.T) {
	s1, err := open("../data/checker/request_property_max_decreased_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_max_decreased_base.yaml")
	require.NoError(t, err)

	max := float64(25)
	newMax := float64(20)
	s1.Spec.Paths.Value("/pets").Post.RequestBody.Value.Content["application/json"].Schema.Value.Max = &max
	s2.Spec.Paths.Value("/pets").Post.RequestBody.Value.Content["application/json"].Schema.Value.Max = &newMax

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyMaxDecreasedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestBodyMaxDecreasedId,
		Args:        []any{20.0},
		Operation:   "POST",
		Path:        "/pets",
		Source:      load.NewSource("../data/checker/request_property_max_decreased_base.yaml"),
		OperationId: "addPet",
	}, errs)
}
