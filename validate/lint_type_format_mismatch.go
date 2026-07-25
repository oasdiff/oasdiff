package validate

import (
	"fmt"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
)

// TypeFormatMismatchID flags a schema whose `format` is a known format defined
// for a different `type`, e.g. {"type": "integer", "format": "date-time"}. The
// format never applies to that instance type, so spec-compliant tooling
// silently ignores it: not a spec violation, almost always an authoring slip
// where the type or the format was changed without the other. See
// oasdiff/oasdiff#1019.
const TypeFormatMismatchID = "type-format-mismatch"

// formatsByType lists the formats each type defines: the OpenAPI Data Type
// table plus the standard JSON Schema string formats. A format outside its
// type's set is what this lint reports.
//
// `bigint` is included for integer because oasdiff's own format-containment
// logic recognizes it (see checker.isFormatContained), so flagging it here
// would contradict the breaking-change checks.
var formatsByType = map[string][]string{
	"integer": {"int32", "int64", "bigint"},
	"number":  {"float", "double"},
	"string": {
		// OpenAPI Data Type table
		"byte", "binary", "date", "date-time", "password",
		// Standard JSON Schema string formats
		"time", "duration", "email", "idn-email", "hostname", "idn-hostname",
		"ipv4", "ipv6", "uri", "uri-reference", "uri-template", "iri",
		"iri-reference", "uuid", "regex", "json-pointer", "relative-json-pointer",
	},
}

// lintTypeFormatMismatch reports, at WARN, every schema declaring a format that
// is known but belongs to a different type. It uses kin-openapi's WalkSchemas so
// traversal and $ref-cycle guarding are owned upstream and each finding's
// Section is the schema's exact RFC 6901 JSON Pointer, matching the other
// native lints.
func lintTypeFormatMismatch(spec *openapi3.T, source string) formatters.Findings {
	var findings formatters.Findings
	// The callback never errors, so WalkSchemas never returns one.
	_ = spec.WalkSchemas(func(jsonPointer string, schema *openapi3.SchemaRef) error {
		s := schema.Value // WalkSchemas guarantees schema and schema.Value are non-nil
		if owner, mismatched := mismatchedFormatOwner(s); mismatched {
			findings = append(findings, newTypeFormatMismatchFinding(s, jsonPointer, owner, source))
		}
		return nil
	})
	return findings
}

// mismatchedFormatOwner reports whether the schema's format is a known format
// that no declared type defines, and if so which type does define it.
//
// It stays silent unless it is certain:
//   - no format, or no declared type: nothing to compare (a format on an untyped
//     schema may apply to whatever type the instance turns out to be).
//   - an unrecognized format: OpenAPI says `format` is an open string-valued
//     property and tools should ignore formats they do not recognize, so a
//     custom format is never a finding.
//   - a type array (OpenAPI 3.1, e.g. ["string","null"]): the format only has to
//     fit one of the declared types. "null" is ignored, since a nullable string
//     is still a string.
func mismatchedFormatOwner(s *openapi3.Schema) (string, bool) {
	if s.Format == "" || s.Type == nil {
		return "", false
	}

	types := withoutNullType(s.Type.Slice())
	if len(types) == 0 {
		return "", false
	}

	owner, known := formatOwner[s.Format]
	if !known {
		// Unrecognized format: allowed by the spec, so not reported.
		return "", false
	}
	if slices.Contains(types, owner) {
		// A declared type defines this format, so it applies.
		return "", false
	}
	return owner, true
}

// formatOwner maps each known format to the type that defines it, derived from
// formatsByType so that table stays the single source of truth (adding a type
// there needs no change here). Every format belongs to exactly one type, which
// TestFormatOwner_OneTypePerFormat asserts.
var formatOwner = invertFormatsByType()

func invertFormatsByType() map[string]string {
	owners := map[string]string{}
	for declaredType, formats := range formatsByType {
		for _, format := range formats {
			owners[format] = declaredType
		}
	}
	return owners
}

// withoutNullType drops the JSON Schema "null" type so an OpenAPI 3.1 nullable
// type (e.g. ["string","null"]) is compared by its non-null type(s).
func withoutNullType(types []string) []string {
	out := make([]string, 0, len(types))
	for _, t := range types {
		if t != "null" {
			out = append(out, t)
		}
	}
	return out
}

func newTypeFormatMismatchFinding(s *openapi3.Schema, section, owner, source string) formatters.Finding {
	declared := strings.Join(withoutNullType(s.Type.Slice()), ", ")
	line, column := schemaFieldLocation(s, "format")
	f := formatters.Finding{
		Id: TypeFormatMismatchID,
		Text: fmt.Sprintf("format %q is defined for type %q, not %q, so it is ignored",
			s.Format, owner, declared),
		Level:   checker.WARN,
		Section: section,
		Source: formatters.Source{
			File:   source,
			Line:   line,
			Column: column,
		},
	}
	f.Fingerprint = checker.ComputeFingerprint(f.Id, "", section, []any{s.Format, declared})
	return f
}
