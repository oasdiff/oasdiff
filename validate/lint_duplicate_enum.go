package validate

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
)

// DuplicateEnumValueID flags an enum array that contains the same value more
// than once. JSON Schema says enum elements SHOULD be unique (a recommendation,
// not a MUST), so kin-openapi does not reject them; this is an oasdiff-native
// SHOULD-level lint. See oasdiff/oasdiff#980.
const DuplicateEnumValueID = "duplicate-enum-value"

// lintDuplicateEnums reports, at WARN, every enum that contains duplicate
// values. It uses kin-openapi's WalkSchemas to visit each schema in the
// document exactly once -- components, paths/operations, webhooks, and all
// sub-schema keywords -- so the traversal and $ref-cycle guarding are owned
// upstream rather than re-derived here, and the finding's Section is the
// schema's exact RFC 6901 JSON Pointer.
func lintDuplicateEnums(spec *openapi3.T, source string) formatters.Findings {
	var findings formatters.Findings
	// The callback never errors, so WalkSchemas never returns one.
	_ = spec.WalkSchemas(func(jsonPointer string, schema *openapi3.SchemaRef) error {
		s := schema.Value // WalkSchemas guarantees schema and schema.Value are non-nil
		if dups := duplicateEnumValues(s.Enum); len(dups) > 0 {
			findings = append(findings, newDuplicateEnumFinding(s, jsonPointer, dups, source))
		}
		return nil
	})
	return findings
}

func newDuplicateEnumFinding(s *openapi3.Schema, section string, dups []string, source string) formatters.Finding {
	quoted := make([]string, len(dups))
	for i, d := range dups {
		quoted[i] = fmt.Sprintf("%q", d)
	}
	line, column := enumLocation(s)
	f := formatters.Finding{
		Id:      DuplicateEnumValueID,
		Text:    fmt.Sprintf("enum contains duplicate value(s): %s", strings.Join(quoted, ", ")),
		Level:   checker.WARN,
		Section: section,
		Source: formatters.Source{
			File:   source,
			Line:   line,
			Column: column,
		},
	}
	args := make([]any, len(dups))
	for i, d := range dups {
		args[i] = d
	}
	f.Fingerprint = checker.ComputeFingerprint(f.Id, "", section, args)
	return f
}

// enumLocation returns the line/column of the schema's enum field when origin
// tracking is enabled, falling back to the schema's own location, then 0.
func enumLocation(s *openapi3.Schema) (int, int) {
	if s.Origin == nil {
		return 0, 0
	}
	if loc, ok := s.Origin.Fields["enum"]; ok {
		return loc.Line, loc.Column
	}
	if s.Origin.Key != nil {
		return s.Origin.Key.Line, s.Origin.Key.Column
	}
	return 0, 0
}

// duplicateEnumValues returns the distinct enum values that appear more than
// once, in first-seen order. Values are compared by their JSON encoding so
// "1" (string) and 1 (number) do not collide.
func duplicateEnumValues(enum []any) []string {
	seen := map[string]bool{}
	reported := map[string]bool{}
	var dups []string
	for _, v := range enum {
		key := enumKey(v)
		if seen[key] {
			if !reported[key] {
				reported[key] = true
				dups = append(dups, displayEnum(v))
			}
			continue
		}
		seen[key] = true
	}
	return dups
}

func enumKey(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

func displayEnum(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}
