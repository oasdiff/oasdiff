package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: changing optional request property to write-only
func TestRequestOptionalPropertyBecameWriteOnly(t *testing.T) {
	s1, err := open("../data/checker/request_optional_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/request_optional_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["name"].Value.WriteOnly = true

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyWriteOnlyReadOnlyCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestOptionalPropertyBecameWriteOnlyCheckId,
		Args:        []any{"name"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_optional_property_write_only_read_only_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// CL: changing optional request property to not write-only
func TestRequestOptionalPropertyBecameNotWriteOnly(t *testing.T) {
	s1, err := open("../data/checker/request_optional_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/request_optional_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s1.Spec.Paths.Value("/api/v1.0/groups").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["name"].Value.WriteOnly = true

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyWriteOnlyReadOnlyCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestOptionalPropertyBecameNonWriteOnlyCheckId,
		Args:        []any{"name"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_optional_property_write_only_read_only_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// CL: changing optional request property to read-only
func TestRequestOptionalPropertyBecameReadOnly(t *testing.T) {
	s1, err := open("../data/checker/request_optional_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/request_optional_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["name"].Value.ReadOnly = true

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyWriteOnlyReadOnlyCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestOptionalPropertyBecameReadOnlyCheckId,
		Args:        []any{"name"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_optional_property_write_only_read_only_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// CL: changing optional request property to not read-only
func TestRequestOptionalPropertyBecameNonReadOnly(t *testing.T) {
	s1, err := open("../data/checker/request_optional_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/request_optional_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s1.Spec.Paths.Value("/api/v1.0/groups").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["name"].Value.ReadOnly = true

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyWriteOnlyReadOnlyCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestOptionalPropertyBecameNonReadOnlyCheckId,
		Args:        []any{"name"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_optional_property_write_only_read_only_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// CL: changing required request property to write-only
func TestRequestRequiredPropertyBecameWriteOnly(t *testing.T) {
	s1, err := open("../data/checker/request_optional_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/request_optional_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["id"].Value.WriteOnly = true

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyWriteOnlyReadOnlyCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestRequiredPropertyBecameWriteOnlyCheckId,
		Args:        []any{"id"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_optional_property_write_only_read_only_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// CL: changing required request property to not write-only
func TestRequestRequiredPropertyBecameNotWriteOnly(t *testing.T) {
	s1, err := open("../data/checker/request_optional_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/request_optional_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s1.Spec.Paths.Value("/api/v1.0/groups").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["id"].Value.WriteOnly = true

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyWriteOnlyReadOnlyCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestRequiredPropertyBecameNonWriteOnlyCheckId,
		Args:        []any{"id"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_optional_property_write_only_read_only_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// CL: changing required request property to read-only
func TestRequestRequiredPropertyBecameReadOnly(t *testing.T) {
	s1, err := open("../data/checker/request_optional_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/request_optional_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["id"].Value.ReadOnly = true

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyWriteOnlyReadOnlyCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestRequiredPropertyBecameReadOnlyCheckId,
		Args:        []any{"id"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_optional_property_write_only_read_only_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// CL: changing required request property to not read-only
func TestRequestRequiredPropertyBecameNonReadOnly(t *testing.T) {
	s1, err := open("../data/checker/request_optional_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/request_optional_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s1.Spec.Paths.Value("/api/v1.0/groups").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["id"].Value.ReadOnly = true

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyWriteOnlyReadOnlyCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestRequiredPropertyBecameNonReadOnlyCheckId,
		Args:        []any{"id"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_optional_property_write_only_read_only_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}
