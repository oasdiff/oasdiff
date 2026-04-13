package checker

import (
	"github.com/oasdiff/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
)

// displayFilePath strips any git revision prefix from a file path.
// e.g. "HEAD:openapi.yaml" → "openapi.yaml", "openapi.yaml" → "openapi.yaml"
func displayFilePath(file string) string {
	return load.NewSource(file).DisplayPath()
}

// Source represents the location of a change in an OpenAPI spec file
type Source struct {
	File   string `json:"file,omitempty" yaml:"file,omitempty"`     // File path (relative or absolute)
	Line   int    `json:"line,omitempty" yaml:"line,omitempty"`     // Line number (1-based)
	Column int    `json:"column,omitempty" yaml:"column,omitempty"` // Column number (1-based)
}

func NewSource(file string, line int, column int) *Source {
	return &Source{
		File:   file,
		Line:   line,
		Column: column,
	}
}
func NewSourceFromOrigin(operationsSources *diff.OperationsSourcesMap, operation *openapi3.Operation, origin *openapi3.Origin) *Source {
	if origin == nil || origin.Key == nil {
		return &Source{File: (*operationsSources)[operation]}
	}

	file := displayFilePath(origin.Key.File)
	if file == "" {
		file = (*operationsSources)[operation]
	}

	return &Source{
		File:   file,
		Line:   origin.Key.Line,
		Column: origin.Key.Column,
	}
}

func NewSourceFromField(operationsSources *diff.OperationsSourcesMap, operation *openapi3.Operation, origin *openapi3.Origin, field string) *Source {
	if origin == nil {
		return nil
	}

	if location, ok := origin.Fields[field]; ok {
		file := displayFilePath(location.File)
		if file == "" {
			file = (*operationsSources)[operation]
		}
		return &Source{
			File:   file,
			Line:   location.Line,
			Column: location.Column,
		}
	}

	return nil
}

// operationSources returns the base and revision source locations for a modified operation pair.
// Returns nil, nil when origin tracking is not enabled (no useful line/column data available).
func operationSources(operationsSources *diff.OperationsSourcesMap, base, revision *openapi3.Operation) (*Source, *Source) {
	// Only populate sources when at least one operation has origin data
	hasOrigin := (base != nil && base.Origin != nil) || (revision != nil && revision.Origin != nil)
	if !hasOrigin {
		return nil, nil
	}

	var baseSource, revisionSource *Source
	if base != nil {
		baseSource = NewSourceFromOrigin(operationsSources, base, base.Origin)
	}
	if revision != nil {
		revisionSource = NewSourceFromOrigin(operationsSources, revision, revision.Origin)
	}
	return baseSource, revisionSource
}

// NewSourceFromSequenceItem returns the source location of a specific item
// in a sequence-valued field (e.g. type: [string, null]).
// It looks up the item by value in Origin.Sequences[field].
// If multiple items share the same value, it returns the first match.
func NewSourceFromSequenceItem(operationsSources *diff.OperationsSourcesMap, operation *openapi3.Operation, origin *openapi3.Origin, field string, value string) *Source {
	if origin == nil || origin.Sequences == nil {
		return nil
	}

	items, ok := origin.Sequences[field]
	if !ok {
		return nil
	}

	for _, item := range items {
		if item.Name == value {
			file := displayFilePath(item.File)
			if file == "" {
				file = (*operationsSources)[operation]
			}
			return &Source{
				File:   file,
				Line:   item.Line,
				Column: item.Column,
			}
		}
	}

	return nil
}

// OperationFieldSources returns source locations from a specific field within operation Origins.
// Falls back to operation-level sources when the field is not found in origin data.
func OperationFieldSources(operationsSources *diff.OperationsSourcesMap, operationItem *diff.MethodDiff, field string) (*Source, *Source) {
	hasOrigin := (operationItem.Base != nil && operationItem.Base.Origin != nil) ||
		(operationItem.Revision != nil && operationItem.Revision.Origin != nil)
	if !hasOrigin {
		return nil, nil
	}

	var baseSource, revisionSource *Source
	if operationItem.Base != nil {
		baseSource = NewSourceFromField(operationsSources, operationItem.Base, operationItem.Base.Origin, field)
	}
	if operationItem.Revision != nil {
		revisionSource = NewSourceFromField(operationsSources, operationItem.Revision, operationItem.Revision.Origin, field)
	}
	return baseSource, revisionSource
}

