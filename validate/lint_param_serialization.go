package validate

import (
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
)

// AmbiguousParameterSerializationID flags a parameter whose schema type union
// mixes a structured type (array or object) with a scalar. Style serialization
// is defined per type and arrays/objects serialize differently from scalars,
// so for `type: [array, integer]` a server cannot tell whether ?token=5 is the
// array ["5"] or the integer 5. Valid JSON Schema, under-specified for OpenAPI
// serialization: an oasdiff-native SHOULD-level lint. See oasdiff/oasdiff#1055.
const AmbiguousParameterSerializationID = "ambiguous-parameter-serialization"

// lintAmbiguousParamSerialization reports, at WARN, every parameter (query,
// path, header, cookie) whose schema type union contains both a structured
// type and a scalar. `null` is excluded from the check: it carries no
// serialization of its own, so [array, null] and [string, null] are
// unambiguous. It uses kin-openapi's WalkParameters to visit each parameter in
// the document exactly once -- components, path items, operations, callbacks,
// and webhooks -- so the traversal and $ref dedup are owned upstream (a shared
// components parameter is reported once, at its definition), and the finding's
// Section is the parameter's exact RFC 6901 JSON Pointer.
func lintAmbiguousParamSerialization(spec *openapi3.T, source string) formatters.Findings {
	var findings formatters.Findings
	// The callback never errors, so WalkParameters never returns one.
	_ = spec.WalkParameters(func(jsonPointer string, ref *openapi3.ParameterRef) error {
		p := ref.Value // WalkParameters guarantees ref and ref.Value are non-nil
		if p.Schema == nil || p.Schema.Value == nil || p.Schema.Value.Type == nil {
			return nil
		}
		if mixed := mixedSerializationTypes(p.Schema.Value.Type.Slice()); mixed != nil {
			findings = append(findings, newAmbiguousParamFinding(p, jsonPointer, mixed, source))
		}
		return nil
	})
	return findings
}

// mixedSerializationTypes returns the type union (null excluded, input order)
// when it mixes a structured type with a scalar, else nil.
func mixedSerializationTypes(types []string) []string {
	var kept []string
	var structured, scalar bool
	for _, t := range types {
		switch t {
		case "null":
			continue
		case "array", "object":
			structured = true
		default:
			scalar = true
		}
		kept = append(kept, t)
	}
	if structured && scalar {
		return kept
	}
	return nil
}

func newAmbiguousParamFinding(p *openapi3.Parameter, section string, types []string, source string) formatters.Finding {
	line, column := paramTypeLocation(p)
	f := formatters.Finding{
		Id:      AmbiguousParameterSerializationID,
		Text:    fmt.Sprintf("parameter %q mixes a structured type with a scalar (%s): its serialization is ambiguous", p.Name, strings.Join(types, ", ")),
		Level:   checker.WARN,
		Section: section,
		Source: formatters.Source{
			File:   source,
			Line:   line,
			Column: column,
		},
	}
	args := make([]any, len(types))
	for i, t := range types {
		args[i] = t
	}
	f.Fingerprint = checker.ComputeFingerprint(f.Id, "", section, args)
	return f
}

// paramTypeLocation returns the line/column of the schema's type field when
// origin tracking is enabled, falling back to the schema's own location, then
// the parameter's, then 0.
func paramTypeLocation(p *openapi3.Parameter) (int, int) {
	if s := p.Schema.Value; s.Origin != nil {
		if loc, ok := s.Origin.Fields["type"]; ok {
			return loc.Line, loc.Column
		}
		if s.Origin.Key != nil {
			return s.Origin.Key.Line, s.Origin.Key.Column
		}
	}
	if p.Origin != nil && p.Origin.Key != nil {
		return p.Origin.Key.Line, p.Origin.Key.Column
	}
	return 0, 0
}
