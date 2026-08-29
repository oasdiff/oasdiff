package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
)

// The severity law: a rule's expected level follows from its declared Effect
// and Guards and its Direction. TestSeverityLaw enforces that every rule's
// stored level equals the derived one or appears in severityDeviations with a
// reason, and that no deviation entry is stale. Severity is thereby one
// reviewed table plus a ledger of exceptions instead of 558 independent
// choices.
//
// Deviations marked "pending review" are the open triage entries from
// SEVERITY-LAW-TRIAGE.md: conventions to be ratified into permanent reasons,
// or candidate severity bugs to be fixed (at which point the entry is removed
// and the law holds directly).

// expectedLevel derives a rule's level from the law.
func expectedLevel(r checker.BackwardCompatibilityRule) checker.Level {
	effect := r.Effect
	direction := r.Direction
	for _, g := range r.Guards {
		switch g {
		case checker.GuardReadOnly:
			// a readOnly property does not appear in requests
			if direction == checker.DirectionRequest {
				effect = checker.EffectNone
			}
		case checker.GuardWriteOnly:
			// a writeOnly property does not appear in responses
			if direction == checker.DirectionResponse {
				effect = checker.EffectNone
			}
		case checker.GuardNonSuccess:
			// the responses map does not promise that the server returns
			// only the statuses it lists
			effect = checker.EffectNone
		case checker.GuardSanctioned:
			// the deprecation contract was honored
			return checker.INFO
		case checker.GuardNegotiated:
			// the client chooses or relies on the variant
			direction = checker.DirectionRequest
		}
	}

	switch effect {
	case checker.EffectViolation, checker.EffectIncomparable:
		return checker.ERR
	case checker.EffectNone:
		return checker.INFO
	case checker.EffectUnknown:
		return checker.WARN
	case checker.EffectNarrows:
		if direction == checker.DirectionResponse {
			return checker.INFO
		}
		return checker.ERR
	case checker.EffectWidens:
		if direction == checker.DirectionResponse {
			return checker.ERR
		}
		return checker.INFO
	}
	return checker.INFO
}

const (
	// above the law: conservative, and a harsher verdict needs a reason
	// rather than a proof (SEVERITY-LAW-TRIAGE.md)
	// below the law: a milder verdict owes a contract argument
	// four rules deviate each way
	reasonPrefixItems = "pending review: stored levels assume a containment direction prefixItems does not have"
)

// severityDeviations records every rule whose stored level deviates from the
// law, with the reason. An unexplained mismatch and a stale entry both fail
// TestSeverityLaw.
var severityDeviations = map[string]string{
	// above the law
	// below the law
	// prefixItems
	"request-body-prefix-items-added":        reasonPrefixItems,
	"request-body-prefix-items-removed":      reasonPrefixItems,
	"request-property-prefix-items-added":    reasonPrefixItems,
	"request-property-prefix-items-removed":  reasonPrefixItems,
	"response-body-prefix-items-added":       reasonPrefixItems,
	"response-body-prefix-items-removed":     reasonPrefixItems,
	"response-property-prefix-items-added":   reasonPrefixItems,
	"response-property-prefix-items-removed": reasonPrefixItems,
}

func TestSeverityLaw(t *testing.T) {
	seen := map[string]bool{}
	for _, rule := range checker.GetAllRules() {
		derived := expectedLevel(rule)
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

func ruleExists(id string) bool {
	for _, rule := range checker.GetAllRules() {
		if rule.Id == id {
			return true
		}
	}
	return false
}
