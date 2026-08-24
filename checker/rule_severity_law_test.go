package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
)

// The severity law: a rule's expected level follows from its declared Effect
// and Guards and its Direction. TestSeverityLaw enforces that every rule's
// stored level equals the derived one or appears in severityDeviations with a
// reason, and that no deviation entry is stale. Severity is thereby one
// reviewed table plus a ledger of exceptions instead of 556 independent
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
	reasonBoundSet     = "below the law, pending review: setting a bound on a request narrows, and the WARN rests on the change being sometimes legitimate, which is an exemption rather than a contract argument"
	reasonBodyRemoved  = "above the law, pending review: the operation withdraws a declared input, so a client that conformed by sending the body can no longer convey it, which the containment question alone does not express"
	reasonRemoval      = "above the law, pending review: the removal reads as widening only while additionalProperties is unrestricted, so the stored level may be the sound one"
	reasonWrapped      = "above the law, pending review: single-to-oneOf wrapping kept breaking per #1037"
	reasonNonSuccess   = "below the law, pending review: removing a non-success status is treated as cleanup, though a client handling that status loses the behaviour"
	reasonOptHeader    = "below the law, pending review: optional softens the verdict on the strength of what clients ought not rely on"
	reasonNotWriteOnly = "above the law, pending review: flags that the field now appears in responses"
	reasonAllOfRemoved = "above the law, pending review: widening kept WARN as caution"
	// below the law: a milder verdict owes a contract argument
	reasonSecurity     = "below the law, pending review: security alternatives removed or scopes added break clients but are INFO"
	reasonAnyOf        = "below the law, pending review: contradicts response one-of-added which is ERR"
	reasonPattern1034  = "below the law, pending review: the #1034 response pattern gap"
	reasonPrefixItems  = "below the law, pending review: stored levels assume a containment direction prefixItems does not have"
	reasonMediaName    = "below the law, pending review: clients negotiating the old media type name break"
	reasonEnumAdded    = "below the law, pending review: the server may emit a value clients never handled"
	reasonParamDefault = "below the law, pending review: parameter defaults are ERR while body and property defaults are INFO"
)

// severityDeviations records every rule whose stored level deviates from the
// law, with the reason. An unexplained mismatch and a stale entry both fail
// TestSeverityLaw.
var severityDeviations = map[string]string{
	// A1: bound-set WARN convention
	"request-body-exclusive-max-set":      reasonBoundSet,
	"request-body-exclusive-min-set":      reasonBoundSet,
	"request-body-max-items-set":          reasonBoundSet,
	"request-body-max-length-set":         reasonBoundSet,
	"request-body-max-properties-set":     reasonBoundSet,
	"request-body-max-set":                reasonBoundSet,
	"request-body-min-items-set":          reasonBoundSet,
	"request-body-min-set":                reasonBoundSet,
	"request-body-multiple-of-set":        reasonBoundSet,
	"request-parameter-exclusive-max-set": reasonBoundSet,
	"request-parameter-exclusive-min-set": reasonBoundSet,
	"request-parameter-max-length-set":    reasonBoundSet,
	"request-parameter-max-set":           reasonBoundSet,
	"request-parameter-min-items-set":     reasonBoundSet,
	"request-parameter-min-set":           reasonBoundSet,
	"request-property-exclusive-max-set":  reasonBoundSet,
	"request-property-exclusive-min-set":  reasonBoundSet,
	"request-property-max-items-set":      reasonBoundSet,
	"request-property-max-length-set":     reasonBoundSet,
	"request-property-max-properties-set": reasonBoundSet,
	"request-property-max-set":            reasonBoundSet,
	"request-property-min-items-set":      reasonBoundSet,
	"request-property-min-set":            reasonBoundSet,
	"request-property-multiple-of-set":    reasonBoundSet,
	// A2: removal signals behavior change
	"request-body-removed":               reasonBodyRemoved,
	"request-parameter-removed":          reasonRemoval,
	"request-property-removed":           reasonRemoval,
	"response-optional-property-removed": reasonRemoval,
	// A3: wrapped in oneOf
	"request-body-wrapped-in-one-of":  reasonWrapped,
	"response-body-wrapped-in-one-of": reasonWrapped,
	// A4: singles
	"response-non-success-status-removed":              reasonNonSuccess,
	"optional-response-header-removed":                 reasonOptHeader,
	"response-required-property-became-not-write-only": reasonNotWriteOnly,
	"request-body-all-of-removed":                      reasonAllOfRemoved,
	"request-property-all-of-removed":                  reasonAllOfRemoved,
	// B1: security
	"api-security-removed":            reasonSecurity,
	"api-global-security-removed":     reasonSecurity,
	"api-security-scope-added":        reasonSecurity,
	"api-global-security-scope-added": reasonSecurity,
	// B2: anyOf vs oneOf
	"response-body-any-of-added":     reasonAnyOf,
	"response-property-any-of-added": reasonAnyOf,
	// B3: #1034
	"response-property-pattern-removed": reasonPattern1034,
	"response-property-pattern-changed": reasonPattern1034,
	// B4: prefixItems
	"request-body-prefix-items-added":        reasonPrefixItems,
	"request-body-prefix-items-removed":      reasonPrefixItems,
	"request-property-prefix-items-added":    reasonPrefixItems,
	"request-property-prefix-items-removed":  reasonPrefixItems,
	"response-body-prefix-items-added":       reasonPrefixItems,
	"response-body-prefix-items-removed":     reasonPrefixItems,
	"response-property-prefix-items-added":   reasonPrefixItems,
	"response-property-prefix-items-removed": reasonPrefixItems,
	// B5: singles
	"response-media-type-name-changed":        reasonMediaName,
	"response-property-enum-value-added":      reasonEnumAdded,
	"request-parameter-default-value-added":   reasonParamDefault,
	"request-parameter-default-value-changed": reasonParamDefault,
	"request-parameter-default-value-removed": reasonParamDefault,
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
