package checker

// Effect is a rule's verdict about the accepted-value set: how the set of
// payloads (or API surface) the contract admits changes when the rule fires.
// It is the semantic axis of a rule, orthogonal to the syntactic edit
// recorded in the rule's location claims, and together with Direction it
// determines the expected severity (see rule_severity_law_test.go).
type Effect int8

const (
	// EffectNone: the change has no accepted-set semantics (metadata,
	// lifecycle annotations, defaults).
	EffectNone Effect = iota
	// EffectWidens: the accepted set provably grows.
	EffectWidens
	// EffectNarrows: the accepted set provably shrinks.
	EffectNarrows
	// EffectIncomparable: provably neither set contains the other.
	EffectIncomparable
	// EffectUnknown: the check cannot decide the containment.
	EffectUnknown
	// EffectViolation: the change breaks oasdiff's lifecycle governance
	// (removal before sunset, invalid or missing sunset, stability decrease)
	// rather than the wire contract.
	EffectViolation
)

func (e Effect) String() string {
	switch e {
	case EffectWidens:
		return "widens"
	case EffectNarrows:
		return "narrows"
	case EffectIncomparable:
		return "incomparable"
	case EffectUnknown:
		return "unknown"
	case EffectViolation:
		return "violation"
	default:
		return "none"
	}
}

// Guard is a named predicate over the document state that a rule requires
// before it fires. Guards partition a cell's edits into sub-cases: rules at
// the same location claims may carry different Effects and severities
// because their guards select different document states.
type Guard string

const (
	// GuardReadOnly: the changed property is readOnly, so it does not
	// appear in requests; request-side effects are nullified.
	GuardReadOnly Guard = "read-only"
	// GuardWriteOnly: the changed property is writeOnly, so it does not
	// appear in responses; response-side effects are nullified.
	GuardWriteOnly Guard = "write-only"
	// GuardSanctioned: the removed element was deprecated and its sunset
	// period was honored, so the removal follows the deprecation contract.
	GuardSanctioned Guard = "sanctioned"
	// GuardNonSuccess: the affected response status is a non-success
	// status.
	GuardNonSuccess Guard = "non-success"
	// GuardHasDefault: the changed element declares a default value.
	GuardHasDefault Guard = "has-default"
	// GuardNegotiated: the rule judges the availability of a client-selected
	// variant (a response status, media type, or header). The client chooses
	// or relies on the variant, so contravariance applies with request
	// polarity even though the variant lives on the response side.
	GuardNegotiated Guard = "negotiated"
)
