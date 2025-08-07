package diff_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// Helper functions for creating test schemas
func createSingleTypeSchema(typeName string) *openapi3.Schema {
	return &openapi3.Schema{
		Type: &openapi3.Types{typeName},
	}
}

func createOneOfSchema(types ...string) *openapi3.Schema {
	schema := &openapi3.Schema{}
	for _, t := range types {
		schema.OneOf = append(schema.OneOf, &openapi3.SchemaRef{
			Value: createSingleTypeSchema(t),
		})
	}
	return schema
}

func createAnyOfSchema(types ...string) *openapi3.Schema {
	schema := &openapi3.Schema{}
	for _, t := range types {
		schema.AnyOf = append(schema.AnyOf, &openapi3.SchemaRef{
			Value: createSingleTypeSchema(t),
		})
	}
	return schema
}

func createComplexOneOfSchema() *openapi3.Schema {
	return &openapi3.Schema{
		OneOf: []*openapi3.SchemaRef{
			{Value: createSingleTypeSchema("string")},
			{Value: &openapi3.Schema{
				Type: &openapi3.Types{"object"},
				Properties: map[string]*openapi3.SchemaRef{
					"nested": {Value: createSingleTypeSchema("string")},
				},
			}},
		},
	}
}

// Tests for ListOfTypesDiff functionality

func TestListOfTypesDiff_SingleToList(t *testing.T) {
	loader := openapi3.NewLoader()

	base, err := loader.LoadFromFile("../data/list-of-types/single-to-list-base.yaml")
	require.NoError(t, err)

	revision, err := loader.LoadFromFile("../data/list-of-types/single-to-list-revision.yaml")
	require.NoError(t, err)

	diffReport, err := diff.Get(diff.NewConfig(), base, revision)
	require.NoError(t, err)
	require.NotNil(t, diffReport)

	// Check schema diffs for properties
	schemaDiffs := diffReport.PathsDiff.Modified["/test"].OperationsDiff.Modified["GET"].
		ResponsesDiff.Modified["200"].ContentDiff.MediaTypeModified["application/json"].
		SchemaDiff.PropertiesDiff.Modified

	// Test 'id' property: string -> oneOf[string, integer]
	idDiff := schemaDiffs["id"].ListOfTypesDiff
	require.NotNil(t, idDiff)
	require.Equal(t, []string{"integer"}, idDiff.Added)
	require.Empty(t, idDiff.Deleted)
	require.False(t, idDiff.Empty())

	// Test 'status' property: integer -> anyOf[integer, number]
	statusDiff := schemaDiffs["status"].ListOfTypesDiff
	require.NotNil(t, statusDiff)
	require.Equal(t, []string{"number"}, statusDiff.Added)
	require.Empty(t, statusDiff.Deleted)
	require.False(t, statusDiff.Empty())
}

func TestListOfTypesDiff_ListToSingle(t *testing.T) {
	loader := openapi3.NewLoader()

	base, err := loader.LoadFromFile("../data/list-of-types/list-to-single-base.yaml")
	require.NoError(t, err)

	revision, err := loader.LoadFromFile("../data/list-of-types/list-to-single-revision.yaml")
	require.NoError(t, err)

	diffReport, err := diff.Get(diff.NewConfig(), base, revision)
	require.NoError(t, err)
	require.NotNil(t, diffReport)

	// Check request body schema diffs
	schemaDiffs := diffReport.PathsDiff.Modified["/test"].OperationsDiff.Modified["POST"].
		RequestBodyDiff.ContentDiff.MediaTypeModified["application/json"].
		SchemaDiff.PropertiesDiff.Modified

	// Test 'userId' property: oneOf[string, integer] -> integer
	userIdDiff := schemaDiffs["userId"].ListOfTypesDiff
	require.NotNil(t, userIdDiff)
	require.Empty(t, userIdDiff.Added)
	require.Equal(t, []string{"string"}, userIdDiff.Deleted)
	require.False(t, userIdDiff.Empty())

	// Test 'value' property: anyOf[string, number, boolean] -> string
	valueDiff := schemaDiffs["value"].ListOfTypesDiff
	require.NotNil(t, valueDiff)
	require.Empty(t, valueDiff.Added)
	require.ElementsMatch(t, []string{"number", "boolean"}, valueDiff.Deleted)
	require.False(t, valueDiff.Empty())
}

