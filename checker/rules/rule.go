package rules

import (
	"slices"

	"github.com/oasdiff/oasdiff/checker/metaschema"
)

// Rule is the metadata of one backward-compatibility rule: where it fires
// (Locations), what it concludes (Effect), under which document states
// (Guards), which side of the wire it judges (Direction), and its severity
// (Level). The checker package binds a Rule to its implementation.
type Rule struct {
	Id          string
	Level       Level
	Description string
	Direction   Direction
	Area        Area
	Kind        Kind
	Effect      Effect
	Guards      []Guard
	// Locations are the edits the rule covers, as
	// "pattern:action[,action...]" claims (see metaschema.ParseClaim).
	Locations []string
}

// Actions returns the syntactic edits the rule covers, derived from its
// location claims: the rule fires on these edit verbs at its locations.
// Unlike Effect, which is the rule's semantic verdict, actions describe what
// literally changed in the document.
func (r Rule) Actions() []metaschema.Action {
	var actions []metaschema.Action
	for _, loc := range r.Locations {
		claim, err := metaschema.ParseClaim(loc)
		if err != nil {
			continue // an invalid claim fails the checker's audit; not reported here
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

// DerivedLevel applies the severity law to the rule's own metadata.
func (r Rule) DerivedLevel() Level {
	return DeriveLevel(r.Effect, r.Direction, r.Guards...)
}
