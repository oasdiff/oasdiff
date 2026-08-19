package checker

import (
	"fmt"
	"sort"
	"strings"

	"github.com/oasdiff/oasdiff/checker/metaschema"
)

// CoverageWaiver explains one family of possible edits that no rule covers.
type CoverageWaiver struct {
	// Pattern is a location glob (see metaschema.MatchLocation), optionally
	// restricted to actions with ":action[,action...]"; without the suffix
	// it waives every action at the location.
	Pattern string
	// Reason starts with a category: resolved-at-usage (component
	// definitions are compared at their referencing operations after $ref
	// resolution), not-contract (the field does not change which payloads
	// are valid), covered-as (the same document edit is already reported
	// under another action at the same location), or open (a candidate
	// missing check, with its tracking issue).
	Reason string
}

// CoverageWaivers records why each wire-relevant edit with no rule is
// deliberately or knowingly uncovered. The checker's TestRuleCoverage fails
// on any uncovered edit no waiver matches (add a rule or a waiver) and on
// any waiver matching no uncovered edit (remove the stale waiver), so the
// list stays an honest, reviewed record.
var CoverageWaivers = []CoverageWaiver{
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
	{"**.schema.multipleOf", "open: response set (narrowing the output) and non-body contexts are unchecked; the breaking directions have rules (tracked in #1159)"},
	{"**.schema.uniqueItems", "open: response set (narrowing the output) and non-body contexts are unchecked; the breaking directions have rules (tracked in #1159)"},
	{"**.schema.maxProperties", "open: remaining directions and non-body contexts are unchecked; the breaking directions have rules (tracked in #1159)"},
	{"**.schema.minProperties", "open: remaining directions and non-body contexts are unchecked; the breaking directions have rules (tracked in #1159, #1171 for the set case)"},
	{"**.schema.items:set,unset", "open: an items subschema appearing on a request narrows accepted arrays (breaking) and is unchecked (tracked in #1054)"},
	{"**.schema.not", "open: a not subschema appearing on a request narrows the accepted set (breaking) and is unchecked (tracked in #1054)"},
	{"**.schema.maxItems", "open: remaining directions (request unset widens, response set/decrease narrow the server's output) and non-body contexts are unchecked; the breaking directions have rules (tracked in #1159)"},
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

// WaiverMatches reports whether a waiver pattern covers the edit.
func WaiverMatches(pattern string, edit metaschema.Edit) (bool, error) {
	if !strings.Contains(pattern, ":") {
		return metaschema.MatchLocation(pattern, edit.Location), nil
	}
	claim, err := metaschema.ParseClaim(pattern)
	if err != nil {
		return false, err
	}
	return claim.Matches(edit), nil
}

// coverageGroup truncates a location to its coarse position in the document, so
// the report groups holes into readable buckets.
func coverageGroup(location string) string {
	segs := strings.Split(location, ".")
	for i, seg := range segs {
		switch seg {
		case "parameters", "requestBody", "responses", "callbacks", "securitySchemes":
			return strings.Join(segs[:min(i+1, len(segs))], ".")
		}
	}
	return strings.Join(segs[:min(3, len(segs))], ".")
}

// CoverageDoc renders the coverage map of the changelog checks over every
// possible edit of an OpenAPI document, as the markdown document served by
// `oasdiff checks coverage` and checked in as docs/COVERAGE.md.
func CoverageDoc() string {
	edits := metaschema.Edits()

	type ruleRef struct {
		id    string
		level Level
	}
	type claimant struct {
		claim metaschema.Claim
		ruleRef
	}
	var claims []claimant
	for _, rule := range GetAllRules() {
		for _, loc := range rule.Locations {
			claim, err := metaschema.ParseClaim(loc)
			if err != nil {
				continue // TestRuleLocations reports it
			}
			claims = append(claims, claimant{claim, ruleRef{rule.Id, rule.Level}})
		}
	}

	covered := map[metaschema.Edit][]ruleRef{}
	var wire, wireCovered int
	var holes []metaschema.Edit
	for _, edit := range edits {
		for _, c := range claims {
			if c.claim.Matches(edit) {
				covered[edit] = append(covered[edit], c.ruleRef)
			}
		}
		if edit.Annotation || edit.Extension {
			continue
		}
		wire++
		if len(covered[edit]) > 0 {
			wireCovered++
		} else {
			holes = append(holes, edit)
		}
	}

	// attribute each uncovered wire-relevant edit to its first matching waiver
	waiverEdits := make([]int, len(CoverageWaivers))
	for _, edit := range holes {
		for i, w := range CoverageWaivers {
			if matches, _ := WaiverMatches(w.Pattern, edit); matches {
				waiverEdits[i]++
				break
			}
		}
	}

	var b strings.Builder
	b.WriteString(`# Coverage of OpenAPI Document Edits

<!-- Generated by "oasdiff checks coverage"; do not edit by hand. -->

This page maps the changelog checks onto every possible edit of an OpenAPI
document. The edits are derived mechanically from the OpenAPI object
model: every field location (a dotted path, with ` + "`*`" + ` standing for any map key
or list index) paired with the syntactic edits applicable there (add, remove,
set, unset, change, increase, decrease). Schema locations are folded: an edit
like ` + "`paths.*.*.requestBody.content.*.schema.maxLength`" + ` stands for that
keyword at any nesting depth inside the request body schema.

Two guarantees are enforced by tests in the checker package:

- Every location a check claims exists in the object model, so this map
  cannot drift from the specification as the parser evolves.
- Every wire-relevant edit (one whose edit can change which payloads are
  valid) is either covered by at least one check or listed in the second
  table below with a reason. A new field in the object model, or a removed
  check, fails the build until this map accounts for it.

Checks for the same edit differ by preconditions on the document (for
example, whether the removed endpoint was deprecated); severity levels are
listed per check.

`)

	fmt.Fprintf(&b, "## Summary\n\n")
	fmt.Fprintf(&b, "- %d possible edits, %d of them wire-relevant\n", len(edits), wire)
	fmt.Fprintf(&b, "- %d wire-relevant edits covered by checks\n", wireCovered)
	fmt.Fprintf(&b, "- %d wire-relevant edits without checks, all accounted for below\n\n", len(holes))

	b.WriteString("## Checked edits\n")
	groups := map[string][]metaschema.Edit{}
	for edit := range covered {
		groups[coverageGroup(edit.Location)] = append(groups[coverageGroup(edit.Location)], edit)
	}
	groupNames := make([]string, 0, len(groups))
	for g := range groups {
		groupNames = append(groupNames, g)
	}
	sort.Strings(groupNames)
	for _, g := range groupNames {
		edits := groups[g]
		sort.Slice(edits, func(i, j int) bool {
			if edits[i].Location != edits[j].Location {
				return edits[i].Location < edits[j].Location
			}
			return edits[i].Action < edits[j].Action
		})
		fmt.Fprintf(&b, "\n### `%s`\n\n", g)
		b.WriteString("| Location | Action | Checks |\n|---|---|---|\n")
		for _, edit := range edits {
			refs := covered[edit]
			sort.Slice(refs, func(i, j int) bool { return refs[i].id < refs[j].id })
			names := make([]string, len(refs))
			for i, r := range refs {
				names[i] = fmt.Sprintf("`%s` (%s)", r.id, r.level.String())
			}
			fmt.Fprintf(&b, "| `%s` | %s | %s |\n", edit.Location, edit.Action, strings.Join(names, ", "))
		}
	}

	b.WriteString(`
## Edits without checks

Each remaining wire-relevant edit matches one of the entries below. Counts
attribute every edit to its first matching entry.

| Pattern | Edits | Reason |
|---|---|---|
`)
	for i, w := range CoverageWaivers {
		fmt.Fprintf(&b, "| `%s` | %d | %s |\n", w.Pattern, waiverEdits[i], w.Reason)
	}
	return b.String()
}
