package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
)

// Every rule's stored level equals the one the severity law derives
// (rules.DeriveLevel) or appears in severityDeviations with a reason, and no
// deviation entry is stale. Severity is thereby one reviewed function plus a
// ledger of exceptions instead of hundreds of independent choices.
func TestSeverityLaw(t *testing.T) {
	seen := map[string]bool{}
	for _, rule := range checker.GetAllRules() {
		derived := rule.DerivedLevel()
		if derived == rule.Level {
			if _, ok := severityDeviations[rule.Id]; ok {
				t.Errorf("stale severity deviation: %q\n  the stored level now matches the law; remove the entry", rule.Id)
			}
			continue
		}
		seen[rule.Id] = true
		if _, ok := severityDeviations[rule.Id]; !ok {
			t.Errorf("severity law violation: %s stored=%s derived=%s (effect=%s, direction=%s)\n  fix the level or add a severityDeviations entry with a reason",
				rule.Id, rule.Level.String(), derived.String(), rule.Effect.String(), rule.Direction.String())
		}
	}
	for id := range severityDeviations {
		if !seen[id] {
			if _, ok := severityDeviations[id]; ok && !ruleExists(id) {
				t.Errorf("stale severity deviation: %q\n  no such rule; remove the entry", id)
			}
		}
	}
}

// severityDeviations records every rule whose stored level deviates from the
// law, with the reason. An unexplained mismatch and a stale entry both fail
// TestSeverityLaw.
var severityDeviations = map[string]string{}

func ruleExists(id string) bool {
	for _, rule := range checker.GetAllRules() {
		if rule.Id == id {
			return true
		}
	}
	return false
}
