package checker_test

import (
	"fmt"
	"regexp"
	"sort"
	"testing"

	"github.com/oasdiff/oasdiff/checker"
)

// Severity-law prototype: assign each rule an Effect (how the accepted-value
// set changes) and Guards from its semantics — the id's keyword and edit verb,
// never its level — then derive the expected level from the contravariance
// law and report agreement with the stored level.
//
// This is the empirical basis for making severity a derived property: the law
// plus a handful of policy rows explains the registry, and every mismatch is
// either a documented convention or a severity bug. Run with:
//
//	go test ./checker -run SeverityLawReport -v

type lawEffect string

const (
	effWidens       lawEffect = "widens"       // accepted set provably grows
	effNarrows      lawEffect = "narrows"      // accepted set provably shrinks
	effIncomparable lawEffect = "incomparable" // provably neither contains the other
	effUnknown      lawEffect = "unknown"      // the check cannot decide
	effNone         lawEffect = "none"         // no accepted-set semantics (metadata)
	effViolation    lawEffect = "violation"    // breaks oasdiff's lifecycle governance
)

type lawEntry struct {
	pattern *regexp.Regexp
	effect  lawEffect
}

func law(pattern string, effect lawEffect) lawEntry {
	return lawEntry{regexp.MustCompile(pattern), effect}
}

