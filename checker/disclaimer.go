package checker

import "encoding/json"

// A Disclaimer names a condition that made a comparison inexact, so a rule's
// severity claims more certainty than the inputs support. It records only what
// was imperfect; what that costs a change is policy in disclaimerPolicies.
//
// A condition need not be an option the caller did not pass: a construct
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

func (d Disclaimer) String() string {
	switch d {
	case DisclaimerAllOfNotFlattened:
		return "all-of-not-flattened"
	case DisclaimerVersionsDiffer:
		return "openapi-versions-differ"
	}
	return "unknown-disclaimer"
}

// MarshalJSON and MarshalYAML emit the name, so a consumer reads
// "all-of-not-flattened" rather than an integer whose meaning depends on
// declaration order.
func (d Disclaimer) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d Disclaimer) MarshalYAML() (any, error) {
	return d.String(), nil
}

func (d Disclaimer) commentId() string {
	return d.String() + "-disclaimer"
}

type disclaimerPolicy struct {
	// maxLevel of NONE leaves the severity alone.
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

// Runs after the checks and before the level filter, so a capped change is
// filtered and counted at the level it ends up with.
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

// A level the caller set wins: an override is a decision about that rule, and
// a disclaimer is not entitled to overrule it.
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
