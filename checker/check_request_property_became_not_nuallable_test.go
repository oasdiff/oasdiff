package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: changing request property to not nullable
func TestRequestPropertyBecameNotNullable(t *testing.T) {
	s1, err := open("../data/checker/request_property_became_nullable_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_became_nullable_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyBecameNotNullableCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestPropertyBecomeNotNullableId,
		Args:        []any{"name"},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_property_became_nullable_base.yaml"),
		OperationId: "addProduct",
	}, errs[0])
}

// CL: changing request property to nullable
func TestRequestPropertyBecameNullable(t *testing.T) {
	s1, err := open("../data/checker/request_property_became_nullable_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_became_nullable_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyBecameNotNullableCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestPropertyBecomeNullableId,
		Args:        []any{"name"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_property_became_nullable_revision.yaml"),
		OperationId: "addProduct",
	}, errs[0])

}

// CL: changing request body to nullable
func TestRequestBodyBecameNullable(t *testing.T) {
	s1, err := open("../data/checker/request_property_became_nullable_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_became_nullable_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Nullable = true

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyBecameNotNullableCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestBodyBecomeNullableId,
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_property_became_nullable_base.yaml"),
		OperationId: "addProduct",
	}, errs[0])
}

// CL: changing request property to nullable via type array (OpenAPI 3.1)
func TestRequestPropertyBecameNullable31(t *testing.T) {
	s1, err := open("../data/checker/request_property_became_nullable_31_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_became_nullable_31_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyBecameNotNullableCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestPropertyBecomeNullableId,
		Args:        []any{"name"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_property_became_nullable_31_revision.yaml"),
		OperationId: "createProduct",
	}, errs[0])
}

// CL: changing request property to not nullable via type array (OpenAPI 3.1)
func TestRequestPropertyBecameNotNullable31(t *testing.T) {
	s1, err := open("../data/checker/request_property_became_nullable_31_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_became_nullable_31_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyBecameNotNullableCheck), d, osm, checker.ERR)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestPropertyBecomeNotNullableId,
		Args:        []any{"name"},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_property_became_nullable_31_base.yaml"),
		OperationId: "createProduct",
	}, errs[0])
}

// CL: type checker does NOT fire for null-only type changes (OpenAPI 3.1)
func TestTypeCheckerSuppressedForNullOnly31(t *testing.T) {
	s1, err := open("../data/checker/request_property_became_nullable_31_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_became_nullable_31_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyTypeChangedCheck), d, osm, checker.INFO)
	require.Empty(t, errs)
}

// CL: changing request body to not nullable
func TestRequestBodyBecameNotNullable(t *testing.T) {
	s1, err := open("../data/checker/request_property_became_nullable_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_became_nullable_base.yaml")
	require.NoError(t, err)

	s1.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Nullable = true

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyBecameNotNullableCheck), d, osm, checker.ERR)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestBodyBecomeNotNullableId,
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_property_became_nullable_base.yaml"),
		OperationId: "addProduct",
	}, errs[0])
}

// CL: response property became nullable via type array (OpenAPI 3.1)
func TestResponsePropertyBecameNullable31(t *testing.T) {
	s1, err := open("../data/checker/request_property_became_nullable_31_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_became_nullable_31_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyBecameNullableCheck), d, osm, checker.ERR)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.ResponsePropertyBecameNullableId,
		Args:        []any{"status", "200"},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_property_became_nullable_31_revision.yaml"),
		OperationId: "createProduct",
	}, errs[0])
}

// CL: request body became nullable via type array (OpenAPI 3.1)
func TestRequestBodyBecameNullable31(t *testing.T) {
	s1, err := open("../data/checker/request_body_nullable_31_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_body_nullable_31_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyBecameNotNullableCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.RequestBodyBecomeNullableId), "expected request-body-became-nullable")
}

// CL: request body became not-nullable via type array (OpenAPI 3.1)
func TestRequestBodyBecameNotNullable31(t *testing.T) {
	s1, err := open("../data/checker/request_body_nullable_31_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_body_nullable_31_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyBecameNotNullableCheck), d, osm, checker.ERR)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.RequestBodyBecomeNotNullableId), "expected request-body-became-not-nullable")
}

// CL: response body became nullable via type array (OpenAPI 3.1)
func TestResponseBodyBecameNullable31(t *testing.T) {
	s1, err := open("../data/checker/request_body_nullable_31_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_body_nullable_31_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyBecameNullableCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.ResponseBodyBecameNullableId), "expected response-body-became-nullable")
}

// CL: response property became nullable via type array (OpenAPI 3.1) — property level check in response
func TestResponsePropertyBecameNullable31Property(t *testing.T) {
	s1, err := open("../data/checker/request_property_became_nullable_31_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_became_nullable_31_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyBecameNullableCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.True(t, containsId(errs, checker.ResponsePropertyBecameNullableId), "expected response-property-became-nullable")
}

// CL: response type checker does NOT fire for null-only type changes (OpenAPI 3.1)
func TestResponseTypeCheckerSuppressedForNullOnly31(t *testing.T) {
	s1, err := open("../data/checker/request_property_became_nullable_31_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_became_nullable_31_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyTypeChangedCheck), d, osm, checker.INFO)
	require.Empty(t, errs)
}
