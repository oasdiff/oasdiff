package checker_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: changing a response schema type
func TestResponseSchemaTypeChangedCheck(t *testing.T) {
	s1, err := open("../data/checker/response_schema_type_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_schema_type_changed_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyTypeChangedCheck), d, osm, checker.ERR)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.ResponseBodyTypeChangedId,
		Args:        []any{[]string{"string"}, "", []string{"object"}, "", "200"},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_schema_type_changed_revision.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// CL: changing a response property schema type from string to integer
func TestResponsePropertyTypeChangedCheck(t *testing.T) {
	s1, err := open("../data/checker/response_schema_type_changed_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_schema_type_changed_revision.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value.Properties["data"].Value.Properties["name"].Value.Type = &openapi3.Types{"integer"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyTypeChangedCheck), d, osm, checker.ERR)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.ResponsePropertyTypeChangedId,
		Args:        []any{"data/name", []string{"string"}, "", []string{"integer"}, "", "200"},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_schema_type_changed_revision.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// CL: changing a response property schema format
func TestResponsePropertyFormatChangedCheck(t *testing.T) {
	s1, err := open("../data/checker/response_schema_format_changed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_schema_format_changed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value.Properties["data"].Value.Properties["name"].Value.Format = "uuid"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyTypeChangedCheck), d, osm, checker.ERR)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.ResponsePropertyTypeChangedId,
		Args:        []any{"data/name", []string{"string"}, "hostname", []string{"string"}, "uuid", "200"},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_schema_format_changed_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// CL: changing properties of subschemas under allOf
func TestResponsePropertyAnyOfModified(t *testing.T) {
	s1, err := open("../data/checker/response_property_any_of_complex_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_any_of_complex_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyTypeChangedCheck), d, osm, checker.INFO)

	require.Len(t, errs, 3)
	require.ElementsMatch(t, []checker.ApiChange{
		{
			Id:          checker.ResponsePropertyTypeChangedId,
			Args:        []any{"anyOf[#/components/schemas/Dog]/breed/anyOf[#/components/schemas/Breed2]/name", []string{"string"}, "", []string{"number"}, "", "200"},
			Level:       checker.ERR,
			Operation:   "GET",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/response_property_any_of_complex_revision.yaml"),
			OperationId: "listPets",
		},
		{
			Id:          checker.ResponsePropertyTypeChangedId,
			Args:        []any{"anyOf[subschema #3: Rabbit]/", []string{"string"}, "", []string{"number"}, "", "200"},
			Level:       checker.ERR,
			Operation:   "GET",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/response_property_any_of_complex_revision.yaml"),
			OperationId: "listPets",
		},
		{
			Id:          checker.ResponsePropertyTypeChangedId,
			Args:        []any{"anyOf[subschema #4 -> subschema #5]/", []string{"string"}, "", []string{"number"}, "", "200"},
			Level:       checker.ERR,
			Operation:   "GET",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/response_property_any_of_complex_revision.yaml"),
			OperationId: "listPets",
		}}, errs)
}

// CL: changing a response property schema type from a single value to to multiple types
func TestResponseSchemaTypeMultiCheck(t *testing.T) {
	s1, err := open("../data/checker/response_schema_type_changed_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_schema_type_changed_revision.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value.Properties["data"].Value.Properties["name"].Value.Type = &openapi3.Types{"integer", "string"}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyTypeChangedCheck), d, osm, checker.ERR)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.ResponsePropertyTypeChangedId,
		Args:        []any{"data/name", []string{"string"}, "", []string{"integer", "string"}, "", "200"},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_schema_type_changed_revision.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// BC: changing an additionalResponse property schema type from integer to string is breaking
func TestResponseAdditionalPropertyTypeChangedCheck(t *testing.T) {
	s1, err := open("../data/additional-properties/base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/additional-properties/revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyTypeChangedCheck), d, osm, checker.ERR)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.ResponsePropertyTypeChangedId,
		Args:        []any{"additionalProperties/property1", []string{"integer"}, "", []string{"string"}, "", "200"},
		Level:       checker.ERR,
		Operation:   "GET",
		Path:        "/value",
		Source:      load.NewSource("../data/additional-properties/revision.yaml"),
		OperationId: "get_value",
	}, errs[0])
}

// BC: changing an embedded additionalResponse property schema type from integer to string is breaking
func TestResponseEmbeddedAdditionalPropertyTypeChangedCheck(t *testing.T) {
	s1, err := open("../data/additional-properties/embedded-base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/additional-properties/embedded-revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyTypeChangedCheck), d, osm, checker.ERR)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.ResponsePropertyTypeChangedId,
		Args:        []any{"composite-property/additionalProperties/property1", []string{"integer"}, "", []string{"string"}, "", "200"},
		Level:       checker.ERR,
		Operation:   "GET",
		Path:        "/value",
		Source:      load.NewSource("../data/additional-properties/embedded-revision.yaml"),
		OperationId: "get_value",
	}, errs[0])
}

// setResponseBodyType sets the 200 response body schema type on the /test GET.
func setResponseBodyType(t *testing.T, s *load.SpecInfo, types *openapi3.Types) {
	t.Helper()
	s.Spec.Paths.Value("/test").Get.Responses.Value("200").Value.Content["application/json"].Schema.Value.Type = types
}

// BC: narrowing a response type set ([string, integer] -> [string]) is backward
// compatible (the server returns fewer kinds of values) and must NOT be reported
// as a breaking response-body-type-changed. (#1003 / #989 Gap 2)
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
}

// BC: adding a type constraint to a previously untyped response (no type -> [string])
// narrows what the server returns, so it is backward compatible.
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

// BC (guard): widening a response type set ([string] -> [string, integer]) IS breaking;
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
	require.True(t, containsId(errs, checker.ResponseBodyTypeChangedId),
		"widening a response type set is breaking")
}

// BC (guard): removing the type entirely from a response ([string] -> no type) IS
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

// BC (guard): a response type narrowing that co-occurs with a breaking format
// change must still be reported; the safe type axis must not mask the format
// axis. [string, integer] -> [integer] narrows the type (backward compatible),
// but int32 -> int64 widens the format (the server may now return values outside
// the range a client expecting int32 can hold), which is breaking.
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