// ParameterFieldSources returns source locations from a specific field within parameter Origins.
// Falls back to parameter-level sources when the field is not found in origin data.
func ParameterFieldSources(operationsSources *diff.OperationsSourcesMap, operationItem *diff.MethodDiff, paramDiff *diff.ParameterDiff, field string) (*Source, *Source) {
	if paramDiff == nil {
		return operationSources(operationsSources, operationItem.Base, operationItem.Revision)
	}

	hasOrigin := (paramDiff.Base != nil && paramDiff.Base.Origin != nil) ||
		(paramDiff.Revision != nil && paramDiff.Revision.Origin != nil)
	if !hasOrigin {
		return operationSources(operationsSources, operationItem.Base, operationItem.Revision)
	}

	var baseSource, revisionSource *Source
	if paramDiff.Base != nil && operationItem.Base != nil {
		baseSource = NewSourceFromField(operationsSources, operationItem.Base, paramDiff.Base.Origin, field)
	}
	if paramDiff.Revision != nil && operationItem.Revision != nil {
		revisionSource = NewSourceFromField(operationsSources, operationItem.Revision, paramDiff.Revision.Origin, field)
	}
	return baseSource, revisionSource
}

// schemaSources returns source locations from schema Origins.
// Falls back to operation-level sources when schema origin data is unavailable.
func SchemaSources(operationsSources *diff.OperationsSourcesMap, operationItem *diff.MethodDiff, schemaDiff *diff.SchemaDiff) (*Source, *Source) {
	if schemaDiff == nil {
		return operationSources(operationsSources, operationItem.Base, operationItem.Revision)
	}

	hasOrigin := (schemaDiff.Base != nil && schemaDiff.Base.Origin != nil) ||
		(schemaDiff.Revision != nil && schemaDiff.Revision.Origin != nil)
	if !hasOrigin {
		return operationSources(operationsSources, operationItem.Base, operationItem.Revision)
	}

	var baseSource, revisionSource *Source
	if schemaDiff.Base != nil && operationItem.Base != nil {
		baseSource = NewSourceFromOrigin(operationsSources, operationItem.Base, schemaDiff.Base.Origin)
	}
	if schemaDiff.Revision != nil && operationItem.Revision != nil {
		revisionSource = NewSourceFromOrigin(operationsSources, operationItem.Revision, schemaDiff.Revision.Origin)
	}
	return baseSource, revisionSource
}

// parameterSources returns source locations from parameter Origins.
// Falls back to operation-level sources when parameter origin data is unavailable.
func ParameterSources(operationsSources *diff.OperationsSourcesMap, operationItem *diff.MethodDiff, paramDiff *diff.ParameterDiff) (*Source, *Source) {
	if paramDiff == nil {
		return operationSources(operationsSources, operationItem.Base, operationItem.Revision)
	}

	hasOrigin := (paramDiff.Base != nil && paramDiff.Base.Origin != nil) ||
		(paramDiff.Revision != nil && paramDiff.Revision.Origin != nil)
	if !hasOrigin {
		return operationSources(operationsSources, operationItem.Base, operationItem.Revision)
	}

	var baseSource, revisionSource *Source
	if paramDiff.Base != nil && operationItem.Base != nil {
		baseSource = NewSourceFromOrigin(operationsSources, operationItem.Base, paramDiff.Base.Origin)
	}
	if paramDiff.Revision != nil && operationItem.Revision != nil {
		revisionSource = NewSourceFromOrigin(operationsSources, operationItem.Revision, paramDiff.Revision.Origin)
	}
	return baseSource, revisionSource
}

