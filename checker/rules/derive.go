package rules

// DeriveLevel is the severity law: a rule's level follows from what the
// change does to the contract (effect), which side of the wire it judges
// (direction), and the document states that qualify the verdict (guards).
//
// The law is asymmetric on purpose. Reporting a breaking change as safe
// ships it silently, so a milder verdict than the effect implies owes a
// proof that the change is safe for every consumer that conformed to the
// old contract, while a harsher one costs a reviewer a glance and needs
// only a reason. "It is sometimes legitimately required" is a reason to
// accept a breaking change, not a reason it is not one, and belongs in
// --severity-levels rather than in the default verdict.
func DeriveLevel(effect Effect, direction Direction, guards ...Guard) Level {
	level, _ := ExplainLevel(effect, direction, guards...)
	return level
}

// ExplainLevel is DeriveLevel with the derivation rendered in words: one
// sentence per guard that requalifies the change, and one for the mapping of
// the resulting effect and direction to a level. The words and the level
// come from the same derivation, so they cannot drift apart.
func ExplainLevel(effect Effect, direction Direction, guards ...Guard) (Level, []string) {
	var steps []string
	for _, g := range guards {
		var sentence string
		effect, direction, sentence = applyGuard(g, effect, direction)
		if sentence != "" {
			steps = append(steps, sentence)
		}
	}
	level, sentence := levelOf(effect, direction)
	return level, append(steps, sentence)
}

// applyGuard requalifies the effect and direction by the document state the
// guard names: a guard either nullifies the effect on the side it speaks
// about or changes which side the change is judged on. The sentence says
// what the guard did, and is empty for a guard that changed nothing.
func applyGuard(g Guard, effect Effect, direction Direction) (Effect, Direction, string) {
	switch g {
	case GuardReadOnly:
		if direction == DirectionRequest {
			return EffectNone, direction, "The property is read-only, so it never appears in requests and the change cannot affect them."
		}
	case GuardWriteOnly:
		if direction == DirectionResponse {
			return EffectNone, direction, "The property is write-only, so it never appears in responses and the change cannot affect them."
		}
	case GuardNonSuccess:
		return EffectNone, direction, "The status is not a success status, and the responses map does not promise that the server returns only the statuses it lists."
	case GuardSanctioned:
		return EffectNone, direction, "The element was deprecated and its sunset date honored, so the change keeps the deprecation contract."
	case GuardNegotiated:
		return effect, DirectionRequest, "The client selects or relies on this element, so the change is judged as if it were on the request side."
	}
	return effect, direction, ""
}

// levelOf maps effect and direction to a level, with the sentence that
// justifies it.
func levelOf(effect Effect, direction Direction) (Level, string) {
	switch effect {
	case EffectViolation:
		return ERR, "The change violates the deprecation or stability contract, which is an error regardless of direction."
	case EffectIncomparable:
		return ERR, "The change both rejects payloads that were valid and accepts payloads that were not, so it breaks consumers on both sides: error."
	case EffectNone:
		return INFO, "The change does not affect which payloads the contract accepts: info."
	case EffectUnknown:
		return WARN, "The spec does not carry enough information to decide the effect, and the change is plausibly breaking: warning."
	case EffectNarrows:
		if direction == DirectionResponse {
			return INFO, "The change narrows what the contract accepts; a narrower response promise still satisfies every existing client: info."
		}
		return ERR, "The change narrows what the contract accepts; requests that were valid become invalid: error."
	case EffectWidens:
		if direction == DirectionResponse {
			return ERR, "The change widens what the contract accepts; clients may receive responses they were never written to handle: error."
		}
		return INFO, "The change widens what the contract accepts; every request that was valid remains valid: info."
	}
	return INFO, ""
}
