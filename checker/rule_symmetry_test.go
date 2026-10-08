package checker_test

import (
	"fmt"
	"sort"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/checker/metaschema"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// This file audits the rule registry (GetAllRules) for broken symmetries: a
// coordinate populated on one side of a symmetry axis but empty on the
// mirror. Coordinates are built from each rule's declared Direction, Area and
// Kind plus the syntactic actions derived from its location claims; the
// semantic axis uses the rule's declared Effect. Each absence is either a
// real missing check or an intentional asymmetry (usually request/response
// contravariance).
//
// TestRuleSymmetry is the guard: every absence must be listed in
// symmetryWaivers with a reason, otherwise the test fails. A new rule that
// breaks symmetry, or a waiver that no longer applies, both fail the build,
// so the waiver list stays an honest, reviewed record of every intentional
// asymmetry.

// symmetryWaivers records every intentional asymmetry. Key is the canonical
// absence string emitted by symmetryAbsences; value is why it is acceptable.
// Removing a real check, or adding one that fills a gap, must update this map.
var symmetryWaivers = map[string]string{
	"add<->remove request/schema/lifecycle missing-remove":  "deprecation rules claim x-* add/change (sunset annotations appearing or changing); deleting a property-level sunset has no check yet, unlike the operation-level sunset-deleted.",
	"add<->remove response/schema/lifecycle missing-remove": "same as the request side: property-level sunset deletion is unchecked.",
}

// symmetryAbsences returns the canonical key for every coordinate that is
// populated on one side of a symmetry axis but completely empty on the mirror.
func symmetryAbsences(rules checker.BackwardCompatibilityRules) []string {
	var out []string

	// a coordinate of the taxonomy; group is a coordinate without its
	// action, the scope within which action duals and effects are compared;
	// mirror is one without its direction, compared across the two
	type coord struct {
		Direction checker.Direction
		Area      checker.Area
		Kind      checker.Kind
		Action    metaschema.Action
	}
	type group struct {
		Direction checker.Direction
		Area      checker.Area
		Kind      checker.Kind
	}
	type mirror struct {
		Area   checker.Area
		Kind   checker.Kind
		Action metaschema.Action
	}

	// Axis 1: request <-> response, restricted to Areas that appear in both
	// directions (parameters/request-body are request-only, responses/headers
	// response-only, so their missing mirror is structural, not a gap).
	reqAreas, respAreas := map[checker.Area]bool{}, map[checker.Area]bool{}
	for _, r := range rules {
		switch r.Direction {
		case checker.DirectionRequest:
			reqAreas[r.Area] = true
		case checker.DirectionResponse:
			respAreas[r.Area] = true
		}
	}
	req, resp := map[mirror]bool{}, map[mirror]bool{}
	present := map[coord]bool{}
	effects := map[group]map[checker.Effect]bool{}
	for _, r := range rules {
		for _, action := range r.Actions() {
			if reqAreas[r.Area] && respAreas[r.Area] {
				k := mirror{r.Area, r.Kind, action}
				switch r.Direction {
				case checker.DirectionRequest:
					req[k] = true
				case checker.DirectionResponse:
					resp[k] = true
				}
			}
			present[coord{r.Direction, r.Area, r.Kind, action}] = true
		}
		e := group{r.Direction, r.Area, r.Kind}
		if effects[e] == nil {
			effects[e] = map[checker.Effect]bool{}
		}
		effects[e][r.Effect] = true
	}
	for k := range req {
		if !resp[k] {
			out = append(out, fmt.Sprintf("request<->response %s/%s/%s missing-response", k.Area.String(), k.Kind.String(), k.Action))
		}
	}
	for k := range resp {
		if !req[k] {
			out = append(out, fmt.Sprintf("request<->response %s/%s/%s missing-request", k.Area.String(), k.Kind.String(), k.Action))
		}
	}

	// Axis 2: dual action pairs within the same Direction/Area/Kind.
	pairs := [][2]metaschema.Action{
		{metaschema.ActionAdd, metaschema.ActionRemove},
		{metaschema.ActionIncrease, metaschema.ActionDecrease},
		{metaschema.ActionSet, metaschema.ActionUnset},
	}
	for _, p := range pairs {
		coords := map[group]bool{}
		for k := range present {
			if k.Action == p[0] || k.Action == p[1] {
				coords[group{k.Direction, k.Area, k.Kind}] = true
			}
		}
		for c := range coords {
			has0 := present[coord{c.Direction, c.Area, c.Kind, p[0]}]
			has1 := present[coord{c.Direction, c.Area, c.Kind, p[1]}]
			coord := fmt.Sprintf("%s/%s/%s", c.Direction.String(), c.Area.String(), c.Kind.String())
			if has0 && !has1 {
				out = append(out, fmt.Sprintf("%s<->%s %s missing-%s", p[0], p[1], coord, p[1]))
			} else if has1 && !has0 {
				out = append(out, fmt.Sprintf("%s<->%s %s missing-%s", p[0], p[1], coord, p[0]))
			}
		}
	}

	// Axis 3: effect duality within the same Direction/Area/Kind. Where a
	// narrowing verdict exists, its widening counterpart should exist too
	// (usually as the safe-direction changelog entry), and vice versa.
	for e, effs := range effects {
		coord := fmt.Sprintf("%s/%s/%s", e.Direction.String(), e.Area.String(), e.Kind.String())
		if effs[checker.EffectNarrows] && !effs[checker.EffectWidens] {
			out = append(out, fmt.Sprintf("widens<->narrows %s missing-widens", coord))
		} else if effs[checker.EffectWidens] && !effs[checker.EffectNarrows] {
			out = append(out, fmt.Sprintf("widens<->narrows %s missing-narrows", coord))
		}
	}

	sort.Strings(out)
	return out
}

func TestRuleSymmetry(t *testing.T) {
	absences := symmetryAbsences(checker.GetAllRules())

	absent := map[string]bool{}
	for _, a := range absences {
		absent[a] = true
		if _, ok := symmetryWaivers[a]; !ok {
			t.Errorf("unwaived rule asymmetry: %q\n  fix it by adding the mirror rule, or document it in symmetryWaivers with a reason", a)
		}
	}
	for w := range symmetryWaivers {
		if !absent[w] {
			t.Errorf("stale symmetry waiver: %q\n  this asymmetry no longer exists; remove the waiver", w)
		}
	}
}

// The coordinates above do not say where in a payload a rule applies: a
// check on a request body schema and one on a request property share
// request/schema/values, so a keyword checked at one and not the other
// passes TestRuleSymmetry. TestRulePositionSymmetry runs each schema edit at
// every position a schema can sit and requires that an edit reported at one
// position is reported at all of them, or is listed in positionWaivers.
// Bound keywords (maximum, maxLength and the like) are left out: their rules
// are generated for every position and TestBoundCellsFire covers each one.

// positionWaivers records each position where a schema edit is reported
// nowhere although another position reports it. Key: "<edit> <position>".
var positionWaivers = map[string]string{
	"pattern-added request-body":         "missing check, #1295",
	"pattern-removed request-body":       "missing check, #1295",
	"pattern-changed request-body":       "missing check, #1295",
	"pattern-added response-body":        "missing check, #1295",
	"pattern-removed response-body":      "missing check, #1295",
	"pattern-changed response-body":      "missing check, #1295",
	"pattern-added response-header":      "missing check, #1295",
	"pattern-removed response-header":    "missing check, #1295",
	"pattern-changed response-header":    "missing check, #1295",
	"const-added request-parameter":      "missing check, #1295",
	"const-removed request-parameter":    "missing check, #1295",
	"const-changed request-parameter":    "missing check, #1295",
	"const-added request-header":         "missing check, #1295",
	"const-removed request-header":       "missing check, #1295",
	"const-changed request-header":       "missing check, #1295",
	"const-added response-header":        "missing check, #1295",
	"const-removed response-header":      "missing check, #1295",
	"const-changed response-header":      "missing check, #1295",
	"enum-value-added response-header":   "missing check, #1295",
	"enum-value-removed response-header": "missing check, #1295",
	"enum-added response-header":         "missing check, #1295",
	"enum-removed response-header":       "missing check, #1295",
	"default-added response-header":      "missing check, #1295",
	"default-removed response-header":    "missing check, #1295",
	"default-changed response-header":    "missing check, #1295",
}

var schemaPositions = []string{
	"request-body", "request-property", "request-parameter", "request-header",
	"response-body", "response-property", "response-header",
}

// positionDoc builds a spec with schema at position and nothing else that
// can change.
func positionDoc(position string, schema *openapi3.Schema) *load.SpecInfo {
	ref := &openapi3.SchemaRef{Value: schema}
	content := func(s *openapi3.SchemaRef) openapi3.Content {
		return openapi3.Content{"application/json": &openapi3.MediaType{Schema: s}}
	}
	object := &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"object"}, Properties: openapi3.Schemas{"p": ref}}}
	response := &openapi3.Response{Description: new("ok")}
	op := &openapi3.Operation{Responses: openapi3.NewResponses(openapi3.WithStatus(200, &openapi3.ResponseRef{Value: response}))}

	switch position {
	case "request-body":
		op.RequestBody = &openapi3.RequestBodyRef{Value: &openapi3.RequestBody{Content: content(ref)}}
	case "request-property":
		op.RequestBody = &openapi3.RequestBodyRef{Value: &openapi3.RequestBody{Content: content(object)}}
	case "request-parameter":
		op.Parameters = openapi3.Parameters{&openapi3.ParameterRef{Value: &openapi3.Parameter{Name: "q", In: "query", Schema: ref}}}
	case "request-header":
		op.Parameters = openapi3.Parameters{&openapi3.ParameterRef{Value: &openapi3.Parameter{Name: "X-Q", In: "header", Schema: ref}}}
	case "response-body":
		response.Content = content(ref)
	case "response-property":
		response.Content = content(object)
	case "response-header":
		response.Headers = openapi3.Headers{"X-R": &openapi3.HeaderRef{Value: &openapi3.Header{Parameter: openapi3.Parameter{Schema: ref}}}}
	}

	return &load.SpecInfo{Spec: &openapi3.T{
		OpenAPI: "3.1.0",
		Info:    &openapi3.Info{Title: "t", Version: "1.0.0"},
		Paths:   openapi3.NewPaths(openapi3.WithPath("/t", &openapi3.PathItem{Post: op})),
	}}
}

