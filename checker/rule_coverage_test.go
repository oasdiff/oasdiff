package checker_test

import (
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/checker/metaschema"
)

// This file audits the rule registry against the metaschema cube: the full
// edit space of an OpenAPI document (every field location x applicable
// syntactic action, see checker/metaschema).
//
// TestRuleLocations is the guard: every rule's location claims must parse and
// match at least one cube cell, so claims cannot drift from the object model.
//
// TestRuleCoverageReport is informational: how much of the wire-relevant edit
// space the mapped rules cover, and where the holes are. Run with:
//
//	go test ./checker -run RuleCoverageReport -v
//
// Set OASDIFF_COVERAGE_DUMP=<path> to also write the full list of uncovered
// cells as TSV.

func TestRuleLocations(t *testing.T) {
	cube := metaschema.Cube()

	for _, rule := range checker.GetAllRules() {
		for _, loc := range rule.Locations {
			claim, err := metaschema.ParseClaim(loc)
			if err != nil {
				t.Errorf("rule %s: %v", rule.Id, err)
				continue
			}
			if !slices.ContainsFunc(cube, claim.Matches) {
				t.Errorf("rule %s: claim %q matches no cell of the metaschema cube", rule.Id, loc)
			}
		}
	}
}

// coverageWaivers records why each wire-relevant cell with no rule is
// deliberately or knowingly uncovered. A pattern is a location glob
// (see metaschema.MatchLocation), optionally restricted to actions with
// ":action[,action...]"; without the suffix it waives every action at the
// location. TestRuleCoverage fails on any uncovered cell no waiver matches
// (add a rule or a waiver) and on any waiver matching no uncovered cell
// (remove the stale waiver), so the list stays an honest, reviewed record.
//
// Reasons start with a category:
//   - resolved-at-usage: component definitions are compared at their
//     referencing operations after $ref resolution
//   - not-contract: the field does not change which payloads are valid
//   - covered-as: the same document edit is already reported under another
//     action at the same location
//   - open: a candidate missing check surfaced by this audit
var coverageWaivers = []struct{ Pattern, Reason string }{
	{"components.**", "resolved-at-usage: edits to component definitions surface as diffs at every referencing operation; only unused-component removal is reported directly (api-schemas-removed)"},
	{"webhooks.**", "open: webhooks are diffed (WebhooksDiff) but checkers only report webhook add/remove; changes inside a webhook's operations have no checks yet (tracked in #1160)"},
	{"paths.*.$ref", "not-contract: a path item $ref is resolved at load time; the diff compares resolved content"},
	{"paths.*.parameters.**", "open: path-level parameter additions are checked (new-request-*-default-parameter-to-existing-path); modifications and removals at path level have no checks yet (tracked in #1163)"},
	{"servers.**", "not-contract: server URLs are deployment metadata, not part of the request/response contract"},
	{"paths.*.servers.**", "not-contract: server URLs are deployment metadata"},
	{"paths.*.*.servers.**", "not-contract: server URLs are deployment metadata"},
	{"paths.*.*.callbacks.**", "open: callbacks are not checked (tracked in #1161)"},
	{"paths.*.*.responses.*.links.**", "not-contract: links document relationships between operations; they do not change accepted payloads"},
	{"paths.*.*.requestBody.content.*.encoding.**", "open: multipart/form encoding metadata (contentType, per-part headers, style) has no checks (tracked in #1165)"},
	{"paths.*.*.responses.*.content.*.encoding.**", "open: encoding metadata has no checks (tracked in #1165)"},
	{"paths.*.*.parameters.*.content.**", "open: parameters serialized via a content map are not checked (only the schema form is) (tracked in #1166)"},
	{"paths.*.*.requestBody.content.*.itemSchema.**", "open: only itemSchema existence is checked (request-body-media-type-item-schema-added/removed); changes inside it are not (tracked in #1167)"},
	{"paths.*.*.responses.*.content.*.itemSchema.**", "open: only itemSchema existence is checked; changes inside it are not (tracked in #1167)"},
	{"openapi", "not-contract: the OpenAPI dialect version governs parsing, not the API contract"},
	{"jsonSchemaDialect", "not-contract: schema dialect governs parsing, not the API contract"},
	{"info", "not-contract: the info object is required metadata; its presence is a validation concern"},
	{"info.version:set,unset", "not-contract: version is required; presence changes are validation errors, value changes are checked (api-version rules)"},

	// response headers
	{"paths.*.*.responses.*.headers.*.**", "open: response headers are checked for existence, required, and schema type/format/nullable only; serialization fields, the content form, and the remaining schema keywords are unchecked (tracked in #1162)"},

	// parameters
	{"paths.*.*.parameters.*.schema.**", "open: parameter schemas are checked for type/format, enum, bounds, pattern, nullable, default, and required/property membership; the remaining schema keywords are unchecked (tracked in #1054, #1155, #1156, #1157, #1159)"},
	{"paths.*.*.parameters.*.schema:set,unset", "open: a parameter schema appearing or disappearing is unchecked (the media-type analog has request-body-media-type-schema-added/removed) (tracked in #1054)"},
	{"paths.*.*.parameters.*.name", "not-contract: parameters are identified by (name, in); a rename surfaces as parameter remove + add"},
	{"paths.*.*.parameters.*.in", "not-contract: parameters are identified by (name, in); a location change surfaces as parameter remove + add"},
	{"paths.*.*.parameters.*.style", "open: parameter serialization style changes the wire format but is unchecked (tracked in #1164)"},
	{"paths.*.*.parameters.*.explode", "open: explode changes the wire format of array/object parameters but is unchecked (tracked in #1164)"},
	{"paths.*.*.parameters.*.allowReserved", "open: allowReserved changes accepted query characters but is unchecked (tracked in #1164)"},
	{"paths.*.*.parameters.*.allowEmptyValue", "not-contract: allowEmptyValue is deprecated by the OpenAPI spec, which advises against its use"},

	// JSON Schema core and reference keywords
	{"**.schema.$id", "not-contract: reference-resolution keyword; its effect is visible in the resolved schemas the diff compares"},
	{"**.schema.$anchor", "not-contract: reference-resolution keyword"},
	{"**.schema.$dynamicAnchor", "not-contract: reference-resolution keyword"},
	{"**.schema.$dynamicRef", "not-contract: reference-resolution keyword"},
	{"**.schema.$schema", "not-contract: schema dialect declaration, governs validation semantics at parse time"},
	{"**.schema.$defs.**", "not-contract: $defs holds definitions that take effect only where referenced"},
	{"**.schema.$comment", "not-contract: comments carry no validation semantics"},
	{"**.schema.allowEmptyValue", "not-contract: allowEmptyValue is a parameter-object field; on a schema it has no effect"},

	// discriminator internals
	{"**.discriminator.mapping.*:set,unset", "covered-as add/remove: a mapping entry appearing or disappearing is the entry add/remove, which is claimed"},
	{"**.discriminator.propertyName:set,unset", "covered-as discriminator set/unset: propertyName is required inside discriminator, so its presence tracks the discriminator's"},

	// unchecked schema keywords in the request/response body contexts;
	// each is a candidate check surfaced by this audit
	{"**.schema.additionalProperties", "open: setting additionalProperties:false narrows accepted request objects (breaking) and is unchecked (tracked in #1054)"},
	{"**.schema.multipleOf", "open: multipleOf set/tightened on a request narrows accepted numbers (breaking) and is unchecked (tracked in #1155)"},
	{"**.schema.uniqueItems", "open: uniqueItems set on a request narrows accepted arrays (breaking) and is unchecked (tracked in #1156)"},
	{"**.schema.maxProperties", "open: maxProperties set/decreased on a request narrows accepted objects (breaking) and is unchecked (tracked in #1157)"},
	{"**.schema.minProperties", "open: minProperties set/increased on a request narrows accepted objects (breaking) and is unchecked (tracked in #1157)"},
	{"**.schema.items:set,unset", "open: an items subschema appearing on a request narrows accepted arrays (breaking) and is unchecked (tracked in #1054)"},
	{"**.schema.not", "open: a not subschema appearing on a request narrows the accepted set (breaking) and is unchecked (tracked in #1054)"},
	{"**.schema.maxItems", "open: request-side maxItems has no rules (the parameter analog does: request-parameter-max-items-decreased); response side unchecked (tracked in #1158)"},
	{"**.schema.maximum", "open: remaining directions (request unset widens, response set/decrease narrows the server's output) are unchecked; the breaking directions have rules (tracked in #1159)"},
	{"**.schema.minimum", "open: remaining directions are unchecked; the breaking directions have rules (tracked in #1159)"},
	{"**.schema.maxLength", "open: remaining directions are unchecked; the breaking directions have rules (tracked in #1159)"},
	{"**.schema.minLength", "open: remaining directions are unchecked; the breaking directions have rules (tracked in #1159)"},
	{"**.schema.minItems", "open: remaining directions are unchecked; the breaking directions have rules (tracked in #1159)"},
	{"**.schema.exclusiveMaximum", "open: remaining directions are unchecked; the breaking directions have rules (tracked in #1159)"},
	{"**.schema.exclusiveMinimum", "open: remaining directions are unchecked; the breaking directions have rules (tracked in #1159)"},
	{"**.schema.minContains:set,unset", "open: minContains presence changes are unchecked; increase/decrease have rules (tracked in #1159)"},
	{"**.schema.maxContains:set,unset", "open: maxContains presence changes are unchecked; increase/decrease have rules (tracked in #1159)"},
	{"**.schema.unevaluatedItems:change", "open: switching unevaluatedItems between boolean and schema form is unchecked; set/unset have rules (tracked in #1054)"},
	{"**.schema.unevaluatedProperties:change", "open: switching unevaluatedProperties between boolean and schema form is unchecked; set/unset have rules (tracked in #1054)"},
}

