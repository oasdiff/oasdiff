package checker

import (
	"slices"
	"sync"

	"github.com/oasdiff/oasdiff/diff"
)

// Some schema changes are really one change that shows up as several raw
// diffs. For example, making a schema nullable by wrapping it in
// oneOf: [{type: "null"}, X] is one change, but the raw diff shows three:
// the type changed, the enum was removed, and a oneOf was added. Without
// special handling, each would be reported as a separate finding, some with
// the wrong verdict (enum values reported as removed when they were not).
//
// We call such a change a schema transition. The transitions table below
// lists the recognized transitions. Each entry defines:
//
//   - present: how to tell that the transition happened at a schema node.
//   - claims: which kinds of raw findings the transition explains; these
//     are dropped. Findings of other kinds are real changes and are still
//     reported: a single edit can widen a type into a oneOf and also drop
//     an enum value, and only the first is the transition.
//   - reportedBy: the checks that report the transition itself. A
//     transition never drops its own reporters, so its finding always
//     survives.
//
// The decision is made when a change is created: ApiChange.WithSchema
// consults this table (see claimedByTransition), so checkers contain no
// suppression logic.
type transition struct {
	// present reports whether the transition occurred at this schema node.
	present func(*diff.SchemaDiff) bool
	// claims is the transition's suppression scope: the kinds of raw
	// field changes that are mechanical echoes of the transition at its node
	// and are dropped. Kinds outside the set are independent changes that
	// keep reporting even when the transition is present: the nullable
	// wrapping claims values because its branch is proven equivalent to the
	// base (an enum diff there can only be an echo of the wrap), while
	// list-of-types does not, so an enum change alongside a widened type is
	// still reported.
	//
	// The scope names what is suppressed, not what survives, and by kind
	// rather than by rule: echoes are structural (a transition perturbs
	// specific field families, and a rule's Kind says which family it
	// reads), so a few kinds cover every current and future rule of those
	// families, where the survivor complement would be hundreds of rule ids
	// kept by hand. It also fails in the safe direction: a rule this table
	// doesn't account for keeps reporting, while a stale survivor list would
	// silently swallow findings.
	claims map[Kind]bool
	// reportedBy names the rules that report the transition itself: in
	// effect, the findings that supersede the claimed raw findings. Note the
	// mechanism, though: suppression is driven entirely by claims;
	// reportedBy only exempts the listed rules from their own transition's
	// claims, so the superseding finding cannot suppress itself.
	// TestTransitionReportersRegistered asserts each is a registered rule.
	reportedBy []string
}

