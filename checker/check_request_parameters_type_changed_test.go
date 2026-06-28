package checker_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// changing request path parameter type
func TestRequestPathParamTypeChanged(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Parameters[0].Value.Schema.Value.Type = &openapi3.Types{"integer"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestParameterTypeChangedId,
		Args:        []any{"path", "groupId", "type", "string", "integer"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_parameter_type_changed_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// changing request query parameter type
func TestRequestQueryParamTypeChanged(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Parameters[1].Value.Schema.Value.Type = &openapi3.Types{"integer"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestParameterTypeChangedId,
		Args:        []any{"query", "token", "type", "string", "integer"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_parameter_type_changed_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// changing request header parameter type
func TestRequestQueryHeaderTypeChanged(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Parameters[2].Value.Schema.Value.Type = &openapi3.Types{"integer"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestParameterTypeChangedId,
		Args:        []any{"header", "X-Request-ID", "type", "string", "integer"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_parameter_type_changed_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// changing request path parameter format
func TestRequestPathParamFormatChanged(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Parameters[0].Value.Schema.Value.Format = "uuid"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestParameterTypeChangedId,
		Args:        []any{"path", "groupId", "format", "none", "uuid"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_parameter_type_changed_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// changing request query parameter format
func TestRequestQueryParamFormatChanged(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Parameters[1].Value.Schema.Value.Format = "uri"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestParameterTypeChangedId,
		Args:        []any{"query", "token", "format", "uuid", "uri"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_parameter_type_changed_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// changing request header parameter format
func TestRequestQueryHeaderFormatChanged(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Parameters[2].Value.Schema.Value.Format = "uri"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestParameterTypeChangedId,
		Args:        []any{"header", "X-Request-ID", "format", "uuid", "uri"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_parameter_type_changed_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// changing request path parameter type by adding "string"
func TestRequestPathParamTypeAddString(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)

	s1.Spec.Paths.Value("/api/v1.0/groups").Post.Parameters[0].Value.Schema.Value.Type = &openapi3.Types{"integer"}
	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Parameters[0].Value.Schema.Value.Type = &openapi3.Types{"integer", "string"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestParameterTypeGeneralizedId,
		Args:        []any{"path", "groupId", "type", "integer", "integer, string"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_parameter_type_changed_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// changing request path parameter type by replacing "integer" with "number"
func TestRequestPathParamTypeIntegerToNumber(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)

	s1.Spec.Paths.Value("/api/v1.0/groups").Post.Parameters[0].Value.Schema.Value.Type = &openapi3.Types{"integer", "string"}
	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Parameters[0].Value.Schema.Value.Type = &openapi3.Types{"number", "string"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestParameterTypeGeneralizedId,
		Args:        []any{"path", "groupId", "type", "integer, string", "number, string"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_parameter_type_changed_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// a query parameter changing from a single type to a oneOf list of types is
// reported once, by the list-of-types checker. RequestParameterTypeChangedCheck
// suppresses its own report for the same change so it is not duplicated when both
// checks run together.
func TestRequestQueryParamSingleToListOfTypesNotDuplicated(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)

	queryParam := s2.Spec.Paths.Value("/api/v1.0/groups").Post.Parameters[1].Value
	queryParam.Schema.Value.Type = nil
	queryParam.Schema.Value.Format = ""
	queryParam.Schema.Value.OneOf = openapi3.SchemaRefs{
		openapi3.NewSchemaRef("", openapi3.NewStringSchema()),
		openapi3.NewSchemaRef("", openapi3.NewIntegerSchema()),
	}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := checker.NewConfig(checker.BackwardCompatibilityChecks{
		checker.RequestParameterTypeChangedCheck,
		checker.RequestParameterListOfTypesChangedCheck,
	})
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)
	requireSingleChange(t, errs, checker.RequestParameterListOfTypesWidenedId)
}

// changing request's query param property type from number to string is breaking
func TestBreaking_ReqQueryParamTypeNumberToString(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_property_type_changed_base_num.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/request_parameter_property_type_changed_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, "for the `query` request parameter `filters`, the `type` of property `groupId` was changed from `number` to `string`", requireChange(t, errs, checker.RequestParameterPropertyTypeChangedId).GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, checker.WARN, errs[0].GetLevel())
}

// specializing request's query param property type from string to number is breaking
func TestBreaking_ReqQueryParamTypeStringToNumber(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_property_type_changed_revision.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/request_parameter_property_type_changed_base_num.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, "for the `query` request parameter `filters`, the `type` of property `groupId` was narrowed from `string` to `number`", requireChange(t, errs, checker.RequestParameterPropertyTypeSpecializedId).GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, checker.ERR, errs[0].GetLevel())
}

// generalizing request's query param property type from integer to number
func TestBreaking_ReqQueryParamTypeIntegerToNumber(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_property_type_changed_base_int.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/request_parameter_property_type_changed_base_num.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, "for the `query` request parameter `filters`, the `type` of property `groupId` was widened from `integer` to `number`", requireChange(t, errs, checker.RequestParameterPropertyTypeGeneralizedId).GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, checker.INFO, errs[0].GetLevel())
}

// widening a query parameter from a scalar to a form/explode array of the
// same type is backwards-compatible (one-element array on the wire), so it
// should be reported as a generalization rather than a breaking change.
// Reproduces issue #689.
func TestRequestQueryParamScalarToFormExplodeArray(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)

	// Wrap the base scalar in a form/explode array whose item is the scalar
	// unchanged (same type and constraints), which is the backwards-compatible
	// widening: a single value on the wire is a valid one-element array.
	queryParam := s2.Spec.Paths.Value("/api/v1.0/groups").Post.Parameters[1].Value
	item := *queryParam.Schema.Value
	queryParam.Schema.Value = &openapi3.Schema{
		Type:  &openapi3.Types{"array"},
		Items: &openapi3.SchemaRef{Value: &item},
	}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
	requireSingleChange(t, errs, checker.RequestParameterTypeGeneralizedId)
	require.Equal(t, checker.INFO, errs[0].GetLevel())
}

// widening a path parameter from scalar to array stays breaking, since
// path parameters use simple style (not form), so a single value on the wire
// no longer matches the new array schema.
func TestRequestPathParamScalarToArrayStillBreaking(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)

	pathParam := s2.Spec.Paths.Value("/api/v1.0/groups").Post.Parameters[0].Value
	pathParam.Schema.Value.Type = &openapi3.Types{"array"}
	pathParam.Schema.Value.Items = &openapi3.SchemaRef{
		Value: &openapi3.Schema{Type: &openapi3.Types{"string"}},
	}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
	requireSingleChange(t, errs, checker.RequestParameterTypeChangedId)
	require.Equal(t, checker.ERR, errs[0].GetLevel())
}

// removing the format constraint of a request path parameter is a generalization, not breaking
func TestRequestPathParamFormatRemoved(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)

	s1.Spec.Paths.Value("/api/v1.0/groups").Post.Parameters[0].Value.Schema.Value.Format = "uuid"
	// s2 has no format set (empty string), simulating the removal

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestParameterTypeGeneralizedId,
		Args:        []any{"path", "groupId", "format", "uuid", "none"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_parameter_type_changed_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// removing the format constraint of a response property is breaking,
// since clients may depend on the declared format for parsing.
func TestResponsePropertyFormatRemovedCheck(t *testing.T) {
	s1, err := open("../data/checker/response_schema_format_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_schema_format_changed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value.Properties["data"].Value.Properties["name"].Value.Format = ""

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyTypeChangedCheck), d, osm, checker.ERR)
	requireSingleChange(t, errs, checker.ResponsePropertyTypeChangedId)
	require.Equal(t, checker.ERR, errs[0].GetLevel())
}

// under OpenAPI 3.1, widening a nullable scalar query parameter to a
// nullable form/explode array of the same scalar is still backwards-compatible
// (#918). "null" is stripped from both sides before the scalar-to-array check,
// so preserving or adding nullability does not turn a safe widening into a
// breaking change.
func TestRequestQueryParamScalarToFormExplodeArray_31Nullable(t *testing.T) {
	for _, tc := range []struct {
		name     string
		base     string
		revision string
	}{
		// scalar string -> [array, null] (null added on the array side)
		{"scalar to nullable array", "request_param_scalar_to_array_31_base.yaml", "request_param_scalar_to_array_31_revision.yaml"},
		// [string, null] -> [array, null], nullable items (nullable on both sides)
		{"nullable scalar to nullable array", "request_param_nullable_to_array_31_base.yaml", "request_param_nullable_to_array_31_revision.yaml"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s1, err := open("../data/checker/" + tc.base)
			require.NoError(t, err)
			s2, err := open("../data/checker/" + tc.revision)
			require.NoError(t, err)

			d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
			require.NoError(t, err)
			errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
			require.Len(t, errs, 1)
			require.Equal(t, checker.RequestParameterTypeGeneralizedId, errs[0].GetId(),
				"nullable scalar->form/explode array widening must be a generalization, not breaking")
			require.Equal(t, checker.INFO, errs[0].GetLevel())
		})
	}
}

// widening a weakly-typed (query) parameter from a union of scalar types to a
// form/explode array is safe when the item type accepts every value the base did.
// A query value is a string on the wire, so [string, integer] -> array<string> is
// a generalization: the string branch already accepted every wire value, and a
// single value is a valid one-element array. Mirrors the scalar-to-scalar case
// [string, integer] -> string, which is also non-breaking. The change carries the
// form/explode comment so the verdict is not surprising.
func TestRequestQueryParamMultiTypeToFormExplodeArraySafe(t *testing.T) {
	s1, err := open("../data/checker/request_param_multitype_to_array_31_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_param_multitype_to_array_31_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterTypeGeneralizedId, errs[0].GetId(),
		"[string,integer]->array<string> is safe: the item type accepts every value the base did on the wire")
	require.Equal(t, checker.INFO, errs[0].GetLevel())
	require.Equal(t, "This parameter uses form/explode serialization, where a single value is a valid one-element array, so widening it to an array whose items still accept the previous values does not break existing clients.", errs[0].GetComment(checker.NewDefaultLocalizer()))
}

// (guard): the widening is only safe when the item type accepts every base
// value. [string, integer] -> array<integer> drops the string branch, so a value
// like ?token=abc that validated under the base (string) is rejected by the
// integer item. It must stay breaking.
func TestRequestQueryParamWideningToNarrowerItemTypeStillBreaking(t *testing.T) {
	s1, err := open("../data/checker/request_param_multitype_to_array_31_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_param_multitype_to_array_integer_31_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterTypeChangedId, errs[0].GetId(),
		"an item type that does not accept every base value (string -> integer) is breaking")
	require.Equal(t, checker.ERR, errs[0].GetLevel())
}

// (soundness): a scalar -> form/explode array is only safe when the array
// items accept every value the base scalar accepted. Adding an item constraint
// (here a pattern that excludes digits) rejects previously-valid values like
// "5", so it must be breaking, not a generalization (#1024 follow-up).
func TestRequestQueryParamScalarToConstrainedArrayBreaking(t *testing.T) {
	s1, err := open("../data/checker/request_param_scalar_to_constrained_array_31_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_param_scalar_to_constrained_array_31_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterTypeChangedId, errs[0].GetId(),
		"scalar->array whose items add a constraint rejects values valid under the base, so it is breaking")
	require.Equal(t, checker.ERR, errs[0].GetLevel())
}

// (soundness): the two safety axes are independent. Here the type axis passes
// via weak typing ([string,integer] is accepted as string on the wire), but the
// item adds a pattern that rejects previously-valid values like "5", so the
// constraint axis fails and the widening is breaking. Guards that the "anything
// to string" type rule does not short-circuit the value-constraint check.
func TestRequestQueryParamMultiTypeToConstrainedArrayBreaking(t *testing.T) {
	s1, err := open("../data/checker/request_param_multitype_to_array_31_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_param_multitype_to_constrained_array_31_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterTypeChangedId, errs[0].GetId(),
		"type axis passes (weak typing) but the added pattern narrows values, so it is breaking")
	require.Equal(t, checker.ERR, errs[0].GetLevel())
}
