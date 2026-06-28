package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: removing an optional write-only property from a response
func TestResponseOptionalPropertyUpdatedCheck(t *testing.T) {
	s1, err := open("../data/checker/response_optional_property_removed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_optional_property_removed_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseOptionalPropertyUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponseOptionalPropertyRemovedId,
		Args:        []any{"data/id", "200"},
		Level:       checker.WARN,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_optional_property_removed_revision.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// CL: adding an optional write-only property to a response
func TestResponseOptionalPropertyAddedCheck(t *testing.T) {
	s1, err := open("../data/checker/response_optional_property_removed_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_optional_property_removed_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseOptionalPropertyUpdatedCheck), d, osm, checker.INFO)

	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponseOptionalPropertyAddedId,
		Args:        []any{"data/id", "200"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_optional_property_removed_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// CL: removing an optional write-only property from a response
func TestResponseOptionalWriteOnlyPropertyRemovedCheck(t *testing.T) {
	s1, err := open("../data/checker/response_optional_property_removed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_optional_property_removed_revision.yaml")
	require.NoError(t, err)

	s1.Spec.Paths.Value("/api/v1.0/groups").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value.Properties["data"].Value.Properties["id"].Value.WriteOnly = true
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseOptionalPropertyUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponseOptionalWriteOnlyPropertyRemovedId,
		Args:        []any{"data/id", "200"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_optional_property_removed_revision.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// CL: adding an optional write-only property to a response
func TestResponseOptionalWriteOnlyPropertyAddedCheck(t *testing.T) {
	s1, err := open("../data/checker/response_optional_property_removed_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_optional_property_removed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value.Properties["data"].Value.Properties["id"].Value.WriteOnly = true
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseOptionalPropertyUpdatedCheck), d, osm, checker.INFO)

	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponseOptionalWriteOnlyPropertyAddedId,
		Args:        []any{"data/id", "200"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_optional_property_removed_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}