// effectTable assigns Effect by first match on the rule id.
var effectTable = []lawEntry{
	// oasdiff's lifecycle governance: violations of the deprecation contract
	law(`sunset-(parse|missing|deleted)|sunset-date.*too-small|sunset-invalid|before-sunset`, effViolation),
	law(`stability-decreased|invalid-stability`, effViolation),

	// lifecycle / metadata: no accepted-set semantics
	law(`deprecated|reactivated|sunset`, effNone),
	law(`stability`, effNone),
	law(`operation-id`, effNone),
	law(`^api-tag-|^api-global-tag`, effNone),
	law(`api-.*version`, effNone),
	law(`discriminator`, effNone),
	law(`example`, effNone),
	law(`annotation-only`, effNone),
	law(`^api-schema-removed`, effNone),

	// element existence and defaults, refined
	law(`request-body-added-required`, effNarrows),
	law(`request-body-added-optional`, effNone),
	law(`new-required-request-default-parameter`, effNarrows),
	law(`new-optional-request-default-parameter`, effNone),
	law(`new-request-path-parameter`, effNarrows),
	law(`default-value`, effNone),
	// read-only / write-only are SHOULD-level advisory metadata in the spec
	law(`became-(not-)?(read|write)-only`, effNone),
	law(`required-write-only-property-added`, effNarrows),
	law(`optional-write-only-property-added`, effNone),

	// security: requirements are OR-alternatives; scopes are AND-ed within one
	law(`security-component.*(added|removed|changed|updated)`, effNone),
	law(`security-scope-added`, effNarrows),
	law(`security-scope-removed`, effWidens),
	law(`security-added`, effWidens),
	law(`security-removed|security-deleted`, effNarrows),
	law(`security.*(updated|changed)`, effIncomparable),

	// existence of API surface
	law(`(api|endpoint|path).*removed`, effNarrows),
	law(`(api|endpoint|path).*added`, effWidens),
	law(`webhook.*removed`, effNarrows),
	law(`webhook.*added`, effWidens),

	// a schema appearing on a media type constrains what was unconstrained
	law(`media-type-(item-)?schema-added`, effNarrows),
	law(`media-type-(item-)?schema-removed(-untyped)?`, effWidens),
	// media type names
	law(`media-type-name-generalized`, effWidens),
	law(`media-type-name-specialized`, effNarrows),
	law(`media-type-name-changed`, effIncomparable),
	// negotiated variants: the client chooses, so removal breaks clients
	law(`media-type.*removed|media-type.*deleted`, effNarrows),
	law(`media-type.*added`, effWidens),
	law(`enum-value-removed`, effNarrows),
	law(`enum-value-added`, effWidens),

	// requiredness
	law(`became-required|required-request-body-added`, effNarrows),
	law(`became-optional`, effWidens),
	law(`new-required-request-(property|parameter|header)`, effNarrows),
	law(`new-optional-request`, effNone),
	law(`required-property-added`, effNarrows),
	law(`optional-property-added`, effNone),
	law(`required-property-removed`, effWidens),
	law(`optional-property-removed`, effNone),

	// nullability
	law(`became-nullable|became-null`, effWidens),
	law(`became-not-nullable|became-non-null`, effNarrows),

	// enums / const
	law(`became-enum`, effNarrows),
	law(`const-added`, effNarrows),
	law(`const-removed`, effWidens),
	law(`const-changed`, effIncomparable),

	// bounds
	law(`(max-length|max-items|max-properties|maximum|max|exclusive-max[a-z-]*)-decreased`, effNarrows),
	law(`(max-length|max-items|max-properties|maximum|max|exclusive-max[a-z-]*)-increased`, effWidens),
	law(`(max-length|max-items|max-properties|maximum|max|exclusive-max[a-z-]*)-set`, effNarrows),
	law(`(max-length|max-items|max-properties|maximum|max|exclusive-max[a-z-]*)-unset`, effWidens),
	law(`(min-length|min-items|min-properties|minimum|min|exclusive-min[a-z-]*)-increased`, effNarrows),
	law(`(min-length|min-items|min-properties|minimum|min|exclusive-min[a-z-]*)-decreased`, effWidens),
	law(`(min-length|min-items|min-properties|minimum|min|exclusive-min[a-z-]*)-set`, effNarrows),
	law(`(min-length|min-items|min-properties|minimum|min|exclusive-min[a-z-]*)-unset`, effWidens),
	law(`min-contains-increased`, effNarrows),
	law(`min-contains-decreased`, effWidens),
	law(`max-contains-increased`, effWidens),
	law(`max-contains-decreased`, effNarrows),

	// pattern
	law(`pattern-added`, effNarrows),
	law(`pattern-removed`, effWidens),
	law(`pattern-generalized`, effWidens),
	law(`pattern-changed`, effUnknown),

	// uniqueItems / multipleOf
	law(`unique-items-set`, effNarrows),
	law(`unique-items-unset`, effWidens),
	law(`multiple-of-set`, effNarrows),
	law(`multiple-of-unset`, effWidens),
	law(`multiple-of-generalized`, effWidens),
	law(`multiple-of-specialized`, effNarrows),
	law(`multiple-of-changed`, effIncomparable),

	// types: "compatible" means changed in the safe direction for the rule's side
	law(`request-parameter-property-type-changed`, effUnknown),
	law(`^request-.*type-compatible`, effWidens),
	law(`^response-.*type-compatible`, effNarrows),
	law(`type-generalized|list-of-types-widened`, effWidens),
	law(`type-specialized|list-of-types-narrowed`, effNarrows),
	law(`type-changed`, effIncomparable),

	// composition: adding an allOf conjunct narrows; anyOf/oneOf alternatives widen
	law(`all-of-added`, effNarrows),
	law(`all-of-removed`, effWidens),
	law(`(any-of|one-of)-added`, effWidens),
	law(`(any-of|one-of)-removed`, effNarrows),
	law(`wrapped-in-one-of`, effUnknown),
	// a dormant then/else is activated by if; adding conditionals narrows
	law(`(if|then|else)-added`, effNarrows),
	law(`(if|then|else)-removed`, effWidens),
	law(`contains-added`, effNarrows),
	law(`contains-removed`, effWidens),
	law(`property-names-added`, effNarrows),
	law(`property-names-removed`, effWidens),
	law(`pattern-property-added`, effNarrows),
	law(`pattern-property-removed`, effWidens),
	law(`dependent-required-added|dependent-schema-added`, effNarrows),
	law(`dependent-required-removed|dependent-schema-removed`, effWidens),
	law(`dependent-required-changed`, effIncomparable),
	law(`unevaluated-(items|properties)-added`, effNarrows),
	law(`unevaluated-(items|properties)-removed`, effWidens),
	law(`content-schema-added`, effNarrows),
	law(`content-schema-removed`, effWidens),
	law(`content-(media-type|encoding)-changed`, effIncomparable),
	// prefixItems reshapes positional constraints; neither side is contained
	law(`prefix-items-(added|removed)`, effUnknown),

	// element existence
	law(`request-body-removed`, effWidens),
	law(`parameter-removed`, effWidens),
	law(`property-removed`, effNone),
	law(`header-removed`, effNarrows),
	law(`header-added`, effWidens),
	law(`response-status-removed|success-status-removed|status-removed`, effNarrows),
	law(`success-status-added|status-added`, effWidens),
	law(`schema-added`, effNarrows),
	law(`schema-removed`, effWidens),
}

