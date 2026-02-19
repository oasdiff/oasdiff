package checker

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

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

	file := origin.Key.File
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
		return &Source{File: (*operationsSources)[operation]}
	}

	location, ok := origin.Fields[field]
	if !ok {
		return &Source{File: (*operationsSources)[operation]}
	}

	file := location.File
	if file == "" {
		file = (*operationsSources)[operation]
	}

	return &Source{
		File:   file,
		Line:   location.Line,
		Column: location.Column,
	}
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

// operationSource returns the source location for a single operation.
// Returns nil when the operation is nil or has no origin data.
func operationSource(operationsSources *diff.OperationsSourcesMap, op *openapi3.Operation) *Source {
	if op == nil || op.Origin == nil {
		return nil
	}
	return NewSourceFromOrigin(operationsSources, op, op.Origin)
}

// NewSourceFromSequenceItem returns the source location of a specific item
// in a sequence-valued field (e.g. type: [string, null]).
// It looks up the item by value in Origin.Sequences[field].
// If multiple items share the same value, it returns the first match.
func NewSourceFromSequenceItem(operationsSources *diff.OperationsSourcesMap, operation *openapi3.Operation, origin *openapi3.Origin, field string, value string) *Source {
	if origin == nil || origin.Sequences == nil {
		return &Source{File: (*operationsSources)[operation]}
	}

	items, ok := origin.Sequences[field]
	if !ok {
		return &Source{File: (*operationsSources)[operation]}
	}

	for _, item := range items {
		if item.Name == value {
			file := item.File
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

	return &Source{File: (*operationsSources)[operation]}
}

func NewEmptySource() *Source {
	return nil
}
