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

// This file audits the rule registry against every possible edit of an
// OpenAPI document (each field location with its applicable syntactic
// actions, enumerated by checker/metaschema).
//
// TestRuleLocations is the guard: every rule's location claims must parse and
// match at least one edit, so claims cannot drift from the object model.
//
// TestRuleCoverageReport is informational: how much of the wire-relevant edit
// space the mapped rules cover, and where the holes are. Run with:
//
//	go test ./checker -run RuleCoverageReport -v
//
// Set OASDIFF_COVERAGE_DUMP=<path> to also write the full list of uncovered
// edits as TSV.

func TestRuleLocations(t *testing.T) {
	edits := metaschema.Edits()

	for _, rule := range checker.GetAllRules() {
		for _, loc := range rule.Locations {
			claim, err := metaschema.ParseClaim(loc)
			if err != nil {
				t.Errorf("rule %s: %v", rule.Id, err)
				continue
			}
			if !slices.ContainsFunc(edits, claim.Matches) {
				t.Errorf("rule %s: claim %q matches no known edit of an OpenAPI document", rule.Id, loc)
			}
		}
	}
}

// TestRuleCoverage is the completeness guard: every wire-relevant edit
// must be covered by a rule's location claim or waived in checker.CoverageWaivers
// with a reason.
func TestRuleCoverage(t *testing.T) {
	holes := uncoveredCells(t)

	waived := make([]bool, len(checker.CoverageWaivers))
	var unwaived []metaschema.Edit
	for _, edit := range holes {
		found := false
		for i, w := range checker.CoverageWaivers {
			if waiverMatches(t, w.Pattern, edit) {
				waived[i], found = true, true
			}
		}
		if !found {
			unwaived = append(unwaived, edit)
		}
	}

	const maxReport = 40
	for i, edit := range unwaived {
		if i == maxReport {
			t.Errorf("... and %d more unwaived edits", len(unwaived)-maxReport)
			break
		}
		t.Errorf("uncovered edit with no waiver: %s %s\n  add a rule claim or a checker.CoverageWaivers entry", edit.Location, edit.Action)
	}
	for i, w := range checker.CoverageWaivers {
		if !waived[i] {
			t.Errorf("stale coverage waiver: %q\n  it matches no uncovered edit; remove it", w.Pattern)
		}
	}
}

func waiverMatches(t *testing.T, pattern string, edit metaschema.Edit) bool {
	t.Helper()
	matches, err := checker.WaiverMatches(pattern, edit)
	if err != nil {
		t.Fatalf("invalid waiver pattern %q: %v", pattern, err)
	}
	return matches
}

// uncoveredCells returns the wire-relevant edits no rule claim covers.
func uncoveredCells(t *testing.T) []metaschema.Edit {
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

	var holes []metaschema.Edit
	for _, edit := range metaschema.Edits() {
		if edit.Annotation || edit.Extension {
			continue
		}
		covered := false
		for _, c := range claims {
			if c.Matches(edit) {
				covered = true
				break
			}
		}
		if !covered {
			holes = append(holes, edit)
		}
	}
	return holes
}

func TestRuleCoverageReport(t *testing.T) {
	edits := metaschema.Edits()
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
	var holes []metaschema.Edit
	for _, edit := range edits {
		if edit.Annotation || edit.Extension {
			continue
		}
		wire++
		editCovered := false
		for _, c := range claims {
			if c.claim.Matches(edit) {
				editCovered = true
				break
			}
		}
		if editCovered {
			covered++
		} else {
			holes = append(holes, edit)
		}
	}

	t.Logf("rules: %d total, %d mapped to locations", len(rules), mapped)
	t.Logf("edits: %d edits, %d wire-relevant (excluding annotation and x-*)", len(edits), wire)
	t.Logf("covered: %d/%d wire-relevant edits (%.1f%%)", covered, wire, 100*float64(covered)/float64(wire))

	byPolarity := map[metaschema.Polarity]int{}
	byContext := map[string]int{}
	for _, edit := range holes {
		byPolarity[edit.Polarity]++
		byContext[reportGroup(edit.Location)]++
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
		for _, edit := range holes {
			fmt.Fprintf(&b, "%s\t%s\t%s\n", edit.Location, edit.Action, edit.Polarity)
		}
		if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
			t.Fatalf("writing coverage dump: %v", err)
		}
		t.Logf("wrote %d uncovered edits to %s", len(holes), path)
	}
}

// reportGroup truncates a location to its coarse position in the document,
// so the informational report groups holes into readable buckets; the
// rendered coverage map uses the same grouping (see checker.CoverageDoc).
func reportGroup(location string) string {
	segs := strings.Split(location, ".")
	for i, seg := range segs {
		switch seg {
		case "parameters", "requestBody", "responses", "callbacks", "securitySchemes":
			return strings.Join(segs[:min(i+1, len(segs))], ".")
		}
	}
	return strings.Join(segs[:min(3, len(segs))], ".")
}
