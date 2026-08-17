package metaschema

import (
	"fmt"
	"slices"
	"strings"
)

// Claim is a rule's statement of which edits it covers: a location pattern
// (see MatchLocation) plus the actions it reports there.
type Claim struct {
	Pattern string
	Actions []Action
}

var validActions = map[Action]bool{
	ActionAdd:      true,
	ActionRemove:   true,
	ActionSet:      true,
	ActionUnset:    true,
	ActionChange:   true,
	ActionIncrease: true,
	ActionDecrease: true,
}

// ParseClaim parses "location:action[,action...]", e.g.
// "paths.*.*.requestBody.content.*.schema.maxLength:decrease,unset".
func ParseClaim(s string) (Claim, error) {
	pattern, actions, ok := strings.Cut(s, ":")
	if !ok || pattern == "" || actions == "" {
		return Claim{}, fmt.Errorf("claim %q: want location:action[,action...]", s)
	}
	c := Claim{Pattern: pattern}
	for a := range strings.SplitSeq(actions, ",") {
		if !validActions[Action(a)] {
			return Claim{}, fmt.Errorf("claim %q: unknown action %q", s, a)
		}
		c.Actions = append(c.Actions, Action(a))
	}
	return c, nil
}

// Matches reports whether the claim covers the edit.
func (c Claim) Matches(edit Edit) bool {
	return MatchLocation(c.Pattern, edit.Location) && slices.Contains(c.Actions, edit.Action)
}
