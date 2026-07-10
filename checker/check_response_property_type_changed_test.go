package checker_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// changing a response schema type
func TestResponseSchemaTypeChangedCheck(t *testing.T) {
	s1, err := open("../data/checker/response_schema_type_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_schema_type_changed_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyTypeChangedCheck), d, osm, checker.ERR)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponseBodyTypeChangedId,
		Args:        []any{"type", "string", "object", "200"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_schema_type_changed_revision.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// changing a response property schema type from string to integer
func TestResponsePropertyTypeChangedCheck(t *testing.T) {
	s1, err := open("../data/checker/response_schema_type_changed_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_schema_type_changed_revision.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value.Properties["data"].Value.Properties["name"].Value.Type = &openapi3.Types{"integer"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyTypeChangedCheck), d, osm, checker.ERR)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponsePropertyTypeChangedId,
		Args:        []any{"data/name", "type", "string", "integer", "200"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_schema_type_changed_revision.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// changing a response property schema format
func TestResponsePropertyFormatChangedCheck(t *testing.T) {
	s1, err := open("../data/checker/response_schema_format_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_schema_format_changed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value.Properties["data"].Value.Properties["name"].Value.Format = "uuid"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyTypeChangedCheck), d, osm, checker.ERR)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponsePropertyTypeChangedId,
		Args:        []any{"data/name", "format", "hostname", "uuid", "200"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_schema_format_changed_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// changing properties of subschemas under allOf
func TestResponsePropertyAnyOfModified(t *testing.T) {
	s1, err := open("../data/checker/response_property_any_of_complex_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_any_of_complex_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyTypeChangedCheck), d, osm, checker.INFO)

	requireApiChanges(t, []checker.ApiChange{
		{
			Id:          checker.ResponsePropertyTypeChangedId,
			Args:        []any{"anyOf[#/components/schemas/Dog]/breed/anyOf[#/components/schemas/Breed2]/name", "type", "string", "number", "200"},
			Operation:   "GET",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/response_property_any_of_complex_revision.yaml"),
			OperationId: "listPets",
		},
		{
			Id:          checker.ResponsePropertyTypeChangedId,
			Args:        []any{"anyOf[subschema #3: Rabbit]/", "type", "string", "number", "200"},
			Operation:   "GET",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/response_property_any_of_complex_revision.yaml"),
			OperationId: "listPets",
		},
		{
			Id:          checker.ResponsePropertyTypeChangedId,
			Args:        []any{"anyOf[subschema #4 -> subschema #5]/", "type", "string", "number", "200"},
			Operation:   "GET",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/response_property_any_of_complex_revision.yaml"),
			OperationId: "listPets",
		}}, errs)
}

// changing a response property schema type from a single value to to multiple types
func TestResponseSchemaTypeMultiCheck(t *testing.T) {
	s1, err := open("../data/checker/response_schema_type_changed_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_schema_type_changed_revision.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value.Properties["data"].Value.Properties["name"].Value.Type = &openapi3.Types{"integer", "string"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyTypeChangedCheck), d, osm, checker.ERR)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponsePropertyTypeGeneralizedId,
		Args:        []any{"data/name", "type", "string", "integer, string", "200"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_schema_type_changed_revision.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// changing an additionalResponse property schema type from integer to string is breaking
func TestResponseAdditionalPropertyTypeChangedCheck(t *testing.T) {
	s1, err := open("../data/additional-properties/base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/additional-properties/revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyTypeChangedCheck), d, osm, checker.ERR)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponsePropertyTypeChangedId,
		Args:        []any{"additionalProperties/property1", "type", "integer", "string", "200"},
		Operation:   "GET",
		Path:        "/value",
		Source:      load.NewSource("../data/additional-properties/revision.yaml"),
		OperationId: "get_value",
	}, errs)
}

// changing an embedded additionalResponse property schema type from integer to string is breaking
func TestResponseEmbeddedAdditionalPropertyTypeChangedCheck(t *testing.T) {
	s1, err := open("../data/additional-properties/embedded-base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/additional-properties/embedded-revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyTypeChangedCheck), d, osm, checker.ERR)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponsePropertyTypeChangedId,
		Args:        []any{"composite-property/additionalProperties/property1", "type", "integer", "string", "200"},
		Operation:   "GET",
		Path:        "/value",
		Source:      load.NewSource("../data/additional-properties/embedded-revision.yaml"),
		OperationId: "get_value",
	}, errs)
}

// setResponseBodyType sets the 200 response body schema type on the /test GET.
func setResponseBodyType(t *testing.T, s *load.SpecInfo, types *openapi3.Types) {
	t.Helper()
	s.Spec.Paths.Value("/test").Get.Responses.Value("200").Value.Content["application/json"].Schema.Value.Type = types
}

// narrowing a response type set ([string, integer] -> [string]) is not
// breaking; the server returns fewer kinds of values, all of which the client
// already handled. (#1003 / #989 Gap 2)
func TestResponseBodyTypeNarrowingMultiTypeNotBreaking(t *testing.T) {
	s1, err := open("../data/type-change/simple-response.yaml")
	require.NoError(t, err)
	s2, err := open("../data/type-change/simple-response.yaml")
	require.NoError(t, err)
	setResponseBodyType(t, s1, &openapi3.Types{"string", "integer"})
	setResponseBodyType(t, s2, &openapi3.Types{"string"})

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyTypeChangedCheck), d, osm, checker.INFO)
	require.False(t, containsId(errs, checker.ResponseBodyTypeChangedId),
		"narrowing a response type set is non-breaking; must not report response-body-type-changed")
	require.True(t, containsId(errs, checker.ResponseBodyTypeSpecializedId),
		"narrowing a response type set is surfaced as a non-breaking specialization (#989 Gap 2)")
}

// CL: narrowing a response type (number -> integer) is reported as a non-breaking
// specialization (info), not suppressed and not breaking. (#989 Gap 2)
func TestResponseBodyTypeSpecializedIsInfo(t *testing.T) {
	s1, err := open("../data/type-change/simple-response.yaml")
	require.NoError(t, err)
	s2, err := open("../data/type-change/simple-response.yaml")
	require.NoError(t, err)
	setResponseBodyType(t, s1, &openapi3.Types{"number"})
	setResponseBodyType(t, s2, &openapi3.Types{"integer"})

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyTypeChangedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	c := requireChange(t, errs, checker.ResponseBodyTypeSpecializedId)
	require.Equal(t, checker.INFO, c.GetLevel())
	require.Equal(t, "the response's body `type` was narrowed from `number` to `integer` for status `200`",
		c.GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// narrowing a previously untyped response to a concrete type (no type ->
// [string]) is not breaking; the server returns fewer kinds of values.
func TestResponseBodyTypeAddedFromUntypedNotBreaking(t *testing.T) {
	s1, err := open("../data/type-change/simple-response.yaml")
	require.NoError(t, err)
	s2, err := open("../data/type-change/simple-response.yaml")
	require.NoError(t, err)
	setResponseBodyType(t, s1, nil)
	setResponseBodyType(t, s2, &openapi3.Types{"string"})

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyTypeChangedCheck), d, osm, checker.INFO)
	require.False(t, containsId(errs, checker.ResponseBodyTypeChangedId),
		"narrowing an untyped response to a concrete type is non-breaking")
}

// widening a response type set ([string] -> [string, integer]) is breaking;
// the server may now return a type the client did not handle.
func TestResponseBodyTypeWideningStillBreaking(t *testing.T) {
	s1, err := open("../data/type-change/simple-response.yaml")
	require.NoError(t, err)
	s2, err := open("../data/type-change/simple-response.yaml")
	require.NoError(t, err)
	setResponseBodyType(t, s1, &openapi3.Types{"string"})
	setResponseBodyType(t, s2, &openapi3.Types{"string", "integer"})

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyTypeChangedCheck), d, osm, checker.ERR)
	require.True(t, containsId(errs, checker.ResponseBodyTypeGeneralizedId),
		"widening a response type set is breaking")
}

// removing the type entirely from a response ([string] -> no type) is
// breaking; the server may now return any value.
func TestResponseBodyTypeRemovedStillBreaking(t *testing.T) {
	s1, err := open("../data/type-change/simple-response.yaml")
	require.NoError(t, err)
	s2, err := open("../data/type-change/simple-response.yaml")
	require.NoError(t, err)
	setResponseBodyType(t, s1, &openapi3.Types{"string"})
	setResponseBodyType(t, s2, nil)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyTypeChangedCheck), d, osm, checker.ERR)
	require.True(t, containsId(errs, checker.ResponseBodyTypeChangedId),
		"removing the type from a response is breaking")
}

// a response type narrowing that co-occurs with a breaking format change is
// breaking; the safe type axis must not mask the format axis. [string, integer]
// -> [integer] narrows the type (not breaking on its own), but int32 -> int64
// widens the format (the server may now return values outside the range a client
// expecting int32 can hold), which is breaking.
func TestResponseBodyTypeNarrowWithBreakingFormatStillBreaking(t *testing.T) {
	s1, err := open("../data/type-change/simple-response.yaml")
	require.NoError(t, err)
	s2, err := open("../data/type-change/simple-response.yaml")
	require.NoError(t, err)
	base := s1.Spec.Paths.Value("/test").Get.Responses.Value("200").Value.Content["application/json"].Schema.Value
	base.Type = &openapi3.Types{"string", "integer"}
	base.Format = "int32"
	rev := s2.Spec.Paths.Value("/test").Get.Responses.Value("200").Value.Content["application/json"].Schema.Value
	rev.Type = &openapi3.Types{"integer"}
	rev.Format = "int64"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyTypeChangedCheck), d, osm, checker.ERR)
	require.True(t, containsId(errs, checker.ResponseBodyTypeChangedId),
		"a type narrowing must not mask a co-occurring breaking format change (int32 -> int64)")
}

// Under a loosely-typed media type (application/xml), string -> object is
// compatible, not a narrowing: "compatible", not "specialized".
func TestResponseBodyTypeLooselyTypedCompatible(t *testing.T) {
	s1, err := open("../data/checker/response_xml_type_compatible_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_xml_type_compatible_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyTypeChangedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponseBodyTypeCompatibleId,
		Args:        []any{"type", "string", "object", "200"},
		Comment:     checker.TypeChangeLooselyTypedCommentId,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_xml_type_compatible_revision.yaml"),
		OperationId: "createOneGroup",
	}, errs)
	require.Equal(t, "the response's body `type` changed from `string` to `object` for status `200` (backward compatible)", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}
