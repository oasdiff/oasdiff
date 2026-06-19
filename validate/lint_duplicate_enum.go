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

// lintDuplicateEnums walks every schema reachable from the spec and reports, at
// WARN, any enum that contains duplicate values. The walk is single-spec (no
// diff), guards against $ref cycles via a visited set, and recurses through the
// common schema-nesting fields.
func lintDuplicateEnums(spec *openapi3.T, source string) formatters.Findings {
	l := &enumLinter{source: source, seen: map[*openapi3.Schema]bool{}}

	if spec.Components != nil {
		for name, ref := range spec.Components.Schemas {
			l.visit(ref, "components/schemas/"+name)
		}
	}

	if spec.Paths != nil {
		for path, item := range spec.Paths.Map() {
			if item == nil {
				continue
			}
			for method, op := range item.Operations() {
				ctx := method + " " + path
				for _, p := range op.Parameters {
					if p.Value != nil {
						l.visit(p.Value.Schema, ctx)
					}
				}
				if op.RequestBody != nil && op.RequestBody.Value != nil {
					for _, mt := range op.RequestBody.Value.Content {
						l.visit(mt.Schema, ctx)
					}
				}
				if op.Responses != nil {
					for _, resp := range op.Responses.Map() {
						if resp.Value == nil {
							continue
						}
						for _, mt := range resp.Value.Content {
							l.visit(mt.Schema, ctx)
						}
						for _, h := range resp.Value.Headers {
							if h.Value != nil {
								l.visit(h.Value.Schema, ctx)
							}
						}
					}
				}
			}
		}
	}

	return l.findings
}

type enumLinter struct {
	source   string
	seen     map[*openapi3.Schema]bool
	findings formatters.Findings
}

func (l *enumLinter) visit(ref *openapi3.SchemaRef, ctx string) {
	if ref == nil || ref.Value == nil {
		return
	}
	s := ref.Value
	if l.seen[s] {
		return
	}
	l.seen[s] = true

	if dups := duplicateEnumValues(s.Enum); len(dups) > 0 {
		l.findings = append(l.findings, l.newFinding(s, ctx, dups))
	}

	for _, sub := range s.Properties {
		l.visit(sub, ctx)
	}
	l.visit(s.Items, ctx)
	for _, sub := range s.OneOf {
		l.visit(sub, ctx)
	}
	for _, sub := range s.AnyOf {
		l.visit(sub, ctx)
	}
	for _, sub := range s.AllOf {
		l.visit(sub, ctx)
	}
	l.visit(s.Not, ctx)
	if s.AdditionalProperties.Schema != nil {
		l.visit(s.AdditionalProperties.Schema, ctx)
	}
}

func (l *enumLinter) newFinding(s *openapi3.Schema, ctx string, dups []string) formatters.Finding {
	quoted := make([]string, len(dups))
	for i, d := range dups {
		quoted[i] = fmt.Sprintf("%q", d)
	}
	line, column := enumLocation(s)
	f := formatters.Finding{
		Id:      DuplicateEnumValueID,
		Text:    fmt.Sprintf("enum contains duplicate value(s): %s", strings.Join(quoted, ", ")),
		Level:   checker.WARN,
		Section: ctx,
		Source: formatters.Source{
			File:   l.source,
			Line:   line,
			Column: column,
		},
	}
	args := make([]any, len(dups))
	for i, d := range dups {
		args[i] = d
	}
	f.Fingerprint = formatters.ComputeFingerprint(f.Id, "", ctx, args)
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