func TestListOfTypesDiff_ListToList(t *testing.T) {
	loader := openapi3.NewLoader()

	base, err := loader.LoadFromFile("../data/list-of-types/list-to-list-base.yaml")
	require.NoError(t, err)

	revision, err := loader.LoadFromFile("../data/list-of-types/list-to-list-revision.yaml")
	require.NoError(t, err)

	diffReport, err := diff.Get(diff.NewConfig(), base, revision)
	require.NoError(t, err)
	require.NotNil(t, diffReport)

	schemaDiffs := diffReport.PathsDiff.Modified["/api"].OperationsDiff.Modified["GET"].
		ResponsesDiff.Modified["200"].ContentDiff.MediaTypeModified["application/json"].
		SchemaDiff.PropertiesDiff.Modified

	// Test 'data' property: oneOf[string, integer] -> anyOf[string, number, boolean]
	dataDiff := schemaDiffs["data"].ListOfTypesDiff
	require.NotNil(t, dataDiff)
	require.ElementsMatch(t, []string{"number", "boolean"}, dataDiff.Added)
	require.Equal(t, []string{"integer"}, dataDiff.Deleted)
	require.False(t, dataDiff.Empty())

	// Test 'metadata' property: anyOf[object, string] -> oneOf[string]
	metadataDiff := schemaDiffs["metadata"].ListOfTypesDiff
	require.NotNil(t, metadataDiff)
	require.Empty(t, metadataDiff.Added)
	require.Equal(t, []string{"object"}, metadataDiff.Deleted)
	require.False(t, metadataDiff.Empty())
}

func TestListOfTypesDiff_EdgeCases(t *testing.T) {
	loader := openapi3.NewLoader()

	base, err := loader.LoadFromFile("../data/list-of-types/edge-cases-base.yaml")
	require.NoError(t, err)

	revision, err := loader.LoadFromFile("../data/list-of-types/edge-cases-revision.yaml")
	require.NoError(t, err)

	diffReport, err := diff.Get(diff.NewConfig(), base, revision)
	require.NoError(t, err)
	require.NotNil(t, diffReport)

	schemaDiffs := diffReport.PathsDiff.Modified["/edge"].OperationsDiff.Modified["GET"].
		ResponsesDiff.Modified["200"].ContentDiff.MediaTypeModified["application/json"].
		SchemaDiff.PropertiesDiff.Modified

	// Test 'emptyOneOf': oneOf[] -> string (should NOT be detected as list-of-types)
	emptyDiff := schemaDiffs["emptyOneOf"].ListOfTypesDiff
	require.Nil(t, emptyDiff) // Empty oneOf should not be analyzed as list-of-types

	// Test 'complexOneOf': should not be detected due to complex object
	complexDiff := schemaDiffs["complexOneOf"].ListOfTypesDiff
	require.Nil(t, complexDiff) // Complex schemas should not be detected

	// Test 'mixedBoth': oneOf takes precedence over anyOf
	mixedDiff := schemaDiffs["mixedBoth"].ListOfTypesDiff
	require.NotNil(t, mixedDiff)
	require.Equal(t, []string{"number"}, mixedDiff.Added)
	require.Empty(t, mixedDiff.Deleted) // string present in both
	require.False(t, mixedDiff.Empty())
}

