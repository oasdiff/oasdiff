package validate

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
)

// RequiredWithDefaultID flags a required parameter or property that also
// declares a default value. The two contradict each other: `required` means the
// value must always be present, so the `default` (the value assumed when it is
// absent) can never apply. It is dead and misleading, and usually a mistake,
// the author meant the field to be optional. Valid OpenAPI, an oasdiff-native
// SHOULD-level lint.
const RequiredWithDefaultID = "required-with-default"

// lintRequiredWithDefault reports, at WARN, every required parameter and every
// required schema property that also has a default value. Parameters are
// visited with WalkParameters and schema properties with WalkSchemas, so each
// is reported once at its definition and $ref dedup is owned upstream.
func lintRequiredWithDefault(spec *openapi3.T, source string) formatters.Findings {
	var findings formatters.Findings

	// Required parameters (query, path, header, cookie) with a default.
	_ = spec.WalkParameters(func(jsonPointer string, ref *openapi3.ParameterRef) error {
		p := ref.Value // WalkParameters guarantees ref and ref.Value are non-nil
		if p.Required && p.Schema != nil && p.Schema.Value != nil && p.Schema.Value.Default != nil {
			line, column := defaultLocation(p.Schema.Value)
			findings = append(findings, newRequiredWithDefaultFinding(
				fmt.Sprintf("required parameter %q has a default value, which is never used because a required parameter must always be sent", p.Name),
				jsonPointer, p.Name, line, column, source))
		}
		return nil
	})

	// Required schema properties with a default. A schema's `required` lists
	// property names; the default lives on each property's own schema.
	_ = spec.WalkSchemas(func(jsonPointer string, ref *openapi3.SchemaRef) error {
		s := ref.Value // WalkSchemas guarantees ref and ref.Value are non-nil
		for _, name := range s.Required {
			prop, ok := s.Properties[name]
			if !ok || prop.Value == nil || prop.Value.Default == nil {
				continue
			}
			line, column := defaultLocation(prop.Value)
			findings = append(findings, newRequiredWithDefaultFinding(
				fmt.Sprintf("required property %q has a default value, which is never used because a required property must always be present", name),
				jsonPointer+"/properties/"+name, name, line, column, source))
		}
		return nil
	})

	return findings
}

func newRequiredWithDefaultFinding(text, section, name string, line, column int, source string) formatters.Finding {
	f := formatters.Finding{
		Id:      RequiredWithDefaultID,
		Text:    text,
		Level:   checker.WARN,
		Section: section,
		Source: formatters.Source{
			File:   source,
			Line:   line,
			Column: column,
		},
	}
	f.Fingerprint = checker.ComputeFingerprint(f.Id, "", section, []any{name})
	return f
}

// defaultLocation returns the line/column of the schema's `default` field when
// origin tracking is enabled, falling back to the schema's own location, then 0.
func defaultLocation(s *openapi3.Schema) (int, int) {
	if s.Origin != nil {
		if loc, ok := s.Origin.Fields["default"]; ok {
			return loc.Line, loc.Column
		}
		if s.Origin.Key != nil {
			return s.Origin.Key.Line, s.Origin.Key.Column
		}
	}
	return 0, 0
}
