package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: changing request body default value
func TestRequestBodyDefaultValueChanged(t *testing.T) {
	s1, err := open("../data/checker/request_body_default_value_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_body_default_value_changed_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyDefaultValueChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestBodyDefaultValueChangedId,
		Args:        []any{"text/plain", "Default", "NewDefault"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_body_default_value_changed_revision.yaml"),
		OperationId: "createProduct",
	}, errs[0])
}

// CL: changing request property default value
func TestRequestPropertyDefaultValueChanged(t *testing.T) {
	s1, err := open("../data/checker/request_property_default_value_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_default_value_changed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["price"].Value.Default = 20.0

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyDefaultValueChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestPropertyDefaultValueChangedId,
		Args:        []any{"price", 10.0, 20.0},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_property_default_value_changed_base.yaml"),
		OperationId: "createProduct",
	}, errs[0])
}

// CL: adding request body default value or request property default value
func TestRequestBodyDefaultValueAdded(t *testing.T) {
	s1, err := open("../data/checker/request_body_default_value_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_body_default_value_changed_base.yaml")
	require.NoError(t, err)

	s1.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["text/plain"].Schema.Value.Default = nil
	s1.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["price"].Value.Default = nil

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyDefaultValueChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 2)
	require.ElementsMatch(t, []checker.ApiChange{{
		Id:          checker.RequestBodyDefaultValueAddedId,
		Args:        []any{"text/plain", "Default"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_body_default_value_changed_base.yaml"),
		OperationId: "createProduct",
		Details:     "(media type: text/plain)",
	}, {
		Id:          checker.RequestPropertyDefaultValueAddedId,
		Args:        []any{"price", 10.0},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_body_default_value_changed_base.yaml"),
		OperationId: "createProduct",
		Details:     "(media type: application/json)",
	}}, errs)
}

// CL: removing request body default value or request property default value
func TestRequestBodyDefaultValueRemoving(t *testing.T) {
	s1, err := open("../data/checker/request_body_default_value_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_body_default_value_changed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["text/plain"].Schema.Value.Default = nil
	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["price"].Value.Default = nil

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyDefaultValueChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 2)
	require.ElementsMatch(t, []checker.ApiChange{{
		Id:          checker.RequestBodyDefaultValueRemovedId,
		Args:        []any{"text/plain", "Default"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_body_default_value_changed_base.yaml"),
		OperationId: "createProduct",
		Details:     "(media type: text/plain)",
	}, {
		Id:          checker.RequestPropertyDefaultValueRemovedId,
		Args:        []any{"price", 10.0},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_body_default_value_changed_base.yaml"),
		OperationId: "createProduct",
		Details:     "(media type: application/json)",
	}}, errs)
}