// responseSources returns source locations from response Origins.
// Falls back to operation-level sources when response origin data is unavailable.
func ResponseSources(operationsSources *diff.OperationsSourcesMap, operationItem *diff.MethodDiff, responseDiff *diff.ResponseDiff) (*Source, *Source) {
	if responseDiff == nil {
		return operationSources(operationsSources, operationItem.Base, operationItem.Revision)
	}

	hasOrigin := (responseDiff.Base != nil && responseDiff.Base.Origin != nil) ||
		(responseDiff.Revision != nil && responseDiff.Revision.Origin != nil)
	if !hasOrigin {
		return operationSources(operationsSources, operationItem.Base, operationItem.Revision)
	}

	var baseSource, revisionSource *Source
	if responseDiff.Base != nil && operationItem.Base != nil {
		baseSource = NewSourceFromOrigin(operationsSources, operationItem.Base, responseDiff.Base.Origin)
	}
	if responseDiff.Revision != nil && operationItem.Revision != nil {
		revisionSource = NewSourceFromOrigin(operationsSources, operationItem.Revision, responseDiff.Revision.Origin)
	}
	return baseSource, revisionSource
}

// schemaFieldSources returns source locations from a specific field within schema Origins.
// Falls back to schema-level sources when the field is not found in origin data.
func SchemaFieldSources(operationsSources *diff.OperationsSourcesMap, operationItem *diff.MethodDiff, schemaDiff *diff.SchemaDiff, field string) (*Source, *Source) {
	if schemaDiff == nil {
		return operationSources(operationsSources, operationItem.Base, operationItem.Revision)
	}

	hasOrigin := (schemaDiff.Base != nil && schemaDiff.Base.Origin != nil) ||
		(schemaDiff.Revision != nil && schemaDiff.Revision.Origin != nil)
	if !hasOrigin {
		return operationSources(operationsSources, operationItem.Base, operationItem.Revision)
	}

	var baseSource, revisionSource *Source
	if schemaDiff.Base != nil && operationItem.Base != nil {
		baseSource = NewSourceFromField(operationsSources, operationItem.Base, schemaDiff.Base.Origin, field)
	}
	if schemaDiff.Revision != nil && operationItem.Revision != nil {
		revisionSource = NewSourceFromField(operationsSources, operationItem.Revision, schemaDiff.Revision.Origin, field)
	}
	return baseSource, revisionSource
}

// SchemaDeletedItemSources returns source locations for a deleted sequence item.
// Base source points to the specific deleted item, revision source is nil.
func SchemaDeletedItemSources(operationsSources *diff.OperationsSourcesMap, operationItem *diff.MethodDiff, schemaDiff *diff.SchemaDiff, field string, value string) (*Source, *Source) {
	if schemaDiff == nil || schemaDiff.Base == nil || operationItem.Base == nil {
		return nil, nil
	}
	baseSource := NewSourceFromSequenceItem(operationsSources, operationItem.Base, schemaDiff.Base.Origin, field, value)
	return baseSource, nil
}

// SchemaAddedItemSources returns source locations for an added sequence item.
// Base source is nil, revision source points to the specific added item.
func SchemaAddedItemSources(operationsSources *diff.OperationsSourcesMap, operationItem *diff.MethodDiff, schemaDiff *diff.SchemaDiff, field string, value string) (*Source, *Source) {
	if schemaDiff == nil || schemaDiff.Revision == nil || operationItem.Revision == nil {
		return nil, nil
	}
	revisionSource := NewSourceFromSequenceItem(operationsSources, operationItem.Revision, schemaDiff.Revision.Origin, field, value)
	return nil, revisionSource
}

