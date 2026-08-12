package checker_test

import (
	"fmt"
	"sort"
	"testing"

	"github.com/oasdiff/oasdiff/checker"
)

// Severity monotonicity under containment: if one change's scope contains
// another's, it cannot be reported at a lower level. Removing a whole response
// schema takes every property with it, so it cannot be milder than removing one
// of those properties.
//
// TestRuleSymmetry cannot catch this. It never reads Level, and it compares
// rules that differ in exactly one taxonomy coordinate. Containment is
// semantic: response-body-media-type-schema-removed and
// response-required-property-removed differ only in Area, which no axis it
// walks ever flips, so the two are never put side by side. #1140 was an
// instance, found by accident.
//
// Containment is curated rather than derived, for that same reason. The list
// does not need to be exhaustive: each pair is one class of inconsistency that
// can no longer regress.
var containmentPairs = [][2]string{
	// A response media type contains its schema, which contains its properties.
	{checker.ResponseMediaTypeRemovedId, checker.ResponseBodyMediaTypeSchemaRemovedId},
	{checker.ResponseBodyMediaTypeSchemaRemovedId, checker.ResponseRequiredPropertyRemovedId},

	// The request body contains its media types, which contain their schemas.
	{checker.RequestBodyRemovedId, checker.RequestBodyMediaTypeRemovedId},
	{checker.RequestBodyMediaTypeRemovedId, checker.RequestBodyMediaTypeSchemaRemovedId},

	// A path contains its operations.
	{checker.APIPathRemovedWithoutDeprecationId, checker.APIRemovedWithoutDeprecationId},

	// A success response contains the headers declared on it.
	{checker.ResponseSuccessStatusRemovedId, checker.RequiredResponseHeaderRemovedId},
}

// monotonicityWaivers records a pair whose severities are deliberately
// inverted. Key is "outer>inner". Empty today; a waiver wants a reason that
// survives review, not a note that the test was failing.
var monotonicityWaivers = map[string]string{}

func TestSeverityMonotonicity(t *testing.T) {
	levels := map[string]checker.Level{}
	for _, r := range checker.GetAllRules() {
		levels[r.Id] = r.Level
	}

	violations := map[string]bool{}
	for _, p := range containmentPairs {
		outer, inner := p[0], p[1]
		lo, ok := levels[outer]
		if !ok {
			t.Errorf("containment pair names an unknown rule: %q", outer)
			continue
		}
		li, ok := levels[inner]
		if !ok {
			t.Errorf("containment pair names an unknown rule: %q", inner)
			continue
		}
		if lo >= li {
			continue
		}
		key := fmt.Sprintf("%s>%s", outer, inner)
		violations[key] = true
		if _, waived := monotonicityWaivers[key]; !waived {
			t.Errorf("severity dips under containment: %q is %v but contains %q, which is %v\n"+
				"  a change cannot be milder than something it subsumes; raise the outer rule, or document the exception in monotonicityWaivers with a reason",
				outer, lo, inner, li)
		}
	}

	stale := make([]string, 0, len(monotonicityWaivers))
	for w := range monotonicityWaivers {
		if !violations[w] {
			stale = append(stale, w)
		}
	}
	sort.Strings(stale)
	for _, w := range stale {
		t.Errorf("stale monotonicity waiver: %q\n  the severities no longer dip here; remove the waiver", w)
	}
}
