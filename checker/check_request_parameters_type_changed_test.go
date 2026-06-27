package checker_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: changing request path parameter type
func TestRequestPathParamTypeChanged(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Parameters[0].Value.Schema.Value.Type = &openapi3.Types{"integer"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestParameterTypeChangedId,
		Args:        []any{"path", "groupId", []string{"string"}, "", []string{"integer"}, ""},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_parameter_type_changed_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// CL: changing request query parameter type
func TestRequestQueryParamTypeChanged(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Parameters[1].Value.Schema.Value.Type = &openapi3.Types{"integer"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestParameterTypeChangedId,
		Args:        []any{"query", "token", []string{"string"}, "uuid", []string{"integer"}, "uuid"},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_parameter_type_changed_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// CL: changing request header parameter type
func TestRequestQueryHeaderTypeChanged(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Parameters[2].Value.Schema.Value.Type = &openapi3.Types{"integer"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestParameterTypeChangedId,
		Args:        []any{"header", "X-Request-ID", []string{"string"}, "uuid", []string{"integer"}, "uuid"},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_parameter_type_changed_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// CL: changing request path parameter format
func TestRequestPathParamFormatChanged(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Parameters[0].Value.Schema.Value.Format = "uuid"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestParameterTypeChangedId,
		Args:        []any{"path", "groupId", []string{"string"}, "", []string{"string"}, "uuid"},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_parameter_type_changed_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// CL: changing request query parameter format
func TestRequestQueryParamFormatChanged(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Parameters[1].Value.Schema.Value.Format = "uri"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestParameterTypeChangedId,
		Args:        []any{"query", "token", []string{"string"}, "uuid", []string{"string"}, "uri"},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_parameter_type_changed_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// CL: changing request header parameter format
func TestRequestQueryHeaderFormatChanged(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_parameter_type_changed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Parameters[2].Value.Schema.Value.Format = "uri"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestParameterTypeChangedId,
		Args:        []any{"header", "X-Request-ID", []string{"string"}, "uuid", []string{"string"}, "uri"},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_parameter_type_changed_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// CL: changing request path parameter type by adding "string"
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
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestParameterTypeGeneralizedId,
		Args:        []any{"path", "groupId", []string{"integer"}, "", []string{"integer", "string"}, ""},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_parameter_type_changed_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// CL: changing request path parameter type by replacing "integer" with "number"
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
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestParameterTypeGeneralizedId,
		Args:        []any{"path", "groupId", []string{"integer", "string"}, "", []string{"number", "string"}, ""},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_parameter_type_changed_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// CL: a query parameter changing from a single type to a oneOf list of types is
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
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterListOfTypesWidenedId, errs[0].GetId())
}

// BC: changing request's query param property type from number to string is breaking
func TestBreaking_ReqQueryParamTypeNumberToString(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_property_type_changed_base_num.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/request_parameter_property_type_changed_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterPropertyTypeChangedId, errs[0].GetId())
	require.Equal(t, "for the `query` request parameter `filters`, the type/format of property `groupId` was changed from `number`/`` to `string`/``", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, checker.WARN, errs[0].GetLevel())
}

// BC: specializing request's query param property type from string to number is breaking
func TestBreaking_ReqQueryParamTypeStringToNumber(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_property_type_changed_revision.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/request_parameter_property_type_changed_base_num.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterPropertyTypeSpecializedId, errs[0].GetId())
	require.Equal(t, "for the `query` request parameter `filters`, the type/format of property `groupId` was specialized from `string`/`` to `number`/``", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, checker.ERR, errs[0].GetLevel())
}

// CL: generalizing request's query param property type from integer to number
func TestBreaking_ReqQueryParamTypeIntegerToNumber(t *testing.T) {
	s1, err := open("../data/checker/request_parameter_property_type_changed_base_int.yaml")
	require.NoError(t, err)

	s2, err := open("../data/checker/request_parameter_property_type_changed_base_num.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterPropertyTypeGeneralizedId, errs[0].GetId())
	require.Equal(t, "for the `query` request parameter `filters`, the type/format of property `groupId` was generalized from `integer`/`` to `number`/``", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, checker.INFO, errs[0].GetLevel())
}

// CL: widening a query parameter from a scalar to a form/explode array of the
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
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterTypeGeneralizedId, errs[0].GetId())
	require.Equal(t, checker.INFO, errs[0].GetLevel())
}

// CL: widening a path parameter from scalar to array stays breaking, since
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
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterTypeChangedId, errs[0].GetId())
	require.Equal(t, checker.ERR, errs[0].GetLevel())
}

// CL: removing the format constraint of a request path parameter is a generalization, not breaking
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
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestParameterTypeGeneralizedId,
		Args:        []any{"path", "groupId", []string{"string"}, "uuid", []string{"string"}, ""},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_parameter_type_changed_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// BC: removing the format constraint of a response property is breaking,
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
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponsePropertyTypeChangedId, errs[0].GetId())
	require.Equal(t, checker.ERR, errs[0].GetLevel())
}

// CL: under OpenAPI 3.1, widening a nullable scalar query parameter to a
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

// CL (guard): the null-stripping relaxation must not over-apply. A multi-non-null
// base type ([string, integer] -> array) is genuinely breaking and must stay so.
func TestRequestQueryParamMultiTypeToArrayStillBreaking(t *testing.T) {
	s1, err := open("../data/checker/request_param_multitype_to_array_31_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_param_multitype_to_array_31_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterTypeChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterTypeChangedId, errs[0].GetId(),
		"a multi-non-null base type to array is breaking; the null relaxation must not apply")
	require.Equal(t, checker.ERR, errs[0].GetLevel())
}

// CL (soundness): a scalar -> form/explode array is only safe when the array
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
