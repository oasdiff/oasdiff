package checker_test

import (
	"testing"

	"github.com/oasdiff/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

func TestApiChange_SourceFile(t *testing.T) {
	apiChangeSourceFile := apiChange
	apiChangeSourceFile.SourceFile = ""
	apiChangeSourceFile.Source = load.NewSource("spec.yaml")

	require.Equal(t, "spec.yaml", apiChangeSourceFile.GetSourceFile())
}

func TestApiChange_SourceUrl(t *testing.T) {
	apiChangeSourceFile := apiChange
	apiChangeSourceFile.SourceFile = ""
	apiChangeSourceFile.Source = load.NewSource("http://google.com/spec.yaml")

	require.Equal(t, "", apiChangeSourceFile.GetSourceFile())
}

func TestApiChangeWithSources_DirectConstruction(t *testing.T) {
	// Test direct construction of ApiChange with BaseSource and RevisionSource
	baseSource := checker.NewSource("base.yaml", 10, 5)
	revisionSource := checker.NewSource("revision.yaml", 12, 7)

	change := checker.ApiChange{
		Id:        "test-id",
		Args:      []any{"arg1"},
		Comment:   "test comment",
		Level:     checker.INFO,
		Operation: "GET",
		Path:      "/test",
		CommonChange: checker.CommonChange{
			BaseSource:     baseSource,
			RevisionSource: revisionSource,
		},
	}

	// Test that the new fields are set correctly
	require.Equal(t, baseSource, change.GetBaseSource())
	require.Equal(t, revisionSource, change.GetRevisionSource())
	require.NotEmpty(t, change.GetBaseSource())
	require.NotEmpty(t, change.GetRevisionSource())
	require.Equal(t, "base.yaml", change.GetBaseSource().File)
	require.Equal(t, 10, change.GetBaseSource().Line)
	require.Equal(t, 5, change.GetBaseSource().Column)
	require.Equal(t, "revision.yaml", change.GetRevisionSource().File)
	require.Equal(t, 12, change.GetRevisionSource().Line)
	require.Equal(t, 7, change.GetRevisionSource().Column)
}

func TestSchemaSources(t *testing.T) {
	baseOp := &openapi3.Operation{}
	revisionOp := &openapi3.Operation{
		Origin: &openapi3.Origin{
			Key: &openapi3.Location{File: "spec.yaml", Line: 1, Column: 1},
		},
	}
	sources := diff.OperationsSourcesMap{baseOp: "base.yaml", revisionOp: "revision.yaml"}
	operationItem := &diff.MethodDiff{Base: baseOp, Revision: revisionOp}

	baseSchema := &openapi3.Schema{
		Origin: &openapi3.Origin{
			Key: &openapi3.Location{File: "base.yaml", Line: 10, Column: 5},
		},
	}
	revisionSchema := &openapi3.Schema{
		Origin: &openapi3.Origin{
			Key: &openapi3.Location{File: "revision.yaml", Line: 15, Column: 7},
		},
	}

	schemaDiff := &diff.SchemaDiff{Base: baseSchema, Revision: revisionSchema}

	baseSource, revisionSource := checker.SchemaSources(&sources, operationItem, schemaDiff)
	require.Equal(t, "base.yaml", baseSource.File)
	require.Equal(t, 10, baseSource.Line)
	require.Equal(t, 5, baseSource.Column)
	require.Equal(t, "revision.yaml", revisionSource.File)
	require.Equal(t, 15, revisionSource.Line)
	require.Equal(t, 7, revisionSource.Column)
}

func TestSchemaSources_NoOrigin_FallsBackToOperation(t *testing.T) {
	baseOp := &openapi3.Operation{
		Origin: &openapi3.Origin{
			Key: &openapi3.Location{File: "base.yaml", Line: 3, Column: 1},
		},
	}
	revisionOp := &openapi3.Operation{
		Origin: &openapi3.Origin{
			Key: &openapi3.Location{File: "revision.yaml", Line: 5, Column: 1},
		},
	}
	sources := diff.OperationsSourcesMap{baseOp: "base.yaml", revisionOp: "revision.yaml"}
	operationItem := &diff.MethodDiff{Base: baseOp, Revision: revisionOp}

	// Schema without origin data
	schemaDiff := &diff.SchemaDiff{Base: &openapi3.Schema{}, Revision: &openapi3.Schema{}}

	baseSource, revisionSource := checker.SchemaSources(&sources, operationItem, schemaDiff)
	require.Equal(t, "base.yaml", baseSource.File)
	require.Equal(t, 3, baseSource.Line)
	require.Equal(t, "revision.yaml", revisionSource.File)
	require.Equal(t, 5, revisionSource.Line)
}

func TestSchemaSources_NilDiff_FallsBackToOperation(t *testing.T) {
	baseOp := &openapi3.Operation{
		Origin: &openapi3.Origin{
			Key: &openapi3.Location{File: "base.yaml", Line: 3, Column: 1},
		},
	}
	revisionOp := &openapi3.Operation{
		Origin: &openapi3.Origin{
			Key: &openapi3.Location{File: "revision.yaml", Line: 5, Column: 1},
		},
	}
	sources := diff.OperationsSourcesMap{baseOp: "base.yaml", revisionOp: "revision.yaml"}
	operationItem := &diff.MethodDiff{Base: baseOp, Revision: revisionOp}

	baseSource, revisionSource := checker.SchemaSources(&sources, operationItem, nil)
	require.Equal(t, "base.yaml", baseSource.File)
	require.Equal(t, 3, baseSource.Line)
	require.Equal(t, "revision.yaml", revisionSource.File)
	require.Equal(t, 5, revisionSource.Line)
}

func TestParameterSources(t *testing.T) {
	baseOp := &openapi3.Operation{}
	revisionOp := &openapi3.Operation{
		Origin: &openapi3.Origin{
			Key: &openapi3.Location{File: "spec.yaml", Line: 1, Column: 1},
		},
	}
	sources := diff.OperationsSourcesMap{baseOp: "base.yaml", revisionOp: "revision.yaml"}
	operationItem := &diff.MethodDiff{Base: baseOp, Revision: revisionOp}

	baseParam := &openapi3.Parameter{
		Origin: &openapi3.Origin{
			Key: &openapi3.Location{File: "base.yaml", Line: 20, Column: 9},
		},
	}
	revisionParam := &openapi3.Parameter{
		Origin: &openapi3.Origin{
			Key: &openapi3.Location{File: "revision.yaml", Line: 25, Column: 9},
		},
	}

	paramDiff := &diff.ParameterDiff{Base: baseParam, Revision: revisionParam}

	baseSource, revisionSource := checker.ParameterSources(&sources, operationItem, paramDiff)
	require.Equal(t, "base.yaml", baseSource.File)
	require.Equal(t, 20, baseSource.Line)
	require.Equal(t, 9, baseSource.Column)
	require.Equal(t, "revision.yaml", revisionSource.File)
	require.Equal(t, 25, revisionSource.Line)
}

func TestResponseSources(t *testing.T) {
	baseOp := &openapi3.Operation{}
	revisionOp := &openapi3.Operation{
		Origin: &openapi3.Origin{
			Key: &openapi3.Location{File: "spec.yaml", Line: 1, Column: 1},
		},
	}
	sources := diff.OperationsSourcesMap{baseOp: "base.yaml", revisionOp: "revision.yaml"}
	operationItem := &diff.MethodDiff{Base: baseOp, Revision: revisionOp}

	baseResp := &openapi3.Response{
		Origin: &openapi3.Origin{
			Key: &openapi3.Location{File: "base.yaml", Line: 30, Column: 7},
		},
	}
	revisionResp := &openapi3.Response{
		Origin: &openapi3.Origin{
			Key: &openapi3.Location{File: "revision.yaml", Line: 35, Column: 7},
		},
	}

	respDiff := &diff.ResponseDiff{Base: baseResp, Revision: revisionResp}

	baseSource, revisionSource := checker.ResponseSources(&sources, operationItem, respDiff)
	require.Equal(t, "base.yaml", baseSource.File)
	require.Equal(t, 30, baseSource.Line)
	require.Equal(t, 7, baseSource.Column)
	require.Equal(t, "revision.yaml", revisionSource.File)
	require.Equal(t, 35, revisionSource.Line)
}

func TestSchemaFieldSources(t *testing.T) {
	baseOp := &openapi3.Operation{}
	revisionOp := &openapi3.Operation{
		Origin: &openapi3.Origin{
			Key: &openapi3.Location{File: "spec.yaml", Line: 1, Column: 1},
		},
	}
	sources := diff.OperationsSourcesMap{baseOp: "base.yaml", revisionOp: "revision.yaml"}
	operationItem := &diff.MethodDiff{Base: baseOp, Revision: revisionOp}

	baseSchema := &openapi3.Schema{
		Origin: &openapi3.Origin{
			Key:    &openapi3.Location{File: "base.yaml", Line: 10, Column: 5},
			Fields: map[string]openapi3.Location{"type": {File: "base.yaml", Line: 11, Column: 7}},
		},
	}
	revisionSchema := &openapi3.Schema{
		Origin: &openapi3.Origin{
			Key:    &openapi3.Location{File: "revision.yaml", Line: 15, Column: 5},
			Fields: map[string]openapi3.Location{"type": {File: "revision.yaml", Line: 16, Column: 7}},
		},
	}

	schemaDiff := &diff.SchemaDiff{Base: baseSchema, Revision: revisionSchema}

	// Field-level precision
	baseSource, revisionSource := checker.SchemaFieldSources(&sources, operationItem, schemaDiff, "type")
	require.Equal(t, "base.yaml", baseSource.File)
	require.Equal(t, 11, baseSource.Line)
	require.Equal(t, 7, baseSource.Column)
	require.Equal(t, "revision.yaml", revisionSource.File)
	require.Equal(t, 16, revisionSource.Line)
	require.Equal(t, 7, revisionSource.Column)

	// Missing field returns nil (field doesn't exist in YAML)
	baseSource, revisionSource = checker.SchemaFieldSources(&sources, operationItem, schemaDiff, "format")
	require.Nil(t, baseSource)
	require.Nil(t, revisionSource)
}

func TestNewSourceFromOrigin_StripsGitRevisionPrefix(t *testing.T) {
	op := &openapi3.Operation{}
	sources := diff.OperationsSourcesMap{op: "openapi.yaml"}
	origin := &openapi3.Origin{
		Key: &openapi3.Location{File: "HEAD:openapi.yaml", Line: 10, Column: 5},
	}

	source := checker.NewSourceFromOrigin(&sources, op, origin)
	require.Equal(t, "openapi.yaml", source.File)
	require.Equal(t, 10, source.Line)
	require.Equal(t, 5, source.Column)
}

func TestNewSourceFromField_StripsGitRevisionPrefix(t *testing.T) {
	op := &openapi3.Operation{}
	sources := diff.OperationsSourcesMap{op: "openapi.yaml"}
	origin := &openapi3.Origin{
		Key: &openapi3.Location{File: "HEAD:openapi.yaml", Line: 1, Column: 1},
		Fields: map[string]openapi3.Location{
			"pattern": {File: "origin/main:openapi.yaml", Line: 15, Column: 14},
		},
	}

	source := checker.NewSourceFromField(&sources, op, origin, "pattern")
	require.Equal(t, "openapi.yaml", source.File)
	require.Equal(t, 15, source.Line)
	require.Equal(t, 14, source.Column)
}

func TestNewSourceFromSequenceItem_StripsGitRevisionPrefix(t *testing.T) {
	op := &openapi3.Operation{}
	sources := diff.OperationsSourcesMap{op: "openapi.yaml"}
	origin := &openapi3.Origin{
		Sequences: map[string][]openapi3.Location{
			"type": {{File: "HEAD:openapi.yaml", Line: 4, Column: 11, Name: "string"}},
		},
	}

	source := checker.NewSourceFromSequenceItem(&sources, op, origin, "type", "string")
	require.Equal(t, "openapi.yaml", source.File)
	require.Equal(t, 4, source.Line)
	require.Equal(t, 11, source.Column)
}

func TestSubschemaSources_AllOf_Added_Inline(t *testing.T) {
	baseOp := &openapi3.Operation{}
	revisionOp := &openapi3.Operation{}
	sources := diff.OperationsSourcesMap{baseOp: "base.yaml", revisionOp: "revision.yaml"}
	operationItem := &diff.MethodDiff{Base: baseOp, Revision: revisionOp}

	// Inline subschemas: SchemaRef.Origin points to the inline definition
	revisionSchema := &openapi3.Schema{
		AllOf: openapi3.SchemaRefs{
			{Origin: &openapi3.Origin{Key: &openapi3.Location{File: "revision.yaml", Line: 15, Column: 17}}, Value: &openapi3.Schema{}},
			{Origin: &openapi3.Origin{Key: &openapi3.Location{File: "revision.yaml", Line: 19, Column: 17}}, Value: &openapi3.Schema{}},
			{Origin: &openapi3.Origin{Key: &openapi3.Location{File: "revision.yaml", Line: 23, Column: 17}}, Value: &openapi3.Schema{}},
		},
		Origin: &openapi3.Origin{
			Key:    &openapi3.Location{File: "revision.yaml", Line: 14, Column: 15},
			Fields: map[string]openapi3.Location{"allOf": {File: "revision.yaml", Line: 14, Column: 15}},
		},
	}
	baseSchema := &openapi3.Schema{
		AllOf: openapi3.SchemaRefs{
			{Origin: &openapi3.Origin{Key: &openapi3.Location{File: "base.yaml", Line: 15, Column: 17}}, Value: &openapi3.Schema{}},
			{Origin: &openapi3.Origin{Key: &openapi3.Location{File: "base.yaml", Line: 19, Column: 17}}, Value: &openapi3.Schema{}},
		},
		Origin: &openapi3.Origin{
			Key:    &openapi3.Location{File: "base.yaml", Line: 14, Column: 15},
			Fields: map[string]openapi3.Location{"allOf": {File: "base.yaml", Line: 14, Column: 15}},
		},
	}

	schemaDiff := &diff.SchemaDiff{Base: baseSchema, Revision: revisionSchema}

	// Added inline subschema at revision index 2 (baseIndex=-1)
	baseSource, revisionSource := checker.SubschemaSources(&sources, operationItem, schemaDiff, "allOf", -1, 2)
	require.Nil(t, baseSource)
	require.NotNil(t, revisionSource)
	require.Equal(t, "revision.yaml", revisionSource.File)
	require.Equal(t, 23, revisionSource.Line) // points to the 3rd subschema, not allOf keyword
	require.Equal(t, 17, revisionSource.Column)
}

func TestSubschemaSources_AllOf_Added_Ref(t *testing.T) {
	baseOp := &openapi3.Operation{}
	revisionOp := &openapi3.Operation{}
	sources := diff.OperationsSourcesMap{baseOp: "base.yaml", revisionOp: "revision.yaml"}
	operationItem := &diff.MethodDiff{Base: baseOp, Revision: revisionOp}

	// $ref subschemas: SchemaRef.Origin points to the $ref line in the allOf array;
	// Value.Origin points to the component definition — we want the $ref line
	revisionSchema := &openapi3.Schema{
		AllOf: openapi3.SchemaRefs{
			{
				Ref:    "#/components/schemas/Dog",
				Origin: &openapi3.Origin{Key: &openapi3.Location{File: "revision.yaml", Line: 15, Column: 11}},
				Value:  &openapi3.Schema{Origin: &openapi3.Origin{Key: &openapi3.Location{File: "revision.yaml", Line: 58, Column: 5}}},
			},
			{
				Ref:    "#/components/schemas/Cat",
				Origin: &openapi3.Origin{Key: &openapi3.Location{File: "revision.yaml", Line: 16, Column: 11}},
				Value:  &openapi3.Schema{Origin: &openapi3.Origin{Key: &openapi3.Location{File: "revision.yaml", Line: 62, Column: 5}}},
			},
			{
				Ref:    "#/components/schemas/Rabbit",
				Origin: &openapi3.Origin{Key: &openapi3.Location{File: "revision.yaml", Line: 17, Column: 11}},
				Value:  &openapi3.Schema{Origin: &openapi3.Origin{Key: &openapi3.Location{File: "revision.yaml", Line: 66, Column: 5}}},
			},
		},
		Origin: &openapi3.Origin{
			Key:    &openapi3.Location{File: "revision.yaml", Line: 14, Column: 15},
			Fields: map[string]openapi3.Location{"allOf": {File: "revision.yaml", Line: 14, Column: 15}},
		},
	}
	baseSchema := &openapi3.Schema{
		AllOf: openapi3.SchemaRefs{
			{
				Ref:    "#/components/schemas/Dog",
				Origin: &openapi3.Origin{Key: &openapi3.Location{File: "base.yaml", Line: 15, Column: 11}},
				Value:  &openapi3.Schema{Origin: &openapi3.Origin{Key: &openapi3.Location{File: "base.yaml", Line: 50, Column: 5}}},
			},
			{
				Ref:    "#/components/schemas/Cat",
				Origin: &openapi3.Origin{Key: &openapi3.Location{File: "base.yaml", Line: 16, Column: 11}},
				Value:  &openapi3.Schema{Origin: &openapi3.Origin{Key: &openapi3.Location{File: "base.yaml", Line: 54, Column: 5}}},
			},
		},
		Origin: &openapi3.Origin{
			Key:    &openapi3.Location{File: "base.yaml", Line: 14, Column: 15},
			Fields: map[string]openapi3.Location{"allOf": {File: "base.yaml", Line: 14, Column: 15}},
		},
	}

	schemaDiff := &diff.SchemaDiff{Base: baseSchema, Revision: revisionSchema}

	// Added $ref subschema at revision index 2: should point to line 17 ($ref line), NOT line 66 (component def)
	baseSource, revisionSource := checker.SubschemaSources(&sources, operationItem, schemaDiff, "allOf", -1, 2)
	require.Nil(t, baseSource)
	require.NotNil(t, revisionSource)
	require.Equal(t, "revision.yaml", revisionSource.File)
	require.Equal(t, 17, revisionSource.Line) // $ref line, not component definition line 66
	require.Equal(t, 11, revisionSource.Column)
}

func TestSubschemaSources_OneOf_Deleted(t *testing.T) {
	baseOp := &openapi3.Operation{}
	revisionOp := &openapi3.Operation{}
	sources := diff.OperationsSourcesMap{baseOp: "base.yaml", revisionOp: "revision.yaml"}
	operationItem := &diff.MethodDiff{Base: baseOp, Revision: revisionOp}

	// Base schema has 3 oneOf subschemas; the 3rd is deleted
	baseSchema := &openapi3.Schema{
		OneOf: openapi3.SchemaRefs{
			{Origin: &openapi3.Origin{Key: &openapi3.Location{File: "base.yaml", Line: 10, Column: 9}}, Value: &openapi3.Schema{}},
			{Origin: &openapi3.Origin{Key: &openapi3.Location{File: "base.yaml", Line: 14, Column: 9}}, Value: &openapi3.Schema{}},
			{Origin: &openapi3.Origin{Key: &openapi3.Location{File: "base.yaml", Line: 18, Column: 9}}, Value: &openapi3.Schema{}},
		},
		Origin: &openapi3.Origin{
			Key:    &openapi3.Location{File: "base.yaml", Line: 9, Column: 7},
			Fields: map[string]openapi3.Location{"oneOf": {File: "base.yaml", Line: 9, Column: 7}},
		},
	}
	revisionSchema := &openapi3.Schema{
		OneOf: openapi3.SchemaRefs{
			{Origin: &openapi3.Origin{Key: &openapi3.Location{File: "revision.yaml", Line: 10, Column: 9}}, Value: &openapi3.Schema{}},
			{Origin: &openapi3.Origin{Key: &openapi3.Location{File: "revision.yaml", Line: 14, Column: 9}}, Value: &openapi3.Schema{}},
		},
		Origin: &openapi3.Origin{
			Key:    &openapi3.Location{File: "revision.yaml", Line: 9, Column: 7},
			Fields: map[string]openapi3.Location{"oneOf": {File: "revision.yaml", Line: 9, Column: 7}},
		},
	}

	schemaDiff := &diff.SchemaDiff{Base: baseSchema, Revision: revisionSchema}

	// Deleted subschema at base index 2 (revisionIndex=-1)
	baseSource, revisionSource := checker.SubschemaSources(&sources, operationItem, schemaDiff, "oneOf", 2, -1)
	require.NotNil(t, baseSource)
	require.Nil(t, revisionSource)
	require.Equal(t, "base.yaml", baseSource.File)
	require.Equal(t, 18, baseSource.Line) // points to the 3rd subschema, not oneOf keyword
	require.Equal(t, 9, baseSource.Column)
}

func TestSubschemaSources_AnyOf_NoOrigin_FallsBack(t *testing.T) {
	baseOp := &openapi3.Operation{}
	revisionOp := &openapi3.Operation{}
	sources := diff.OperationsSourcesMap{baseOp: "base.yaml", revisionOp: "revision.yaml"}
	operationItem := &diff.MethodDiff{Base: baseOp, Revision: revisionOp}

	// Subschemas have no Origin — should fall back to field-level source
	revisionSchema := &openapi3.Schema{
		AnyOf: openapi3.SchemaRefs{
			{Value: &openapi3.Schema{}},
			{Value: &openapi3.Schema{}},
		},
		Origin: &openapi3.Origin{
			Key:    &openapi3.Location{File: "revision.yaml", Line: 9, Column: 7},
			Fields: map[string]openapi3.Location{"anyOf": {File: "revision.yaml", Line: 10, Column: 9}},
		},
	}
	baseSchema := &openapi3.Schema{
		AnyOf: openapi3.SchemaRefs{
			{Value: &openapi3.Schema{}},
		},
		Origin: &openapi3.Origin{
			Key:    &openapi3.Location{File: "base.yaml", Line: 9, Column: 7},
			Fields: map[string]openapi3.Location{"anyOf": {File: "base.yaml", Line: 10, Column: 9}},
		},
	}

	schemaDiff := &diff.SchemaDiff{Base: baseSchema, Revision: revisionSchema}

	// Subschema has no origin → falls back to field-level source
	baseSource, revisionSource := checker.SubschemaSources(&sources, operationItem, schemaDiff, "anyOf", -1, 1)
	// Fallback to field-level: both base and revision should have anyOf field source
	require.NotNil(t, baseSource)
	require.Equal(t, 10, baseSource.Line) // anyOf field line, not schema key line
	require.NotNil(t, revisionSource)
	require.Equal(t, 10, revisionSource.Line) // anyOf field line
}

func TestSubschemaSources_NilSchemaDiff(t *testing.T) {
	baseOp := &openapi3.Operation{
		Origin: &openapi3.Origin{Key: &openapi3.Location{File: "base.yaml", Line: 3, Column: 1}},
	}
	revisionOp := &openapi3.Operation{
		Origin: &openapi3.Origin{Key: &openapi3.Location{File: "revision.yaml", Line: 5, Column: 1}},
	}
	sources := diff.OperationsSourcesMap{baseOp: "base.yaml", revisionOp: "revision.yaml"}
	operationItem := &diff.MethodDiff{Base: baseOp, Revision: revisionOp}

	// Nil schemaDiff falls back to operation sources
	baseSource, revisionSource := checker.SubschemaSources(&sources, operationItem, nil, "allOf", -1, 0)
	require.Equal(t, "base.yaml", baseSource.File)
	require.Equal(t, 3, baseSource.Line)
	require.Equal(t, "revision.yaml", revisionSource.File)
	require.Equal(t, 5, revisionSource.Line)
}

func TestSubschemaSources_InvalidField(t *testing.T) {
	baseOp := &openapi3.Operation{}
	revisionOp := &openapi3.Operation{}
	sources := diff.OperationsSourcesMap{baseOp: "base.yaml", revisionOp: "revision.yaml"}
	operationItem := &diff.MethodDiff{Base: baseOp, Revision: revisionOp}

	revisionSchema := &openapi3.Schema{
		AllOf: openapi3.SchemaRefs{
			{Value: &openapi3.Schema{Origin: &openapi3.Origin{Key: &openapi3.Location{File: "revision.yaml", Line: 15, Column: 17}}}},
		},
		Origin: &openapi3.Origin{
			Key: &openapi3.Location{File: "revision.yaml", Line: 14, Column: 15},
		},
	}
	baseSchema := &openapi3.Schema{
		Origin: &openapi3.Origin{
			Key: &openapi3.Location{File: "base.yaml", Line: 14, Column: 15},
		},
	}

	schemaDiff := &diff.SchemaDiff{Base: baseSchema, Revision: revisionSchema}

	// Invalid field name returns nil from subschemaSource, falls back to field-level (also nil for invalid field)
	baseSource, revisionSource := checker.SubschemaSources(&sources, operationItem, schemaDiff, "invalid", -1, 0)
	require.Nil(t, baseSource)
	require.Nil(t, revisionSource)
}

func TestSubschemaSources_IndexOutOfRange(t *testing.T) {
	baseOp := &openapi3.Operation{}
	revisionOp := &openapi3.Operation{}
	sources := diff.OperationsSourcesMap{baseOp: "base.yaml", revisionOp: "revision.yaml"}
	operationItem := &diff.MethodDiff{Base: baseOp, Revision: revisionOp}

	revisionSchema := &openapi3.Schema{
		AllOf: openapi3.SchemaRefs{
			{Value: &openapi3.Schema{Origin: &openapi3.Origin{Key: &openapi3.Location{File: "revision.yaml", Line: 15, Column: 17}}}},
		},
		Origin: &openapi3.Origin{
			Key:    &openapi3.Location{File: "revision.yaml", Line: 14, Column: 15},
			Fields: map[string]openapi3.Location{"allOf": {File: "revision.yaml", Line: 14, Column: 15}},
		},
	}
	baseSchema := &openapi3.Schema{
		Origin: &openapi3.Origin{
			Key:    &openapi3.Location{File: "base.yaml", Line: 14, Column: 15},
			Fields: map[string]openapi3.Location{"allOf": {File: "base.yaml", Line: 14, Column: 15}},
		},
	}

	schemaDiff := &diff.SchemaDiff{Base: baseSchema, Revision: revisionSchema}

	// Index 99 is out of range → falls back to field-level source
	baseSource, revisionSource := checker.SubschemaSources(&sources, operationItem, schemaDiff, "allOf", -1, 99)
	require.NotNil(t, baseSource)
	require.Equal(t, 14, baseSource.Line)
	require.NotNil(t, revisionSource)
	require.Equal(t, 14, revisionSource.Line)
}

func TestNewSourceFromSequenceItem(t *testing.T) {
	op := &openapi3.Operation{}
	sources := diff.OperationsSourcesMap{op: "spec.yaml"}

	origin := &openapi3.Origin{
		Sequences: map[string][]openapi3.Location{
			"type": {
				{File: "spec.yaml", Line: 4, Column: 11, Name: "string"},
				{File: "spec.yaml", Line: 5, Column: 11, Name: "null"},
				{File: "spec.yaml", Line: 6, Column: 11, Name: "integer"},
			},
		},
	}

	// Lookup existing item
	source := checker.NewSourceFromSequenceItem(&sources, op, origin, "type", "null")
	require.Equal(t, "spec.yaml", source.File)
	require.Equal(t, 5, source.Line)
	require.Equal(t, 11, source.Column)

	// Lookup first item
	source = checker.NewSourceFromSequenceItem(&sources, op, origin, "type", "string")
	require.Equal(t, 4, source.Line)

	// Lookup missing item returns nil
	source = checker.NewSourceFromSequenceItem(&sources, op, origin, "type", "boolean")
	require.Nil(t, source)

	// Lookup missing field returns nil
	source = checker.NewSourceFromSequenceItem(&sources, op, origin, "enum", "foo")
	require.Nil(t, source)

	// Nil origin returns nil
	source = checker.NewSourceFromSequenceItem(&sources, op, nil, "type", "string")
	require.Nil(t, source)
}
