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
	effect, direction = applyGuards(effect, direction, guards)
	return levelOf(effect, direction)
}

// applyGuards requalifies the effect and direction by the document states
// the guards name: a guard either nullifies the effect on the side it
// speaks about or changes which side the change is judged on.
func applyGuards(effect Effect, direction Direction, guards []Guard) (Effect, Direction) {
	for _, g := range guards {
		switch g {
		case GuardReadOnly:
			// a readOnly property does not appear in requests
			if direction == DirectionRequest {
				effect = EffectNone
			}
		case GuardWriteOnly:
			// a writeOnly property does not appear in responses
			if direction == DirectionResponse {
				effect = EffectNone
			}
		case GuardNonSuccess:
			// the responses map does not promise that the server returns
			// only the statuses it lists
			effect = EffectNone
		case GuardSanctioned:
			// the deprecation contract was honored
			effect = EffectNone
		case GuardNegotiated:
			// the client chooses or relies on the variant
			direction = DirectionRequest
		}
	}
	return effect, direction
}

// levelOf maps effect and direction to a level: narrowing breaks request
// consumers, widening breaks response consumers, an incomparable change
// breaks both, and an unknown one is a warning.
func levelOf(effect Effect, direction Direction) Level {
	switch effect {
	case EffectViolation, EffectIncomparable:
		return ERR
	case EffectNone:
		return INFO
	case EffectUnknown:
		return WARN
	case EffectNarrows:
		if direction == DirectionResponse {
			return INFO
		}
		return ERR
	case EffectWidens:
		if direction == DirectionResponse {
			return ERR
		}
		return INFO
	}
	return INFO
}
