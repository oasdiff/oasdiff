package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// changing request body const value
func TestRequestBodyConstChanged(t *testing.T) {
	s1, err := open("../data/checker/request_body_const_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_body_const_changed_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyConstChangedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestBodyConstChangedId,
		Args:        []any{"text/plain", "FixedValue", "NewFixedValue"},
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_body_const_changed_revision.yaml"),
		OperationId: "createProduct",
	}, errs)
}

// changing request property const value
func TestRequestPropertyConstChanged(t *testing.T) {
	s1, err := open("../data/checker/request_body_const_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_body_const_changed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["status"].Value.Const = "inactive"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyConstChangedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyConstChangedId,
		Args:        []any{"status", "active", "inactive"},
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_body_const_changed_base.yaml"),
		OperationId: "createProduct",
	}, errs)
}

// adding request body const value or request property const value
func TestRequestBodyConstAdded(t *testing.T) {
	s1, err := open("../data/checker/request_body_const_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_body_const_changed_base.yaml")
	require.NoError(t, err)

	s1.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["text/plain"].Schema.Value.Const = nil
	s1.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["status"].Value.Const = nil

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyConstChangedCheck), d, osm, checker.INFO)
	requireApiChanges(t, []checker.ApiChange{{
		Id:          checker.RequestBodyConstAddedId,
		Args:        []any{"text/plain", "FixedValue"},
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_body_const_changed_base.yaml"),
		OperationId: "createProduct",
		Details:     "(media type: text/plain)",
	}, {
		Id:          checker.RequestPropertyConstAddedId,
		Args:        []any{"status", "active"},
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_body_const_changed_base.yaml"),
		OperationId: "createProduct",
		Details:     "(media type: application/json)",
	}}, errs)
}

// removing request body const value or request property const value
func TestRequestBodyConstRemoved(t *testing.T) {
	s1, err := open("../data/checker/request_body_const_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_body_const_changed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["text/plain"].Schema.Value.Const = nil
	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["status"].Value.Const = nil

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyConstChangedCheck), d, osm, checker.INFO)
	requireApiChanges(t, []checker.ApiChange{{
		Id:          checker.RequestBodyConstRemovedId,
		Args:        []any{"text/plain", "FixedValue"},
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_body_const_changed_base.yaml"),
		OperationId: "createProduct",
		Details:     "(media type: text/plain)",
	}, {
		Id:          checker.RequestPropertyConstRemovedId,
		Args:        []any{"status", "active"},
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_body_const_changed_base.yaml"),
		OperationId: "createProduct",
		Details:     "(media type: application/json)",
	}}, errs)
}