// TestRuleCoverage is the completeness guard: every wire-relevant cell of
// the metaschema cube must be covered by a rule's location claim or waived
// in coverageWaivers with a reason.
func TestRuleCoverage(t *testing.T) {
	holes := uncoveredCells(t)

	waived := make([]bool, len(coverageWaivers))
	var unwaived []metaschema.Cell
	for _, cell := range holes {
		found := false
		for i, w := range coverageWaivers {
			if waiverMatches(t, w.Pattern, cell) {
				waived[i], found = true, true
			}
		}
		if !found {
			unwaived = append(unwaived, cell)
		}
	}

	const maxReport = 40
	for i, cell := range unwaived {
		if i == maxReport {
			t.Errorf("... and %d more unwaived cells", len(unwaived)-maxReport)
			break
		}
		t.Errorf("uncovered cell with no waiver: %s %s\n  add a rule claim or a coverageWaivers entry", cell.Location, cell.Action)
	}
	for i, w := range coverageWaivers {
		if !waived[i] {
			t.Errorf("stale coverage waiver: %q\n  it matches no uncovered cell; remove it", w.Pattern)
		}
	}
}

func waiverMatches(t *testing.T, pattern string, cell metaschema.Cell) bool {
	t.Helper()
	if !strings.Contains(pattern, ":") {
		return metaschema.MatchLocation(pattern, cell.Location)
	}
	claim, err := metaschema.ParseClaim(pattern)
	if err != nil {
		t.Fatalf("invalid waiver pattern %q: %v", pattern, err)
	}
	return claim.Matches(cell)
}

