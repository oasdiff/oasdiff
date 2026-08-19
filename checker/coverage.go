package checker

import (
	"fmt"
	"sort"
	"strings"

	"github.com/oasdiff/oasdiff/checker/metaschema"
)

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
// possible edit of an OpenAPI document, as the markdown served by
// `oasdiff checks changelog coverage`.
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
		if !metaschema.WireRelevant(edit) {
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
			if matches, _ := metaschema.MatchPattern(w.Pattern, edit); matches {
				waiverEdits[i]++
				break
			}
		}
	}

	// count the edits each non-contract entry excludes
	nonContractEdits := make([]int, len(metaschema.NonContracts))
	for _, edit := range edits {
		if edit.Annotation || edit.Extension {
			continue
		}
		for i, nc := range metaschema.NonContracts {
			if matches, _ := metaschema.MatchPattern(nc.Pattern, edit); matches {
				nonContractEdits[i]++
				break
			}
		}
	}

	var b strings.Builder
	b.WriteString(`# Coverage of OpenAPI Document Edits

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
- Every wire-relevant edit (one that can change which payloads are valid)
  is either covered by at least one check or listed in the waiver table
  below with a reason. A new field in the object model, or a removed
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