func TestRulePositionSymmetry(t *testing.T) {
	str := func(set func(*openapi3.Schema)) *openapi3.Schema {
		s := &openapi3.Schema{Type: &openapi3.Types{"string"}}
		set(s)
		return s
	}
	plain := func(*openapi3.Schema) {}
	edits := []struct {
		name           string
		base, revision func(*openapi3.Schema)
	}{
		{"enum-value-added", func(s *openapi3.Schema) { s.Enum = []any{"a", "b"} }, func(s *openapi3.Schema) { s.Enum = []any{"a", "b", "c"} }},
		{"enum-value-removed", func(s *openapi3.Schema) { s.Enum = []any{"a", "b", "c"} }, func(s *openapi3.Schema) { s.Enum = []any{"a", "b"} }},
		{"enum-added", plain, func(s *openapi3.Schema) { s.Enum = []any{"a", "b"} }},
		{"enum-removed", func(s *openapi3.Schema) { s.Enum = []any{"a", "b"} }, plain},
		{"pattern-added", plain, func(s *openapi3.Schema) { s.Pattern = "^a$" }},
		{"pattern-removed", func(s *openapi3.Schema) { s.Pattern = "^a$" }, plain},
		{"pattern-changed", func(s *openapi3.Schema) { s.Pattern = "^a$" }, func(s *openapi3.Schema) { s.Pattern = "^b$" }},
		{"format-added", plain, func(s *openapi3.Schema) { s.Format = "uuid" }},
		{"format-removed", func(s *openapi3.Schema) { s.Format = "uuid" }, plain},
		{"format-changed", func(s *openapi3.Schema) { s.Format = "uuid" }, func(s *openapi3.Schema) { s.Format = "email" }},
		{"const-added", plain, func(s *openapi3.Schema) { s.Const = "a" }},
		{"const-removed", func(s *openapi3.Schema) { s.Const = "a" }, plain},
		{"const-changed", func(s *openapi3.Schema) { s.Const = "a" }, func(s *openapi3.Schema) { s.Const = "b" }},
		{"default-added", plain, func(s *openapi3.Schema) { s.Default = "a" }},
		{"default-removed", func(s *openapi3.Schema) { s.Default = "a" }, plain},
		{"default-changed", func(s *openapi3.Schema) { s.Default = "a" }, func(s *openapi3.Schema) { s.Default = "b" }},
		{"type-changed", plain, func(s *openapi3.Schema) { s.Type = &openapi3.Types{"integer"} }},
		{"type-widened", plain, func(s *openapi3.Schema) { s.Type = &openapi3.Types{"string", "integer"} }},
		{"type-narrowed", func(s *openapi3.Schema) { s.Type = &openapi3.Types{"string", "integer"} }, plain},
		{"null-added", plain, func(s *openapi3.Schema) { s.Type = &openapi3.Types{"string", "null"} }},
		{"null-removed", func(s *openapi3.Schema) { s.Type = &openapi3.Types{"string", "null"} }, plain},
	}

	config := allChecksConfig()
	unreported := map[string]bool{}
	for _, edit := range edits {
		reportedAnywhere := false
		var silent []string
		for _, position := range schemaPositions {
			d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), positionDoc(position, str(edit.base)), positionDoc(position, str(edit.revision)))
			require.NoError(t, err)
			if len(checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)) > 0 {
				reportedAnywhere = true
			} else {
				silent = append(silent, position)
			}
		}
		if !reportedAnywhere {
			t.Errorf("%s is reported at no position; the probe no longer exercises a check", edit.name)
			continue
		}
		for _, position := range silent {
			unreported[edit.name+" "+position] = true
		}
	}

	for key := range unreported {
		if _, ok := positionWaivers[key]; !ok {
			t.Errorf("unwaived position asymmetry: %q is reported at other positions but not here\n  fix it by adding the check, or document it in positionWaivers with a reason", key)
		}
	}
	for key := range positionWaivers {
		if !unreported[key] {
			t.Errorf("stale position waiver: %q is reported now; remove the waiver", key)
		}
	}
}
