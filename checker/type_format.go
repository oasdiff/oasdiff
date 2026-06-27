package checker

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

// getTypeFormatDimension names what changed in a type/format change, for the
// message: "type" when only the type changed, "format" when only the format
// changed, and "type/format" when both did. It is passed as a message argument
// so a type-only change does not read "the type/format was changed" when no
// format was involved.
func getTypeFormatDimension(schemaDiff *diff.SchemaDiff) string {
	typeChanged := !schemaDiff.TypeDiff.Empty()
	formatChanged := !schemaDiff.FormatDiff.Empty()
	switch {
	case typeChanged && formatChanged:
		return "type/format"
	case typeChanged:
		return "type"
	case formatChanged:
		return "format"
	default:
		// Neither changed; the caller only emits a change when one did, so this
		// is unreachable in practice. Return empty rather than mislabel it "type".
		return ""
	}
}

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
	if s.Format != "" {
		if result == "" {
			return s.Format
		}
		return result + "/" + s.Format
	}
	if result == "" {
		// No type constraint: render "any" rather than an empty value, which
		// reads as nothing (e.g. a removed type generalizes "string" to "any").
		return "any"
	}
	return result
}
