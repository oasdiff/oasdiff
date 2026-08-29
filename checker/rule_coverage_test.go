package checker_test

import (
	"slices"
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/checker/coverage"
	"github.com/oasdiff/oasdiff/checker/metaschema"
)

// This file audits the rule registry against every possible edit of an
// OpenAPI document (each field location with its applicable syntactic
// actions, enumerated by checker/metaschema).
//
// TestRuleLocations guards the claims: every rule's location claims must
// parse and match at least one edit, so claims cannot drift from the object
// model.
//
// TestRuleCoverage guards completeness: every wire-relevant edit must be
// covered by a rule or waived with a reason. Run
// `oasdiff checks changelog coverage` for the full accounting.

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

func TestRuleCoverage(t *testing.T) {
	const maxReport = 40

	waivedReasons := map[string]bool{}
	uncovered := 0
	for _, edit := range coverage.Analyze(checker.GetAllRules().Metadata()) {
		switch edit.Status {
		case coverage.Waived:
			waivedReasons[edit.Reason] = true
		case coverage.Uncovered:
			uncovered++
			if uncovered <= maxReport {
				t.Errorf("uncovered edit with no waiver: %s %s\n  add a rule claim or a coverage waiver", edit.Location, edit.Action)
			}
		}
	}
	if uncovered > maxReport {
		t.Errorf("... and %d more uncovered edits", uncovered-maxReport)
	}

	for _, waiver := range coverage.Waivers {
		if !waivedReasons[waiver.Reason] {
			t.Errorf("stale coverage waiver: %q\n  it accounts for no uncovered edit; remove it", waiver.Pattern)
		}
	}
}