// uncoveredCells returns the wire-relevant cells no rule claim covers.
func uncoveredCells(t *testing.T) []metaschema.Cell {
	t.Helper()

	var claims []metaschema.Claim
	for _, rule := range checker.GetAllRules() {
		for _, loc := range rule.Locations {
			claim, err := metaschema.ParseClaim(loc)
			if err != nil {
				continue // TestRuleLocations reports it
			}
			claims = append(claims, claim)
		}
	}

	var holes []metaschema.Cell
	for _, cell := range metaschema.Cube() {
		if cell.Annotation || cell.Extension {
			continue
		}
		covered := false
		for _, c := range claims {
			if c.Matches(cell) {
				covered = true
				break
			}
		}
		if !covered {
			holes = append(holes, cell)
		}
	}
	return holes
}

func TestRuleCoverageReport(t *testing.T) {
	cube := metaschema.Cube()
	rules := checker.GetAllRules()

	type claimant struct {
		claim metaschema.Claim
		id    string
	}
	var claims []claimant
	mapped := 0
	for _, rule := range rules {
		if len(rule.Locations) == 0 {
			continue
		}
		mapped++
		for _, loc := range rule.Locations {
			claim, err := metaschema.ParseClaim(loc)
			if err != nil {
				continue // TestRuleLocations reports it
			}
			claims = append(claims, claimant{claim, rule.Id})
		}
	}

	var wire, covered int
	var holes []metaschema.Cell
	for _, cell := range cube {
		if cell.Annotation || cell.Extension {
			continue
		}
		wire++
		cellCovered := false
		for _, c := range claims {
			if c.claim.Matches(cell) {
				cellCovered = true
				break
			}
		}
		if cellCovered {
			covered++
		} else {
			holes = append(holes, cell)
		}
	}

	t.Logf("rules: %d total, %d mapped to locations", len(rules), mapped)
	t.Logf("cube: %d cells, %d wire-relevant (excluding annotation and x-*)", len(cube), wire)
	t.Logf("covered: %d/%d wire-relevant cells (%.1f%%)", covered, wire, 100*float64(covered)/float64(wire))

	byPolarity := map[metaschema.Polarity]int{}
	byContext := map[string]int{}
	for _, cell := range holes {
		byPolarity[cell.Polarity]++
		byContext[context(cell.Location)]++
	}
	t.Logf("holes by polarity: %v", byPolarity)
	contexts := make([]string, 0, len(byContext))
	for c := range byContext {
		contexts = append(contexts, c)
	}
	sort.Slice(contexts, func(i, j int) bool { return byContext[contexts[i]] > byContext[contexts[j]] })
	for _, c := range contexts {
		t.Logf("  %5d  %s", byContext[c], c)
	}

	if path := os.Getenv("OASDIFF_COVERAGE_DUMP"); path != "" {
		var b strings.Builder
		for _, cell := range holes {
			fmt.Fprintf(&b, "%s\t%s\t%s\n", cell.Location, cell.Action, cell.Polarity)
		}
		if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
			t.Fatalf("writing coverage dump: %v", err)
		}
		t.Logf("wrote %d uncovered cells to %s", len(holes), path)
	}
}

// context truncates a location to its coarse position in the document, so
// the report groups holes into readable buckets.
func context(location string) string {
	segs := strings.Split(location, ".")
	for i, seg := range segs {
		switch seg {
		case "parameters", "requestBody", "responses", "callbacks", "securitySchemes":
			return strings.Join(segs[:min(i+1, len(segs))], ".")
		}
	}
	return strings.Join(segs[:min(3, len(segs))], ".")
}
