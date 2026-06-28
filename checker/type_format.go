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
	return typeFormatValue(schemaDiff.Base, getTypeFormatDimension(schemaDiff))
}

func getRevisionTypeFormat(schemaDiff *diff.SchemaDiff) string {
	return typeFormatValue(schemaDiff.Revision, getTypeFormatDimension(schemaDiff))
}

// typeFormatValue renders the side of a type/format change so the value agrees
// with the changed dimension (see getTypeFormatDimension): just the type for a
// "type" change, just the format for a "format" change, and the combined
// "type/format" when both changed. Without this a format-only change would read
// "the format was changed from `string/uuid` to `string`", forcing the reader to
// spot the missing "/uuid".
func typeFormatValue(s *openapi3.Schema, dimension string) string {
	switch dimension {
	case "format":
		return formatString(s)
	case "type":
		return typeString(s)
	default:
		return typeFormatString(s)
	}
}

// typeString renders a schema's type (without format): the type names joined by
// ", " (so a nullable ["string","null"] reads "string, null"), with an array's
// item type shown inline ("array<string>") so a bare "array" does not hide what
// the array holds, and "any" when no type constraint is set.
func typeString(s *openapi3.Schema) string {
	if s == nil {
		return "any"
	}

	parts := make([]string, 0, len(s.Type.Slice()))
	for _, t := range s.Type.Slice() {
		if t == "array" && s.Items != nil && s.Items.Value != nil {
			parts = append(parts, "array<"+typeFormatString(s.Items.Value)+">")
		} else {
			parts = append(parts, t)
		}
	}

	if len(parts) == 0 {
		return "any"
	}
	return strings.Join(parts, ", ")
}

// formatString renders a schema's format, or "none" when no format is set (so a
// removed format reads "generalized from `uuid` to `none`" rather than "to “").
func formatString(s *openapi3.Schema) string {
	if s == nil || s.Format == "" {
		return "none"
	}
	return s.Format
}

// typeFormatString renders type and format together ("string",
// "string/date-time", "array<string>, null"), appending "/format" only when a
// format is set so there is no "string/" noise from an empty format.
func typeFormatString(s *openapi3.Schema) string {
	if s == nil {
		return ""
	}

	t := typeString(s)
	if s.Format == "" {
		return t
	}
	if t == "any" {
		// No type but a format is set: show just the format, not "any/uuid".
		return s.Format
	}
	return t + "/" + s.Format
}
