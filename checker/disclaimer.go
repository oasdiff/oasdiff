package checker

// A rule's severity says what a change means when oasdiff can see the whole
// picture. Sometimes it cannot, and the verdict is then less certain than the
// severity claims.
//
// A Disclaimer names one such condition. It records only what was imperfect;
// what that costs a change, a comment and a severity ceiling, is policy and
// lives in disclaimerPolicies. Keeping the two apart means a check states a
// fact about its inputs without also deciding what the fact is worth, and the
// worth can be revised in one table.
//
// Conditions need not be about options the caller did not pass. A construct
// oasdiff cannot compare exactly however it is invoked belongs here too.
type Disclaimer int

const (
	// DisclaimerAllOfNotFlattened marks a change found inside an allOf branch
	// that was compared branch by branch. A sibling branch can require what
	// this one dropped, so the change may have no effect on the merged schema.
	DisclaimerAllOfNotFlattened Disclaimer = iota + 1

	// DisclaimerVersionsDiffer marks a comparison whose two specs declare
	// different OpenAPI versions. The dialects rewrite constructs without
	// changing what they mean, so a difference can be notation alone.
	DisclaimerVersionsDiffer
)

// String returns the disclaimer's stable name, used in output and as the stem
// of its localization key.
func (d Disclaimer) String() string {
	switch d {
	case DisclaimerAllOfNotFlattened:
		return "all-of-not-flattened"
	case DisclaimerVersionsDiffer:
		return "openapi-versions-differ"
	}
	return "unknown-disclaimer"
}

func (d Disclaimer) commentId() string {
	return d.String() + "-disclaimer"
}

// disclaimerPolicy is what a disclaimer costs a change.
type disclaimerPolicy struct {
	// maxLevel caps the change's severity. NONE leaves the severity alone,
	// which is right when the condition makes some changes doubtful but not
	// the whole comparison.
	maxLevel Level
}

var disclaimerPolicies = map[Disclaimer]disclaimerPolicy{
	// Every change under an unflattened allOf is doubtful in the same way, so
	// the ceiling applies to all of them: warning states the doubt, which is
	// what the severity rule reserves warning for.
	DisclaimerAllOfNotFlattened: {maxLevel: WARN},

	// No ceiling. Differing versions make dialect-shaped changes doubtful, but
	// most changes are not dialect-shaped: a removed endpoint breaks consumers
	// whichever version either spec declares, and capping it would be worse
	// than the false positive the disclaimer warns about.
	DisclaimerVersionsDiffer: {},
}

// applyDisclaimers records the conditions that hold for this comparison and
// applies their policy. Runs after the checks and before the level filter, so
// a capped change is filtered and counted at the level it ends up with.
func applyDisclaimers(config *Config, changes Changes, versionsDiffer bool) Changes {
	for i, change := range changes {
		apiChange, ok := change.(ApiChange)
		if !ok {
			continue
		}
		if versionsDiffer {
			apiChange.Disclaimers = append(apiChange.Disclaimers, DisclaimerVersionsDiffer)
		}
		changes[i] = capByDisclaimers(config, apiChange)
	}
	return changes
}

// capByDisclaimers lowers a change to the strictest ceiling its disclaimers
// impose. A level the caller set explicitly wins: an override is a decision
// about this rule, and a disclaimer is not entitled to overrule it.
func capByDisclaimers(config *Config, change ApiChange) ApiChange {
	if config.overriddenLevels[change.Id] {
		return change
	}
	for _, d := range change.Disclaimers {
		if max := disclaimerPolicies[d].maxLevel; max != NONE && change.Level > max {
			change.Level = max
		}
	}
	return change
}
