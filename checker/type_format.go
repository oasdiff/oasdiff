package checker

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

func getBaseTypeFormat(schemaDiff *diff.SchemaDiff) string {
	return typeFormatString(schemaDiff.Base)
}

func getRevisionTypeFormat(schemaDiff *diff.SchemaDiff) string {
	return typeFormatString(schemaDiff.Revision)
}

// typeFormatString renders a schema's type for a human-readable message: the type
// names joined by ", " (so a nullable ["string","null"] reads "string, null"),
// with an array's item type shown inline ("array<string>") and the format
// appended after "/" only when one is set ("string/date-time"). Rendering only
// "/format" when a format is present avoids the "string/" noise of an empty
// format, and showing the item type avoids a bare "array" that hides what the
// array holds.
func typeFormatString(s *openapi3.Schema) string {
	if s == nil {
		return ""
	}

	types := s.Type.Slice()
	parts := make([]string, 0, len(types))
	for _, t := range types {
		if t == "array" && s.Items != nil && s.Items.Value != nil {
			parts = append(parts, "array<"+typeFormatString(s.Items.Value)+">")
		} else {
			parts = append(parts, t)
		}
	}

	result := strings.Join(parts, ", ")
	if s.Format == "" {
		return result
	}
	if result == "" {
		return s.Format
	}
	return result + "/" + s.Format
}
