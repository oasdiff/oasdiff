package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// changing optional request property to write-only
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

// changing optional request property to not write-only
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

// changing optional request property to read-only
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
		Comment:     checker.RequestPropertyBecameReadOnlyCommentId,
		Args:        []any{"name"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_optional_property_write_only_read_only_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
	require.Equal(t, checker.WARN, errs[0].GetLevel())
}

// changing optional request property to not read-only
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
	require.Equal(t, checker.INFO, errs[0].GetLevel())
}

// changing required request property to write-only
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

// changing required request property to not write-only
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

// changing required request property to read-only
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
		Comment:     checker.RequestPropertyBecameReadOnlyCommentId,
		Args:        []any{"id"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_optional_property_write_only_read_only_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
	require.Equal(t, checker.WARN, errs[0].GetLevel())
}

// changing required request property to not read-only
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
	require.Equal(t, checker.ERR, errs[0].GetLevel())
}

// a property that stops being read-only and leaves the required list in the
// same change is not required in requests on either side
func TestRequestRequiredPropertyBecameNonReadOnlyAndOptional(t *testing.T) {
	s1, err := open("../data/checker/request_optional_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/request_optional_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s1.Spec.Paths.Value("/api/v1.0/groups").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["id"].Value.ReadOnly = true
	s2.Spec.Paths.Value("/api/v1.0/groups").Post.RequestBody.Value.Content["application/json"].Schema.Value.Required = nil

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyWriteOnlyReadOnlyCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestOptionalPropertyBecameNonReadOnlyCheckId,
		Args:        []any{"id"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_optional_property_write_only_read_only_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
	require.Equal(t, checker.INFO, errs[0].GetLevel())
}

// OpenAPI 3.1 dropped the rule that required applies to a readOnly property in
// responses only, so clients already sent the property and nothing becomes
// required in requests
func TestRequestRequiredPropertyBecameNonReadOnly31(t *testing.T) {
	s1, err := open("../data/checker/request_optional_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/request_optional_property_write_only_read_only_base.yaml")
	require.NoError(t, err)

	s1.Spec.OpenAPI, s2.Spec.OpenAPI = "3.1.0", "3.1.0"
	s1.Spec.Paths.Value("/api/v1.0/groups").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["id"].Value.ReadOnly = true

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyWriteOnlyReadOnlyCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestRequiredPropertyBecameNonReadOnly31Id, errs[0].GetId())
	require.Equal(t, checker.INFO, errs[0].GetLevel())
	require.Equal(t, "In OpenAPI 3.1 and later, required applies to a read-only property in requests too, so clients already send it and this change cannot invalidate a request.", errs[0].GetComment(checker.NewDefaultLocalizer()))
}
