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
	file := (*operationsSources)[operation]

	if origin == nil || origin.Key == nil {
		return &Source{File: file}
	}

	line := origin.Key.Line
	column := origin.Key.Column
	return &Source{
		File:   file,
		Line:   line,
		Column: column,
	}
}

func NewEmptySource() *Source {
	return nil
}
