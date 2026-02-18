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

func NewEmptySource() *Source {
	return nil
}
