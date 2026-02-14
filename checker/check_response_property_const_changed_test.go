package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: changing response body const value and response property const value
func TestResponsePropertyConstChanged(t *testing.T) {
	s1, err := open("../data/checker/response_property_const_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_const_changed_revision.yaml")
	require.NoError(t, err)
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyConstChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 2)
	require.ElementsMatch(t, []checker.ApiChange{{
		Id:          checker.ResponsePropertyConstChangedId,
		Args:        []any{"status", "ok", "success", "200"},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_property_const_changed_revision.yaml"),
		OperationId: "createOneGroup",
	}, {
		Id:          checker.ResponseBodyConstChangedId,
		Args:        []any{"text/plain", "NotFound", "Error", "404"},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_property_const_changed_revision.yaml"),
		OperationId: "createOneGroup",
	}}, errs)
}

// CL: adding response body const value or response property const value
func TestResponsePropertyConstAdded(t *testing.T) {
	s1, err := open("../data/checker/response_property_const_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_const_changed_base.yaml")
	require.NoError(t, err)

	s1.Spec.Paths.Value("/api/v1.0/groups").Post.Responses.Value("404").Value.Content["text/plain"].Schema.Value.Const = nil
	s1.Spec.Paths.Value("/api/v1.0/groups").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value.Properties["status"].Value.Const = nil

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyConstChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 2)
	require.ElementsMatch(t, []checker.ApiChange{{
		Id:          checker.ResponseBodyConstAddedId,
		Args:        []any{"text/plain", "NotFound", "404"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_property_const_changed_base.yaml"),
		OperationId: "createOneGroup",
	}, {
		Id:          checker.ResponsePropertyConstAddedId,
		Args:        []any{"status", "ok", "200"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_property_const_changed_base.yaml"),
		OperationId: "createOneGroup",
	}}, errs)
}

// CL: removing response body const value or response property const value
func TestResponsePropertyConstRemoved(t *testing.T) {
	s1, err := open("../data/checker/response_property_const_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_const_changed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Responses.Value("404").Value.Content["text/plain"].Schema.Value.Const = nil
	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value.Properties["status"].Value.Const = nil

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyConstChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 2)
	require.ElementsMatch(t, []checker.ApiChange{{
		Id:          checker.ResponseBodyConstRemovedId,
		Args:        []any{"text/plain", "NotFound", "404"},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_property_const_changed_base.yaml"),
		OperationId: "createOneGroup",
	}, {
		Id:          checker.ResponsePropertyConstRemovedId,
		Args:        []any{"status", "ok", "200"},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_property_const_changed_base.yaml"),
		OperationId: "createOneGroup",
	}}, errs)
}
