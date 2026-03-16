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
