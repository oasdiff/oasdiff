package checker

// A Disclaimer says why oasdiff could not compare something exactly. When one
// applies, the rule's severity may be too strong for the change it describes.
//
// A Disclaimer only names the problem. The ceiling it puts on a severity is in
// disclaimerCeilings; its explanation is a localized message.
type Disclaimer int

const (
	// DisclaimerAllOfNotFlattened marks a change found inside an allOf branch
	// that was compared branch by branch. A sibling branch can require what
	// this one dropped, so the change may have no effect on the merged schema.
	DisclaimerAllOfNotFlattened Disclaimer = iota + 1
)

func (d Disclaimer) String() string {
	switch d {
	case DisclaimerAllOfNotFlattened:
		return "all-of-not-flattened"
	}
	return "unknown-disclaimer"
}

// A disclaimer with no ceiling here leaves the severity alone.
var disclaimerCeilings = map[Disclaimer]Level{
	// Every change under an unflattened allOf is doubtful in the same way, so
	// the ceiling applies to all of them: warning states the doubt, which is
	// what the severity rule reserves warning for.
	DisclaimerAllOfNotFlattened: WARN,
}

func applyDisclaimerPolicies(config *Config, changes Changes) Changes {
	for i, change := range changes {
		if apiChange, ok := change.(ApiChange); ok {
			changes[i] = capByDisclaimers(config, apiChange)
		}
	}
	return changes
}

// A level the caller set is not lowered.
func capByDisclaimers(config *Config, change ApiChange) ApiChange {
	if config.overriddenLevels[change.Id] {
		return change
	}
	for _, d := range change.Disclaimers {
		if ceiling := disclaimerCeilings[d]; ceiling != NONE && change.Level > ceiling {
			change.Level = ceiling
		}
	}
	return change
}
