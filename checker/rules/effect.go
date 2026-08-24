package rules

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
	// EffectIncomparable: the change both rejects payloads that were valid
	// and accepts payloads that were not, so neither set contains the other.
	EffectIncomparable
	// EffectUnknown: the check cannot tell which way the change goes.
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
