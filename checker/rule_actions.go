package checker

import (
	"slices"

	"github.com/oasdiff/oasdiff/checker/metaschema"
)

// Actions returns the syntactic edits the rule covers, derived from its
// location claims: the rule fires on these edit verbs at its locations.
// Unlike Effect, which is the rule's semantic verdict, actions describe what
// literally changed in the document.
func (r BackwardCompatibilityRule) Actions() []metaschema.Action {
	var actions []metaschema.Action
	for _, loc := range r.Locations {
		claim, err := metaschema.ParseClaim(loc)
		if err != nil {
			continue // TestRuleLocations reports invalid claims
		}
		for _, a := range claim.Actions {
			if !slices.Contains(actions, a) {
				actions = append(actions, a)
			}
		}
	}
	slices.Sort(actions)
	return actions
}