var transitions = []transition{
	// Nullable wrapping: base X became oneOf: [{type: "null"}, X], the
	// OpenAPI 3.1 idiom for making a $ref'd schema nullable (see
	// diff.NullableWrappingDiff). Claims every raw reflection of the wrap:
	// the type reads as changed, the value and constraint keywords (enum,
	// pattern) read as removed since they moved into the branch, and the
	// oneOf reads as added. Reported as became-nullable at body and property
	// level; at parameter level, where no became-nullable rule exists, the
	// parameter list-of-types finding is the reporter, and listing it here is
	// what keeps it alive there via the reportedBy exemption.
	{
		present: func(d *diff.SchemaDiff) bool { return !d.NullableWrappingDiff.Empty() },
		claims:  map[Kind]bool{KindType: true, KindValues: true, KindConstraints: true, KindStructure: true},
		reportedBy: []string{
			RequestBodyBecomeNullableId, RequestPropertyBecomeNullableId,
			ResponseBodyBecameNullableId, ResponsePropertyBecameNullableId,
			RequestParameterListOfTypesWidenedId,
		},
	},
	// oneOf wrapping: a concrete object schema became a oneOf of object
	// alternatives (see diff.OneOfWrappingDiff), a breaking restructuring.
	// Claims the raw type change (the wrapper reads as type "any") and the
	// raw oneOf membership changes. Reported as body-wrapped-in-one-of at
	// body level; at property level no dedicated finding exists, so the
	// property one-of-added finding is the reporter (upgrading it to a
	// dedicated finding is a reportedBy swap here).
	{
		present: func(d *diff.SchemaDiff) bool { return !d.OneOfWrappingDiff.Empty() },
		claims:  map[Kind]bool{KindType: true, KindStructure: true},
		reportedBy: []string{
			RequestBodyWrappedInOneOfId, ResponseBodyWrappedInOneOfId,
			RequestPropertyOneOfAddedId, ResponsePropertyOneOfAddedId,
		},
	},
	// List-of-types: a single type became a oneOf/anyOf of scalar types, or
	// back (see diff.ListOfTypesDiff). Claims the raw type change and the raw
	// oneOf/anyOf membership changes. Reported as list-of-types
	// widened/narrowed at every level.
	{
		present: func(d *diff.SchemaDiff) bool { return !d.ListOfTypesDiff.Empty() },
		claims:  map[Kind]bool{KindType: true, KindStructure: true},
		reportedBy: []string{
			RequestBodyListOfTypesWidenedId, RequestBodyListOfTypesNarrowedId,
			RequestPropertyListOfTypesWidenedId, RequestPropertyListOfTypesNarrowedId,
			ResponseBodyListOfTypesWidenedId, ResponseBodyListOfTypesNarrowedId,
			ResponsePropertyListOfTypesWidenedId, ResponsePropertyListOfTypesNarrowedId,
			RequestParameterListOfTypesWidenedId, RequestParameterListOfTypesNarrowedId,
			RequestParameterPropertyListOfTypesWidenedId, RequestParameterPropertyListOfTypesNarrowedId,
		},
	},
	// Null-only type change: "null" was added to or removed from the type
	// set, with no other type or format change (the OpenAPI 3.1
	// type: [string, "null"] form). Claims the raw type change. Reported as
	// became-(not-)nullable.
	{
		present: func(d *diff.SchemaDiff) bool { return isNullTypeChange(d.TypeDiff) && d.FormatDiff.Empty() },
		claims:  map[Kind]bool{KindType: true},
		reportedBy: []string{
			RequestBodyBecomeNullableId, RequestBodyBecomeNotNullableId,
			RequestPropertyBecomeNullableId, RequestPropertyBecomeNotNullableId,
			ResponseBodyBecameNullableId, ResponsePropertyBecameNullableId,
		},
	},
}

// ruleKinds indexes each registered rule's Kind so claimedByTransition can
// derive a change's kind from its rule definition. A function rather than an
// initialized package var: the rule handlers reference the claim path, so a
// var initializer calling GetAllRules would be an initialization cycle.
var (
	ruleKindsOnce sync.Once
	ruleKindsMap  map[string]Kind
)

func ruleKinds() map[string]Kind {
	ruleKindsOnce.Do(func() {
		ruleKindsMap = make(map[string]Kind, len(GetAllRules()))
		for _, rule := range GetAllRules() {
			ruleKindsMap[rule.Id] = rule.Kind
		}
	})
	return ruleKindsMap
}

// claimedByTransition reports whether a change with the given rule id,
// computed from the given schema node, is a raw reflection of a recognized
// transition there and should be suppressed. Called by ApiChange.WithSchema.
//
// Kind matching: the caller does not state the change's kind; it is looked up
// from the rule's registry entry (rules.go), so the pairing between a
// transition and the checks it suppresses cannot drift from the rule
// definitions. A change is claimed when some transition present at the node
// claims the change's kind, unless the rule is one of that transition's own
// reporters. A nil node or an unregistered id never claims, so the failure
// mode of missing plumbing is over-reporting, never a lost finding.
func claimedByTransition(schemaDiff *diff.SchemaDiff, ruleId string) bool {
	if schemaDiff == nil {
		return false
	}
	kind, ok := ruleKinds()[ruleId]
	if !ok {
		return false
	}
	for _, t := range transitions {
		if t.claims[kind] && t.present(schemaDiff) && !slices.Contains(t.reportedBy, ruleId) {
			return true
		}
	}
	return false
}
