package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: changing response body property default value
func TestResponsePropertyDefaultValueUpdatedCheck(t *testing.T) {
	s1, err := open("../data/checker/response_property_default_value_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_default_value_changed_revision.yaml")
	require.NoError(t, err)
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyDefaultValueChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 2)
	requireApiChanges(t, []checker.ApiChange{{
		Id:          checker.ResponsePropertyDefaultValueChangedId,
		Args:        []any{"created", "2020-01-01T00:00:00Z", "2020-02-01T00:00:00Z", "200"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_property_default_value_changed_revision.yaml"),
		OperationId: "createOneGroup",
	}, {
		Id:          checker.ResponsePropertyDefaultValueChangedId,
		Args:        []any{"enabled", false, true, "200"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_property_default_value_changed_revision.yaml"),
		OperationId: "createOneGroup",
	}}, errs)
}

// CL: changing response body default value
func TestResponseSchemaDefaultValueUpdatedCheck(t *testing.T) {
	s1, err := open("../data/checker/response_property_default_value_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_default_value_changed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Responses.Value("404").Value.Content["text/plain"].Schema.Value.Default = "new default value"
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyDefaultValueChangedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponseBodyDefaultValueChangedId,
		Args:        []any{"text/plain", "Error", "new default value", "404"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_property_default_value_changed_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// CL: adding response body default value or response body property default value
func TestResponsePropertyDefaultValueAddedCheck(t *testing.T) {
	s1, err := open("../data/checker/response_property_default_value_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_default_value_changed_base.yaml")
	require.NoError(t, err)

	s1.Spec.Paths.Value("/api/v1.0/groups").Post.Responses.Value("404").Value.Content["text/plain"].Schema.Value.Default = nil
	s1.Spec.Components.Schemas["GroupView"].Value.Properties["data"].Value.Properties["created"].Value.Default = nil

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyDefaultValueChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 2)
	requireApiChanges(t, []checker.ApiChange{{
		Id:          checker.ResponseBodyDefaultValueAddedId,
		Args:        []any{"text/plain", "Error", "404"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_property_default_value_changed_base.yaml"),
		OperationId: "createOneGroup",
	}, {
		Id:          checker.ResponsePropertyDefaultValueAddedId,
		Args:        []any{"created", "2020-01-01T00:00:00Z", "200"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_property_default_value_changed_base.yaml"),
		OperationId: "createOneGroup",
	}}, errs)
}

// CL: removing response body default value or response body property default value
func TestResponsePropertyDefaultValueRemovedCheck(t *testing.T) {
	s1, err := open("../data/checker/response_property_default_value_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_default_value_changed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Responses.Value("404").Value.Content["text/plain"].Schema.Value.Default = nil
	s2.Spec.Components.Schemas["GroupView"].Value.Properties["data"].Value.Properties["created"].Value.Default = nil

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyDefaultValueChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 2)
	requireApiChanges(t, []checker.ApiChange{{
		Id:          checker.ResponseBodyDefaultValueRemovedId,
		Args:        []any{"text/plain", "Error", "404"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_property_default_value_changed_base.yaml"),
		OperationId: "createOneGroup",
	}, {
		Id:          checker.ResponsePropertyDefaultValueRemovedId,
		Args:        []any{"created", "2020-01-01T00:00:00Z", "200"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_property_default_value_changed_base.yaml"),
		OperationId: "createOneGroup",
	}}, errs)
}