// SubschemaSources returns source locations for a specific subschema within an allOf/oneOf/anyOf array.
// For added subschemas, baseIndex should be -1; for deleted subschemas, revisionIndex should be -1.
func SubschemaSources(operationsSources *diff.OperationsSourcesMap, operationItem *diff.MethodDiff, schemaDiff *diff.SchemaDiff, field string, baseIndex, revisionIndex int) (*Source, *Source) {
	if schemaDiff == nil {
		return operationSources(operationsSources, operationItem.Base, operationItem.Revision)
	}

	var baseSource, revisionSource *Source

	if baseIndex >= 0 && schemaDiff.Base != nil && operationItem.Base != nil {
		baseSource = subschemaSource(operationsSources, operationItem.Base, schemaDiff.Base, field, baseIndex)
	}
	if revisionIndex >= 0 && schemaDiff.Revision != nil && operationItem.Revision != nil {
		revisionSource = subschemaSource(operationsSources, operationItem.Revision, schemaDiff.Revision, field, revisionIndex)
	}

	// Fall back to field-level source if subschema origin is not available
	if baseSource == nil && revisionSource == nil {
		return SchemaFieldSources(operationsSources, operationItem, schemaDiff, field)
	}

	return baseSource, revisionSource
}

// subschemaSource extracts the source location from a specific subschema in an allOf/oneOf/anyOf array.
func subschemaSource(operationsSources *diff.OperationsSourcesMap, operation *openapi3.Operation, schema *openapi3.Schema, field string, index int) *Source {
	var refs openapi3.SchemaRefs
	switch field {
	case "allOf":
		refs = schema.AllOf
	case "oneOf":
		refs = schema.OneOf
	case "anyOf":
		refs = schema.AnyOf
	default:
		return nil
	}

	if index >= len(refs) || refs[index] == nil {
		return nil
	}

	// Prefer the SchemaRef origin (points to the $ref line or inline schema in the array)
	// over the resolved Value origin (points to the component definition for $ref schemas)
	origin := refs[index].Origin
	if origin == nil || origin.Key == nil {
		if refs[index].Value == nil {
			return nil
		}
		origin = refs[index].Value.Origin
		if origin == nil || origin.Key == nil {
			return nil
		}
	}

	file := displayFilePath(origin.Key.File)
	if file == "" {
		file = (*operationsSources)[operation]
	}

	return &Source{
		File:   file,
		Line:   origin.Key.Line,
		Column: origin.Key.Column,
	}
}

// parameterSource returns the source location of a specific parameter.
// Returns nil when the parameter has no origin data.
func parameterSource(operationsSources *diff.OperationsSourcesMap, op *openapi3.Operation, param *openapi3.Parameter) *Source {
	if op == nil || param == nil || param.Origin == nil {
		return nil
	}
	return NewSourceFromOrigin(operationsSources, op, param.Origin)
}

// requestBodyFieldSources returns source locations from a specific field within request body Origins.
func requestBodyFieldSources(operationsSources *diff.OperationsSourcesMap, operationItem *diff.MethodDiff, field string) (*Source, *Source) {
	var baseOrigin, revisionOrigin *openapi3.Origin
	if operationItem.Base != nil && operationItem.Base.RequestBody != nil && operationItem.Base.RequestBody.Value != nil {
		baseOrigin = operationItem.Base.RequestBody.Value.Origin
	}
	if operationItem.Revision != nil && operationItem.Revision.RequestBody != nil && operationItem.Revision.RequestBody.Value != nil {
		revisionOrigin = operationItem.Revision.RequestBody.Value.Origin
	}

	if baseOrigin == nil && revisionOrigin == nil {
		return operationSources(operationsSources, operationItem.Base, operationItem.Revision)
	}

	var baseSource, revisionSource *Source
	if operationItem.Base != nil {
		baseSource = NewSourceFromField(operationsSources, operationItem.Base, baseOrigin, field)
	}
	if operationItem.Revision != nil {
		revisionSource = NewSourceFromField(operationsSources, operationItem.Revision, revisionOrigin, field)
	}
	return baseSource, revisionSource
}

// requestBodySource returns the source location of the request body of an operation.
// Returns nil when the request body has no origin data.
func requestBodySource(operationsSources *diff.OperationsSourcesMap, op *openapi3.Operation) *Source {
	if op == nil || op.RequestBody == nil || op.RequestBody.Value == nil || op.RequestBody.Value.Origin == nil {
		return nil
	}
	return NewSourceFromOrigin(operationsSources, op, op.RequestBody.Value.Origin)
}

