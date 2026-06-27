package checker

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func ptrSchema(s *openapi3.Schema) *openapi3.SchemaRef {
	return &openapi3.SchemaRef{Value: s}
}

// typeFormatString renders a schema's type for a human message: format is shown
// only when present (no "string/" noise), an array shows its item type, and a
// nullable / multi type is comma-joined.
func TestTypeFormatString(t *testing.T) {
	for _, tc := range []struct {
		name   string
		schema *openapi3.Schema
		want   string
	}{
		{"nil", nil, ""},
		{"type only", &openapi3.Schema{Type: &openapi3.Types{"string"}}, "string"},
		{"type and format", &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "date-time"}, "string/date-time"},
		{"format only", &openapi3.Schema{Format: "uuid"}, "uuid"},
		{"nullable scalar", &openapi3.Schema{Type: &openapi3.Types{"string", "null"}}, "string, null"},
		{"multi type", &openapi3.Schema{Type: &openapi3.Types{"string", "integer"}}, "string, integer"},
		{"array of string", &openapi3.Schema{Type: &openapi3.Types{"array"}, Items: ptrSchema(&openapi3.Schema{Type: &openapi3.Types{"string"}})}, "array<string>"},
		{"nullable array of string", &openapi3.Schema{Type: &openapi3.Types{"array", "null"}, Items: ptrSchema(&openapi3.Schema{Type: &openapi3.Types{"string"}})}, "array<string>, null"},
		{"array of formatted item", &openapi3.Schema{Type: &openapi3.Types{"array"}, Items: ptrSchema(&openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "date-time"})}, "array<string/date-time>"},
		{"array without items", &openapi3.Schema{Type: &openapi3.Types{"array"}}, "array"},
		{"no type no format", &openapi3.Schema{}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, typeFormatString(tc.schema))
		})
	}
}