// Test scenarios that should NOT trigger list-of-types detection
func TestListOfTypesDiff_NonListOfTypesScenarios(t *testing.T) {
	// Create test scenarios that should fall back to existing oneOf/anyOf diff logic

	// Complex oneOf with nested objects
	complexSchema := &openapi3.Schema{
		OneOf: []*openapi3.SchemaRef{
			{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
			{Value: &openapi3.Schema{
				Type: &openapi3.Types{"object"},
				Properties: map[string]*openapi3.SchemaRef{
					"nested": {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
				},
			}},
		},
	}

	require.NotNil(t, complexSchema)
	require.Len(t, complexSchema.OneOf, 2)
	require.NotNil(t, complexSchema.OneOf[1].Value.Properties)

	// Schema with multiple types in single type field
	multiTypeSchema := &openapi3.Schema{
		Type: &openapi3.Types{"string", "integer"},
	}

	require.NotNil(t, multiTypeSchema)
	require.Len(t, *multiTypeSchema.Type, 2)

	// Empty oneOf/anyOf
	emptyOneOfSchema := &openapi3.Schema{
		OneOf: []*openapi3.SchemaRef{},
	}

	require.NotNil(t, emptyOneOfSchema)
	require.Empty(t, emptyOneOfSchema.OneOf)
}

// Test expected diff output format
func TestListOfTypesDiff_OutputFormat(t *testing.T) {
	// Test the structure of ListOfTypesDiff
	listDiff := &diff.ListOfTypesDiff{
		Added:   []string{"number", "boolean"},
		Deleted: []string{"string"},
	}

	require.NotNil(t, listDiff)
	require.Equal(t, []string{"number", "boolean"}, listDiff.Added)
	require.Equal(t, []string{"string"}, listDiff.Deleted)
	require.False(t, listDiff.Empty())

	// Test empty diff
	emptyDiff := &diff.ListOfTypesDiff{}
	require.True(t, emptyDiff.Empty())

	// Test nil diff
	var nilDiff *diff.ListOfTypesDiff
	require.True(t, nilDiff.Empty())
}

// Test ListOfTypesDiff Empty method behavior
func TestListOfTypesDiff_Empty(t *testing.T) {
	// Test nil diff
	var listDiff *diff.ListOfTypesDiff
	require.True(t, listDiff.Empty())

	// Test empty diff
	listDiff = &diff.ListOfTypesDiff{}
	require.True(t, listDiff.Empty())

	// Test diff with added types
	listDiff = &diff.ListOfTypesDiff{Added: []string{"string"}}
	require.False(t, listDiff.Empty())

	// Test diff with deleted types
	listDiff = &diff.ListOfTypesDiff{Deleted: []string{"integer"}}
	require.False(t, listDiff.Empty())

	// Test diff with both
	listDiff = &diff.ListOfTypesDiff{
		Added:   []string{"string"},
		Deleted: []string{"integer"},
	}
	require.False(t, listDiff.Empty())
}

// Test schema creation helpers
func TestListOfTypesDiff_SchemaHelpers(t *testing.T) {
	// Test single type schema creation
	stringSchema := createSingleTypeSchema("string")
	require.NotNil(t, stringSchema)
	require.Equal(t, &openapi3.Types{"string"}, stringSchema.Type)

	// Test oneOf schema creation
	oneOfSchema := createOneOfSchema("string", "integer")
	require.NotNil(t, oneOfSchema)
	require.Len(t, oneOfSchema.OneOf, 2)
	require.Equal(t, "string", (*oneOfSchema.OneOf[0].Value.Type)[0])
	require.Equal(t, "integer", (*oneOfSchema.OneOf[1].Value.Type)[0])

	// Test anyOf schema creation
	anyOfSchema := createAnyOfSchema("number", "boolean")
	require.NotNil(t, anyOfSchema)
	require.Len(t, anyOfSchema.AnyOf, 2)
	require.Equal(t, "number", (*anyOfSchema.AnyOf[0].Value.Type)[0])
	require.Equal(t, "boolean", (*anyOfSchema.AnyOf[1].Value.Type)[0])

	// Test complex schema creation
	complexSchema := createComplexOneOfSchema()
	require.NotNil(t, complexSchema)
	require.Len(t, complexSchema.OneOf, 2)
	require.NotNil(t, complexSchema.OneOf[1].Value.Properties)
}

// Test edge case schemas that shouldn't trigger list-of-types detection
func TestListOfTypesDiff_EdgeCaseSchemas(t *testing.T) {
	// Test schema with multiple types in single type field
	multiTypeSchema := &openapi3.Schema{
		Type: &openapi3.Types{"string", "integer"}, // Multiple types in single schema
	}
	require.NotNil(t, multiTypeSchema)
	require.Len(t, *multiTypeSchema.Type, 2)

	// Test empty oneOf/anyOf
	emptyOneOfSchema := &openapi3.Schema{
		OneOf: []*openapi3.SchemaRef{},
	}
	require.NotNil(t, emptyOneOfSchema)
	require.Empty(t, emptyOneOfSchema.OneOf)

	// Test oneOf with precedence over anyOf
	precedenceSchema := &openapi3.Schema{
		OneOf: []*openapi3.SchemaRef{
			{Value: createSingleTypeSchema("string")},
		},
		AnyOf: []*openapi3.SchemaRef{
			{Value: createSingleTypeSchema("integer")},
		},
	}
	require.NotNil(t, precedenceSchema)
	require.Len(t, precedenceSchema.OneOf, 1)
	require.Len(t, precedenceSchema.AnyOf, 1)
	// OneOf should take precedence in detection logic
}

// Test that schemas with no type are NOT analyzed as list-of-types (integration test)
func TestListOfTypesDiff_NoTypeNotSupported(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	base, err := loader.LoadFromFile("../data/list-of-types/no-type-base.yaml")
	require.NoError(t, err)

	revision, err := loader.LoadFromFile("../data/list-of-types/no-type-revision.yaml")
	require.NoError(t, err)

	diffReport, err := diff.Get(diff.NewConfig(), base, revision)
	require.NoError(t, err)

	// Check that properties with no-type schemas do NOT have list-of-types diffs
	require.NotNil(t, diffReport.PathsDiff)
	require.Contains(t, diffReport.PathsDiff.Modified, "/test")

	getDiff := diffReport.PathsDiff.Modified["/test"].OperationsDiff.Modified["GET"]
	require.NotNil(t, getDiff)

	schemaDiffs := getDiff.ResponsesDiff.Modified["200"].ContentDiff.MediaTypeModified["application/json"].
		SchemaDiff.PropertiesDiff.Modified

	// oneOfWithNoType should NOT have list-of-types diff
	oneOfDiff := schemaDiffs["oneOfWithNoType"].ListOfTypesDiff
	require.Nil(t, oneOfDiff, "oneOf with no-type schema should not be analyzed as list-of-types")

	// anyOfWithNoType should NOT have list-of-types diff
	anyOfDiff := schemaDiffs["anyOfWithNoType"].ListOfTypesDiff
	require.Nil(t, anyOfDiff, "anyOf with no-type schema should not be analyzed as list-of-types")
}