// requestBodyMediaTypeSource returns the source location of a specific media type within a request body.
// Returns nil when the media type has no origin data.
func requestBodyMediaTypeSource(operationsSources *diff.OperationsSourcesMap, op *openapi3.Operation, mediaType string) *Source {
	if op == nil || op.RequestBody == nil || op.RequestBody.Value == nil || op.RequestBody.Value.Content == nil {
		return nil
	}
	if mt := op.RequestBody.Value.Content[mediaType]; mt != nil && mt.Origin != nil {
		return NewSourceFromOrigin(operationsSources, op, mt.Origin)
	}
	return nil
}

// mediaTypeSource returns the source location of a specific media type within a response.
// Returns nil when the media type has no origin data.
func mediaTypeSource(operationsSources *diff.OperationsSourcesMap, op *openapi3.Operation, response *openapi3.Response, mediaType string) *Source {
	if op == nil || response == nil || response.Content == nil {
		return nil
	}
	if mt := response.Content[mediaType]; mt != nil && mt.Origin != nil {
		return NewSourceFromOrigin(operationsSources, op, mt.Origin)
	}
	return nil
}

// responseMediaTypeNameSources returns source locations for a renamed media type within a response.
// It points to the old media-type line in the base and the new media-type line in the revision.
// Falls back to response-level sources when media type origin data is unavailable.
func responseMediaTypeNameSources(operationsSources *diff.OperationsSourcesMap, operationItem *diff.MethodDiff, responseDiff *diff.ResponseDiff, fromMediaType, toMediaType string) (*Source, *Source) {
	baseSource := mediaTypeSource(operationsSources, operationItem.Base, responseDiff.Base, fromMediaType)
	revisionSource := mediaTypeSource(operationsSources, operationItem.Revision, responseDiff.Revision, toMediaType)
	if baseSource == nil && revisionSource == nil {
		return ResponseSources(operationsSources, operationItem, responseDiff)
	}
	return baseSource, revisionSource
}

// headerSources returns source locations from the base and revision headers within a response.
// Falls back to response-level sources when header origin data is unavailable.
func headerSources(operationsSources *diff.OperationsSourcesMap, operationItem *diff.MethodDiff, responseDiff *diff.ResponseDiff, headerName string) (*Source, *Source) {
	var baseOrigin, revisionOrigin *openapi3.Origin
	if responseDiff.Base != nil {
		if h := responseDiff.Base.Headers[headerName]; h != nil && h.Value != nil {
			baseOrigin = h.Value.Origin
		}
	}
	if responseDiff.Revision != nil {
		if h := responseDiff.Revision.Headers[headerName]; h != nil && h.Value != nil {
			revisionOrigin = h.Value.Origin
		}
	}

	hasOrigin := baseOrigin != nil || revisionOrigin != nil
	if !hasOrigin {
		return ResponseSources(operationsSources, operationItem, responseDiff)
	}

	var baseSource, revisionSource *Source
	if operationItem.Base != nil {
		baseSource = NewSourceFromOrigin(operationsSources, operationItem.Base, baseOrigin)
	}
	if operationItem.Revision != nil {
		revisionSource = NewSourceFromOrigin(operationsSources, operationItem.Revision, revisionOrigin)
	}
	return baseSource, revisionSource
}

// propertySource returns the source location of a specific property schema.
// Returns nil when the schema has no origin data.
func propertySource(operationsSources *diff.OperationsSourcesMap, op *openapi3.Operation, schema *openapi3.Schema) *Source {
	if op == nil || schema == nil || schema.Origin == nil {
		return nil
	}
	return NewSourceFromOrigin(operationsSources, op, schema.Origin)
}

// sourceFromOrigin creates a Source directly from an Origin.
// Used for component-level changes where no operation context is available.
func sourceFromOrigin(origin *openapi3.Origin) *Source {
	if origin == nil || origin.Key == nil {
		return nil
	}
	return &Source{
		File:   displayFilePath(origin.Key.File),
		Line:   origin.Key.Line,
		Column: origin.Key.Column,
	}
}

func NewEmptySource() *Source {
	return nil
}