// variantCells: the client chooses or relies on the variant's existence
// (statuses, media types, response headers), so contravariance applies with
// request polarity regardless of the syntactic side.
var variantCells = regexp.MustCompile(
	`(response-media-type-(added|removed)$)|(status-(added|removed))|(response-header-(added|removed))`)

var guardTable = map[string]*regexp.Regexp{
	"readOnly":   regexp.MustCompile(`read-only`),
	"writeOnly":  regexp.MustCompile(`write-only`),
	"sanctioned": regexp.MustCompile(`with-deprecation`),
	"nonSuccess": regexp.MustCompile(`non-success`),
	"hasDefault": regexp.MustCompile(`with-default`),
}

func classifyRule(id string) (lawEffect, map[string]bool, bool) {
	var effect lawEffect
	for _, e := range effectTable {
		if e.pattern.MatchString(id) {
			effect = e.effect
			break
		}
	}
	guards := map[string]bool{}
	for name, pat := range guardTable {
		if pat.MatchString(id) {
			guards[name] = true
		}
	}
	return effect, guards, variantCells.MatchString(id)
}

func expectedLevel(effect lawEffect, guards map[string]bool, direction checker.Direction, variant bool) checker.Level {
	if guards["readOnly"] && direction == checker.DirectionRequest {
		effect = effNone
	}
	if guards["writeOnly"] && direction == checker.DirectionResponse {
		effect = effNone
	}
	if guards["sanctioned"] {
		return checker.INFO
	}
	if variant {
		direction = checker.DirectionRequest
	}
	switch effect {
	case effViolation, effIncomparable:
		return checker.ERR
	case effNone:
		return checker.INFO
	case effUnknown:
		return checker.WARN
	case effNarrows:
		if direction == checker.DirectionResponse {
			return checker.INFO
		}
		return checker.ERR
	case effWidens:
		if direction == checker.DirectionResponse {
			return checker.ERR
		}
		return checker.INFO
	}
	return checker.INFO
}

func TestSeverityLawReport(t *testing.T) {
	var matches int
	var unclassified []string
	type mismatch struct {
		id              string
		stored, derived checker.Level
		effect          lawEffect
	}
	var mismatches []mismatch

	for _, rule := range checker.GetAllRules() {
		effect, guards, variant := classifyRule(rule.Id)
		if effect == "" {
			unclassified = append(unclassified, rule.Id)
			continue
		}
		derived := expectedLevel(effect, guards, rule.Direction, variant)
		if derived == rule.Level {
			matches++
		} else {
			mismatches = append(mismatches, mismatch{rule.Id, rule.Level, derived, effect})
		}
	}

	total := len(checker.GetAllRules())
	t.Logf("rules: %d  matches: %d (%.1f%%)  mismatches: %d  unclassified: %d",
		total, matches, 100*float64(matches)/float64(total), len(mismatches), len(unclassified))
	for _, id := range unclassified {
		t.Logf("unclassified: %s", id)
	}
	sort.Slice(mismatches, func(i, j int) bool {
		if mismatches[i].effect != mismatches[j].effect {
			return mismatches[i].effect < mismatches[j].effect
		}
		return mismatches[i].id < mismatches[j].id
	})
	for _, m := range mismatches {
		t.Logf("%s", fmt.Sprintf("mismatch: %-58s stored=%-8s derived=%-8s effect=%s",
			m.id, m.stored.String(), m.derived.String(), m.effect))
	}
}
