package checker_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/oasdiff/oasdiff/utils"
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
		Args:        []any{utils.StringList{"string"}, "", utils.StringList{"object"}, "", "200"},
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
		Args:        []any{"data/name", utils.StringList{"string"}, "", utils.StringList{"integer"}, "", "200"},
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
		Args:        []any{"data/name", utils.StringList{"string"}, "hostname", utils.StringList{"string"}, "uuid", "200"},
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
			Args:        []any{"/anyOf[#/components/schemas/Dog]/breed/anyOf[#/components/schemas/Breed2]/name", utils.StringList{"string"}, "", utils.StringList{"number"}, "", "200"},
			Level:       checker.ERR,
			Operation:   "GET",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/response_property_any_of_complex_revision.yaml"),
			OperationId: "listPets",
		},
		{
			Id:          checker.ResponsePropertyTypeChangedId,
			Args:        []any{"/anyOf[subschema #3: Rabbit]/", utils.StringList{"string"}, "", utils.StringList{"number"}, "", "200"},
			Level:       checker.ERR,
			Operation:   "GET",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/response_property_any_of_complex_revision.yaml"),
			OperationId: "listPets",
		},
		{
			Id:          checker.ResponsePropertyTypeChangedId,
			Args:        []any{"/anyOf[subschema #4 -> subschema #5]/", utils.StringList{"string"}, "", utils.StringList{"number"}, "", "200"},
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
		Args:        []any{"data/name", utils.StringList{"string"}, "", utils.StringList{"integer", "string"}, "", "200"},
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
		Args:        []any{"/additionalProperties/property1", utils.StringList{"integer"}, "", utils.StringList{"string"}, "", "200"},
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
		Args:        []any{"composite-property/additionalProperties/property1", utils.StringList{"integer"}, "", utils.StringList{"string"}, "", "200"},
		Level:       checker.ERR,
		Operation:   "GET",
		Path:        "/value",
		Source:      load.NewSource("../data/additional-properties/embedded-revision.yaml"),
		OperationId: "get_value",
	}, errs[0])
}

// CL: specializing type from oneOf: {string, number} to string is not reported as type-changed
func TestResponseSchemaTypeSpecialized(t *testing.T) {
	s1, err := open("../data/checker/response_property_one_of_specialized_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_one_of_specialized_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyTypeChangedCheck), d, osm, checker.ERR)
	require.Empty(t, errs)
}

// CL: changing response field type while also modifying oneOf (not specialized) is still reported as type-changed
func TestResponseSchemaTypeNotSpecialized(t *testing.T) {
	s1, err := open("../data/checker/response_property_one_of_specialized_base2.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_one_of_specialized_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyTypeChangedCheck), d, osm, checker.INFO)

	require.Len(t, errs, 1)

	require.ElementsMatch(t, []checker.ApiChange{
		{
			Id:        checker.ResponsePropertyTypeChangedId,
			Args:      []any{"id", utils.StringList{"string"}, "", utils.StringList{"integer"}, "", "201"},
			Level:     checker.ERR,
			Operation: "POST",
			Path:      "/users",
			Source:    load.NewSource("../data/checker/response_property_one_of_specialized_revision.yaml"),
		}}, errs)
}
